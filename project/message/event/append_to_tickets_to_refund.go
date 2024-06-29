package event

import (
	"context"

	"tickets/entities"
)

func (h *Handler) AppendToTicketsToRefundSpreadsheet(ctx context.Context, event *entities.TicketBookingCanceled) error {
	currency := event.Price.Currency
	if currency == "" {
		currency = "USD"
	}
	return h.spreadsheetsAPI.AppendRow(
		ctx,
		"tickets-to-refund",
		[]string{event.TicketID, event.CustomerEmail, event.Price.Amount, currency},
	)
}
