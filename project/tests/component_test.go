package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"tickets/clients"
	"tickets/entities"
	"tickets/message"
	"tickets/service"
	"time"

	"github.com/avast/retry-go"
	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/lithammer/shortuuid/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {

	err := godotenv.Load("../.env")
	if err != nil {
		panic(err)
	}

	db, err := sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	redisClient := message.NewRedisClient(os.Getenv("REDIS_ADDR"))
	defer redisClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	receiptService := &clients.ReceiptsServiceMock{}
	spreadsheetsService := &clients.SpreadsheetsServiceMock{}
	filesService := &clients.FilesServiceMock{}
	deadNationService := &clients.DeadNationClientMock{}

	go func() {
		svc := service.New(db, redisClient, receiptService, spreadsheetsService, filesService, deadNationService)

		assert.NoError(t, svc.Run(ctx))
	}()

	t.Run("Post ticket status", func(t *testing.T) {
		waitForHttpServer(t)

		ticket0ConfirmedStatus := TicketStatus{
			TicketID: uuid.NewString(),
			Status:   "confirmed",
			Price: Money{
				Amount:   "1.00",
				Currency: "USD",
			},
			Email:     "happy.customer@email.com",
			BookingID: uuid.NewString(),
		}

		ticket1ConfirmedStatus := TicketStatus{
			TicketID: uuid.NewString(),
			Status:   "confirmed",
			Price: Money{
				Amount:   "3.00",
				Currency: "USD",
			},
			Email:     "disappointed.customer@email.com",
			BookingID: uuid.NewString(),
		}

		ticket1CanceledStatus := TicketStatus{
			TicketID:  ticket1ConfirmedStatus.TicketID,
			Status:    "canceled",
			Price:     ticket1ConfirmedStatus.Price,
			Email:     ticket1ConfirmedStatus.Email,
			BookingID: ticket1ConfirmedStatus.BookingID,
		}

		req := TicketsStatusRequest{
			Tickets: []TicketStatus{ticket0ConfirmedStatus, ticket1ConfirmedStatus},
		}

		sendTicketsStatus(t, req)

		assertReceiptForTicketIssued(t, receiptService, ticket0ConfirmedStatus)

		assertTicketAddedToSpreadsheet(t, ticket0ConfirmedStatus, "tickets-to-print", spreadsheetsService)

		assertTicketSavedInDb(t, db, ticket0ConfirmedStatus)

		assertTicketFileSaved(t, filesService, ticket0ConfirmedStatus)

		req = TicketsStatusRequest{
			Tickets: []TicketStatus{ticket1CanceledStatus, ticket0ConfirmedStatus},
		}

		sendTicketsStatus(t, req)

		assertTicketAddedToSpreadsheet(t, ticket1CanceledStatus, "tickets-to-refund", spreadsheetsService)

		assertTicketDeletedFromDb(t, db, ticket1CanceledStatus)
	})

	t.Run("Book Tickets", func(t *testing.T) {
		waitForHttpServer(t)

		showCapacity := 5
		showReq := PostShowRequest{
			DeadNationID:   uuid.NewString(),
			NumberOfTicket: showCapacity,
			StartTime:      time.Now().Add(time.Hour * 72),
			Title:          gofakeit.Sentence(5),
			Venue:          gofakeit.Address().Address,
		}
		showID := createShow(t, showReq)
		assert.NotEmpty(t, showID)

		bookTickets(t, ctx, PostBookTicketsRequest{
			ShowID:         showID,
			NumberOfTicket: showCapacity,
			CustomerEmail:  gofakeit.Email(),
		})

		assertTicketsBookedInDeadNation(t, deadNationService, showReq.DeadNationID, showCapacity)
	})

	t.Run("Book Tickets - overbooking", func(t *testing.T) {
		waitForHttpServer(t)
		// Create a show
		showCapacity := 5
		showReq := PostShowRequest{
			DeadNationID:   uuid.NewString(),
			NumberOfTicket: showCapacity,
			StartTime:      time.Now().Add(time.Hour * 72),
			Title:          gofakeit.Sentence(5),
			Venue:          gofakeit.Address().Address,
		}
		showID := createShow(t, showReq)
		assert.NotEmpty(t, showID)
		// Spawn a bunch of requests to book tickets concurrently
		bookTicketsConcurrently(t, showID, showCapacity, showCapacity+20)

		// Verify that the show doesn't get overbooked
		assertShowIsFullyBooked(t, db, showID, showCapacity)
	})

	t.Run("Refund Tickets", func(t *testing.T) {
		waitForHttpServer(t)

		postShowReq := PostShowRequest{
			DeadNationID:   uuid.NewString(),
			NumberOfTicket: gofakeit.Number(200, 300),
			StartTime:      time.Now().Add(time.Hour * 72),
			Title:          gofakeit.Sentence(5),
			Venue:          gofakeit.Address().Address,
		}

		showID := createShow(t, postShowReq)

		postBookTicketsReq := PostBookTicketsRequest{
			ShowID:         showID,
			NumberOfTicket: gofakeit.Number(3, 5),
			CustomerEmail:  gofakeit.Email(),
		}
		bookTickets(t, ctx, postBookTicketsReq)

		var ticketIDs []string
		var ticketsStatuses []TicketStatus
		for i := 0; i < postBookTicketsReq.NumberOfTicket; i++ {
			ticketID := uuid.NewString()
			ticketIDs = append(ticketIDs, ticketID)
			ticketStatus := TicketStatus{
				TicketID:  ticketID,
				Status:    "confirmed",
				Price:     Money{Amount: "1.00", Currency: "USD"},
				Email:     gofakeit.Email(),
				BookingID: uuid.NewString(),
			}
			ticketsStatuses = append(ticketsStatuses, ticketStatus)
		}

		ticketsStatusReq := TicketsStatusRequest{
			Tickets: ticketsStatuses,
		}
		sendTicketsStatus(t, ticketsStatusReq)

		requestTicketRefund(t, ctx, ticketIDs...)

		assertTicketRefunded(t, spreadsheetsService, ticketIDs...)

	})
}

