package event

import (
	"context"
	"fmt"
	"tickets/entities"
)

func (h *Handler) IssueReceipt(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	currency := event.Price.Currency
	if currency == "" {
		currency = "USD"
	}
	resp, err := h.receiptService.IssueReceipt(ctx, entities.IssueReceiptRequest{
		IdempotencyKey: event.Header.IdempotencyKey,
		TicketID:       event.TicketID,
		Price: entities.Money{
			Amount:   event.Price.Amount,
			Currency: currency,
		},
	})
	if err != nil {
		return fmt.Errorf("calling receipt service: %w", err)
	}
	ticketReceiptIssued := entities.TicketReceiptIssued{
		Header:        entities.NewEventHeaderWithIdempotencyKey(event.Header.IdempotencyKey),
		TicketID:      event.TicketID,
		ReceiptNumber: resp.ReceiptNumber,
		IssuedAt:      resp.IssuedAt,
	}
	err = h.eventBus.Publish(ctx, ticketReceiptIssued)
	if err != nil {
		return fmt.Errorf("publishing ticket receipt issued event: %w", err)
	}

	return nil
}
