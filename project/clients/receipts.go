package clients

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
)

type receiptsClient struct {
	clients *clients.Clients
}

func NewReceiptsClient(clients *clients.Clients) receiptsClient {
	return receiptsClient{
		clients: clients,
	}
}

func (c receiptsClient) IssueReceipt(ctx context.Context, req entities.IssueReceiptRequest) error {
	idempotencyKey := req.IdempotencyKey + req.TicketID
	body := receipts.PutReceiptsJSONRequestBody{
		IdempotencyKey: &idempotencyKey,
		TicketId:       req.TicketID,
		Price: receipts.Money{
			MoneyAmount:   req.Price.Amount,
			MoneyCurrency: req.Price.Currency,
		},
	}

	receiptsResp, err := c.clients.Receipts.PutReceiptsWithResponse(ctx, body)
	if err != nil {
		return err
	}
	if receiptsResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", receiptsResp.StatusCode())
	}

	return nil
}

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
