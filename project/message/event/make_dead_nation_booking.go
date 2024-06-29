package event

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/google/uuid"
)

func (h *Handler) MakeDeadNationBooking(ctx context.Context, event *entities.BookingMade) error {
	log.FromContext(ctx).Info("making Dead Nation booking")

	show, err := h.showsRepository.FindByID(ctx, event.ShowID)
	if err != nil {
		return fmt.Errorf("failed to get show: %w", err)
	}

	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return fmt.Errorf("failed to parse booking ID: %w", err)
	}

	deadNationEventID, err := uuid.Parse(show.DeadNationID)
	if err != nil {
		return fmt.Errorf("failed to parse Dead Nation event ID: %w", err)
	}

	request := entities.DeadNationBookingRequest{
		CustomerEmail:     event.CustomerEmail,
		DeadNationEventID: deadNationEventID,
		NumberOfTickets:   event.NumberOfTickets,
		BookingID:         bookingID,
	}
	err = h.deadNationService.BookTickets(
		ctx,
		request,
	)
	if err != nil {
		return fmt.Errorf("failed to book tickets: %w", err)
	}
	return nil
}
