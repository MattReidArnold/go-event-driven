package event

import (
	"context"
	"tickets/entities"
)

func (h *Handler) IssueReceipt(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	currency := event.Price.Currency
	if currency == "" {
		currency = "USD"
	}
	return h.receiptService.IssueReceipt(ctx, entities.IssueReceiptRequest{
		IdempotencyKey: event.Header.IdempotencyKey,
		TicketID:       event.TicketID,
		Price: entities.Money{
			Amount:   event.Price.Amount,
			Currency: currency,
		},
	})
}
