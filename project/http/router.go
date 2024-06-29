package http

import (
	"net/http"
	"time"

	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
)

type TicketsStatusRequest struct {
	Tickets []TicketStatus `json:"tickets"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type TicketStatus struct {
	TicketID      string `json:"ticket_id"`
	Status        string `json:"status"`
	Price         Money  `json:"price"`
	CustomerEmail string `json:"customer_email"`
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
	ShowID          string `json:"show_id"`
	NumberOfTickets int    `json:"number_of_tickets"`
	CustomerEmail   string `json:"customer_email"`
}

type PostBookTicketsResponse struct {
	BookingID string `json:"booking_id"`
}

func NewHttpRouter(
	commandBus *cqrs.CommandBus,
	eventBus *cqrs.EventBus,
	spreadsheetsAPI SpreadsheetsAPI,
	ticketsRepo TicketsRepository,
	showsRepository ShowsRepository,
	bookingsRepository BookingsRepository,
) *echo.Echo {
	e := commonHTTP.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	handler := Handler{
		commandBus:         commandBus,
		eventBus:           eventBus,
		spreadsheetsAPI:    spreadsheetsAPI,
		ticketsRepository:  ticketsRepo,
		showsRepository:    showsRepository,
		bookingsRepository: bookingsRepository,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)

	e.GET("/tickets", handler.GetTickets)

	e.POST("/shows", handler.PostShows)

	e.POST("/book-tickets", handler.PostBookTickets)

	e.PUT("/ticket-refund/:ticket_id", handler.PutTicketRefund)

	return e
}
