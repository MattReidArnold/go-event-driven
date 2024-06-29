package event

import (
	"context"
	"fmt"

	"tickets/entities"
)

func (h *Handler) StoreTicket(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	currency := event.Price.Currency
	if currency == "" {
		currency = "USD"
	}
	ticket := entities.Ticket{
		TicketID:      event.TicketID,
		CustomerEmail: event.CustomerEmail,
		Price: entities.Money{
			Amount:   event.Price.Amount,
			Currency: currency,
		},
	}
	err := h.ticketsRepository.AddTicket(ctx, ticket)
	if err != nil {
		return fmt.Errorf("could not store ticket: %w", err)
	}
	return nil
}
