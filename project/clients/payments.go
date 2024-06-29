package clients

import (
	"context"
	"fmt"
	"tickets/entities"
	"tickets/message/command"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/payments"
)

type PaymentsClient struct {
	clients *clients.Clients
}

func NewPaymentsClient(clients *clients.Clients) *PaymentsClient {
	return &PaymentsClient{clients: clients}
}

func (p *PaymentsClient) RefundTicket(ctx context.Context, command *entities.RefundTicket) error {
	_, err := p.clients.Payments.PutRefundsWithResponse(ctx, payments.PaymentRefundRequest{
		PaymentReference: command.TicketID,
		Reason:           "customer requested refund",
		DeduplicationId:  &command.Header.IdempotencyKey,
	})
	if err != nil {
		return fmt.Errorf("sending refund %s to Payments: %w", command.TicketID, err)
	}
	return nil
}

var _ command.PaymentRefunder = &PaymentsClient{}
