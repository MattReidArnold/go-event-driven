package event

import (
	"context"
	"tickets/entities"
)

func (h *Handler) AppendToTicketsToPrintSpreadsheet(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	currency := event.Price.Currency
	if currency == "" {
		currency = "USD"
	}
	return h.spreadsheetsAPI.AppendRow(
		ctx,
		"tickets-to-print",
		[]string{event.TicketID, event.CustomerEmail, event.Price.Amount, currency},
	)
}
