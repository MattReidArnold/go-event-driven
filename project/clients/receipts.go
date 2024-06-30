package clients

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"tickets/entities"
	"tickets/message/command"
	"tickets/message/event"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
)

type receiptsClient struct {
	clients *clients.Clients
}

var _ event.ReceiptService = &receiptsClient{}

func NewReceiptsClient(clients *clients.Clients) receiptsClient {
	return receiptsClient{
		clients: clients,
	}
}

func (c receiptsClient) IssueReceipt(ctx context.Context, req entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error) {
	idempotencyKey := req.IdempotencyKey + req.TicketID
	body := receipts.PutReceiptsJSONRequestBody{
		IdempotencyKey: &idempotencyKey,
		TicketId:       req.TicketID,
		Price: receipts.Money{
			MoneyAmount:   req.Price.Amount,
			MoneyCurrency: req.Price.Currency,
		},
	}

	resp, err := c.clients.Receipts.PutReceiptsWithResponse(ctx, body)
	if err != nil {
		return entities.IssueReceiptResponse{}, fmt.Errorf("making call to receipts client: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		// receipt already exists
		return entities.IssueReceiptResponse{
			ReceiptNumber: resp.JSON200.Number,
			IssuedAt:      resp.JSON200.IssuedAt,
		}, nil
	case http.StatusCreated:
		// receipt was created
		return entities.IssueReceiptResponse{
			ReceiptNumber: resp.JSON201.Number,
			IssuedAt:      resp.JSON201.IssuedAt,
		}, nil
	default:
		return entities.IssueReceiptResponse{}, fmt.Errorf("unexpected status code for POST receipts-api/receipts: %d", resp.StatusCode())
	}
}

func (c receiptsClient) VoidReceipt(ctx context.Context, command *entities.RefundTicket) error {
	_, err := c.clients.Receipts.PutVoidReceiptWithResponse(ctx, receipts.VoidReceiptRequest{
		Reason:       "customer requested refund",
		TicketId:     command.TicketID,
		IdempotentId: &command.Header.IdempotencyKey,
	})
	if err != nil {
		return fmt.Errorf("PUT void receipt request %s: %w", command.TicketID, err)
	}
	return nil
}

var _ command.ReceiptVoider = receiptsClient{}

type ReceiptsServiceMock struct {
	mock           sync.Mutex
	IssuedReceipts []entities.IssueReceiptRequest
}

func (rs *ReceiptsServiceMock) IssueReceipt(ctx context.Context, req entities.IssueReceiptRequest) error {
	rs.mock.Lock()
	defer rs.mock.Unlock()

	rs.IssuedReceipts = append(rs.IssuedReceipts, req)

	return nil
}