type TicketsStatusRequest struct {
	Tickets []TicketStatus `json:"tickets"`
}

type TicketStatus struct {
	TicketID  string `json:"ticket_id"`
	Status    string `json:"status"`
	Price     Money  `json:"price"`
	Email     string `json:"customer_email"`
	BookingID string `json:"booking_id"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type PostShowRequest struct {
	DeadNationID   string    `json:"dead_nation_id"`
	NumberOfTicket int       `json:"number_of_tickets"`
	StartTime      time.Time `json:"start_time"`
	Title          string    `json:"title"`
	Venue          string    `json:"venue"`
}

type PostShowResponse struct {
	ShowID string `json:"show_id"`
}

type PostBookTicketsRequest struct {
	ShowID         string `json:"show_id"`
	NumberOfTicket int    `json:"number_of_tickets"`
	CustomerEmail  string `json:"customer_email"`
}

type PostBookTicketsResponse struct {
	BookingID string `json:"booking_id"`
}

func waitForHttpServer(t *testing.T) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			resp, err := http.Get("http://localhost:8080/health")
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
				return
			}
		},
		time.Second*10,
		time.Millisecond*50,
	)
}

func sendTicketsStatus(t *testing.T, req TicketsStatusRequest) {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	correlationID := shortuuid.New()

	ticketIDs := make([]string, 0, len(req.Tickets))
	for _, ticket := range req.Tickets {
		ticketIDs = append(ticketIDs, ticket.TicketID)
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Correlation-ID", correlationID)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", uuid.NewString())

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func createShow(t *testing.T, req PostShowRequest) string {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	correlationID := shortuuid.New()

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/shows",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Correlation-ID", correlationID)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", uuid.NewString())

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	showResp := PostShowResponse{}
	err = json.NewDecoder(resp.Body).Decode(&showResp)
	require.NoError(t, err)
	return showResp.ShowID
}

var ErrUnexpectedStatusCode = errors.New("unexpected status code")

func bookTickets(t *testing.T, ctx context.Context, req PostBookTicketsRequest) (bookingID string, isSoldOut bool) {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)

	idempotencyKey := uuid.NewString()
	correlationID := shortuuid.New()

	err = retry.Do(
		func() error {
			httpReq, err := http.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"http://localhost:8080/book-tickets",
				bytes.NewBuffer(payload),
			)
			require.NoError(t, err)

			httpReq.Header.Set("Correlation-ID", correlationID)
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("Idempotency-Key", idempotencyKey)
			resp, err := http.DefaultClient.Do(httpReq)
			if err != nil {
				return fmt.Errorf("http request: %w", err)
			}
			if resp.StatusCode == http.StatusBadRequest {
				isSoldOut = true
				return nil
			}
			if resp.StatusCode != http.StatusCreated {
				return ErrUnexpectedStatusCode
			}
			bookingResp := PostBookTicketsResponse{}
			err = json.NewDecoder(resp.Body).Decode(&bookingResp)
			if err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}
			bookingID = bookingResp.BookingID
			return nil
		},
		retry.Attempts(100),
		retry.MaxJitter(500*time.Millisecond),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			return retry.RandomDelay(n, err, config)
		}),
		retry.RetryIf(func(err error) bool {
			return errors.Is(err, ErrUnexpectedStatusCode)
		}),
	)
	require.NoError(t, err, "book tickets exhausted all retries")

	return
}

func bookTicketsConcurrently(t *testing.T, showID string, showCapacity int, workersCount int) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(workersCount)

	bookingResults := make(chan bool, workersCount)

	unlock := make(chan struct{})

	for i := 0; i < workersCount; i++ {
		go func() {
			defer wg.Done()
			req := PostBookTicketsRequest{
				ShowID:         showID,
				NumberOfTicket: 1,
				CustomerEmail:  gofakeit.Email(),
			}
			<-unlock

			_, isSoldOut := bookTickets(t, ctx, req)
			bookingResults <- isSoldOut

		}()
	}

	close(unlock)

	wg.Wait()

	close(bookingResults)

	failedWorkers := 0
	succeededWorkers := 0

	for wasSoldOut := range bookingResults {
		if wasSoldOut {
			failedWorkers++
		} else {
			succeededWorkers++
		}
	}

	require.Equal(t, showCapacity, succeededWorkers)
	require.Equal(t, workersCount-showCapacity, failedWorkers)
}

func requestTicketRefund(t *testing.T, ctx context.Context, ticketIDs ...string) {
	t.Helper()

	for _, ticketID := range ticketIDs {
		correlationID := shortuuid.New()

		httpReq, err := http.NewRequestWithContext(
			ctx,
			http.MethodPut,
			fmt.Sprintf("http://localhost:8080/ticket-refund/%s", ticketID),
			nil,
		)
		require.NoError(t, err)

		httpReq.Header.Set("Correlation-ID", correlationID)
		httpReq.Header.Set("Idempotency-Key", uuid.NewString())

		resp, err := http.DefaultClient.Do(httpReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	}
}

func assertShowIsFullyBooked(t *testing.T, db *sqlx.DB, showID string, showCapacity int) {
	t.Helper()

	var numberOfBookedTickets = 0
	err := db.Get(
		&numberOfBookedTickets,
		`SELECT
			COALESCE(SUM(number_of_tickets), 0) as number_of_booked_tickets
		FROM
			bookings
		WHERE
			show_id = $1`,
		showID,
	)
	require.NoError(t, err)
	assert.Equal(t, numberOfBookedTickets, showCapacity, "show is not booked to capacity")
}

func assertReceiptForTicketIssued(t *testing.T, receiptsService *clients.ReceiptsServiceMock, ticket TicketStatus) {
	assert.EventuallyWithT(
		t,
		func(collectT *assert.CollectT) {
			issuedReceipts := len(receiptsService.IssuedReceipts)
			t.Log("issued receipts", issuedReceipts)

			assert.Greater(collectT, issuedReceipts, 0, "no receipts issued")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var receipt entities.IssueReceiptRequest
	var ok bool
	for _, issuedReceipt := range receiptsService.IssuedReceipts {
		if issuedReceipt.TicketID != ticket.TicketID {
			continue
		}
		receipt = issuedReceipt
		ok = true
		break
	}
	require.Truef(t, ok, "receipt for ticket %s not found", ticket.TicketID)

	assert.Equal(t, ticket.TicketID, receipt.TicketID)
	assert.Equal(t, ticket.Price.Amount, receipt.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, receipt.Price.Currency)
}

func assertTicketAddedToSpreadsheet(
	t *testing.T,
	ticketStatus TicketStatus,
	spreadsheetName string,
	spreadsheetsService *clients.SpreadsheetsServiceMock,
) {
	var appendedRows [][]string
	assert.EventuallyWithT(
		t,
		func(collect *assert.CollectT) {
			appendedRows = spreadsheetsService.AppendedRows[spreadsheetName]

			refundedTickets := len(appendedRows)
			t.Log(spreadsheetName, "appended rows", refundedTickets)

			assert.Greater(collect, refundedTickets, 0, "no rows appended to %s", spreadsheetName)
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var row []string
	var ok bool
	for _, r := range appendedRows {
		if r[0] != ticketStatus.TicketID {
			continue
		}
		row = r
		ok = true
		break
	}
	require.Truef(t, ok, "row in %s for ticket %s not found", spreadsheetName, ticketStatus.TicketID)

	assert.Equal(
		t,
		[]string{
			ticketStatus.TicketID,
			ticketStatus.Email,
			ticketStatus.Price.Amount,
			ticketStatus.Price.Currency,
		},
		row,
	)
}

func assertTicketsBookedInDeadNation(t *testing.T, deadNationService *clients.DeadNationClientMock, deadNationEventID string, numberOfTickets int) {
	t.Helper()

	assert.EventuallyWithT(t, func(collect *assert.CollectT) {
		deadNationBookings := deadNationService.DeadNationBookings
		t.Log("booked tickets", deadNationBookings)
		bookedTickets := 0
		for _, booking := range deadNationBookings {
			if booking.DeadNationEventID.String() != deadNationEventID {
				continue
			}
			bookedTickets += booking.NumberOfTickets
		}

		assert.Equal(collect, bookedTickets, numberOfTickets, "incorrect tickets booked in DeadNation")
	}, 10*time.Second, 100*time.Millisecond)
}

func assertTicketFileSaved(t *testing.T, filesService *clients.FilesServiceMock, ticket TicketStatus) {
	var savedBody string
	var ok bool
	var ticketFileID = ticket.TicketID + "-ticket.html"
	assert.EventuallyWithT(
		t,
		func(collect *assert.CollectT) {
			savedBody, ok = filesService.SavedFiles[ticketFileID]

			assert.True(collect, ok, "file for ticket %s not found", ticket.TicketID)
		},
		10*time.Second,
		100*time.Millisecond,
	)

	assert.Contains(t, savedBody, ticket.TicketID)
	assert.Contains(t, savedBody, ticket.Price.Amount)
	assert.Contains(t, savedBody, ticket.Price.Currency)
}

func assertTicketSavedInDb(t *testing.T, db *sqlx.DB, ticketStatus TicketStatus) {
	var tickets []entities.Ticket

	assert.EventuallyWithT(
		t,
		func(collect *assert.CollectT) {
			err := db.SelectContext(
				context.Background(),
				&tickets,
				`SELECT 
					ticket_id, 
					customer_email, 
					price_amount as "price.amount", 
					price_currency as "price.currency" 
				FROM tickets`,
			)
			assert.NoError(collect, err)

			assert.Greater(collect, len(tickets), 0)
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var ok bool
	for _, ticket := range tickets {
		if ticket.TicketID != ticketStatus.TicketID {
			continue
		}
		ok = true
		assert.Equal(t, ticketStatus.Email, ticket.CustomerEmail)
		assert.Equal(t, ticketStatus.Price.Amount, ticket.Price.Amount)
		assert.Equal(t, ticketStatus.Price.Currency, ticket.Price.Currency)
		break
	}
	assert.True(t, ok, "ticket %s not found in db", ticketStatus.TicketID)
}

func assertTicketDeletedFromDb(t *testing.T, db *sqlx.DB, ticketStatus TicketStatus) {
	var tickets []entities.Ticket

	assert.EventuallyWithT(
		t,
		func(collect *assert.CollectT) {
			err := db.SelectContext(
				context.Background(),
				&tickets,
				`SELECT 
					ticket_id, 
					customer_email, 
					price_amount as "price.amount", 
					price_currency as "price.currency" 
				FROM tickets`,
			)
			assert.NoError(collect, err)

			assert.Greater(collect, len(tickets), 0)
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var ok bool
	for _, ticket := range tickets {
		if ticket.TicketID != ticketStatus.TicketID {
			continue
		}
		ok = true
		break
	}
	assert.False(t, ok, "ticket %s found in db", ticketStatus.TicketID)
}

func assertTicketRefunded(t *testing.T, spreadsheetsService *clients.SpreadsheetsServiceMock, ticketIDs ...string) {
	refundedTickets := make(map[string][]string)
	assert.EventuallyWithT(
		t,
		func(collect *assert.CollectT) {
			appendedRows := spreadsheetsService.AppendedRows["tickets-to-refund"]
			for _, row := range appendedRows {
				refundedTickets[row[0]] = row
			}
			for _, ticketID := range ticketIDs {
				_, ok := refundedTickets[ticketID]
				assert.True(collect, ok, "ticket %s not found in refunded tickets", ticketID)
			}
		},
		10*time.Second,
		100*time.Millisecond,
	)
}
