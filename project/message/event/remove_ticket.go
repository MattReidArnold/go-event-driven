package event

import (
	"context"
	"fmt"
	"tickets/entities"
)

func (h *Handler) RemoveTicket(ctx context.Context, event *entities.TicketBookingCanceled) error {
	err := h.ticketsRepository.RemoveTicket(ctx, event.TicketID)
	if err != nil {
		return fmt.Errorf("unable to remove ticket: %w", err)
	}
	return nil
}
