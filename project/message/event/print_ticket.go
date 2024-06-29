package event

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

func (h *Handler) PrintTicket(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("printing ticket")

	ticketHTML := `
		<html>
			<head>
				<title>Ticket</title>
			</head>
			<body>
				<h1>Ticket ` + event.TicketID + `</h1>
				<p>Price: ` + event.Price.Amount + ` ` + event.Price.Currency + `</p>	
			</body>
		</html>
`
	fileName := event.TicketID + "-ticket.html"
	err := h.filesService.UploadFile(ctx, fileName, ticketHTML)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	ticketPrintedEvent := entities.TicketPrinted{
		Header:   entities.NewEventHeader(),
		TicketID: event.TicketID,
		FileName: fileName,
	}
	err = h.eventBus.Publish(ctx, ticketPrintedEvent)
	if err != nil {
		return fmt.Errorf("could not publish TicketPrinted: %w", err)
	}
	return nil
}
