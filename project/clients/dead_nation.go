package clients

import (
	"context"
	"fmt"
	"net/http"

	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/dead_nation"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

type DeadNationClient struct {
	clients *clients.Clients
}

func NewDeadNationClient(clients *clients.Clients) DeadNationClient {
	return DeadNationClient{
		clients: clients,
	}
}

func (c DeadNationClient) BookTickets(ctx context.Context, req entities.DeadNationBookingRequest) error {
	log.FromContext(ctx).Info(fmt.Sprintf("booking %d tickets for booking %s", req.NumberOfTickets, req.BookingID))
	resp, err := c.clients.DeadNation.PostTicketBookingWithResponse(
		ctx,
		dead_nation.PostTicketBookingRequest{
			CustomerAddress: req.CustomerEmail,
			EventId:         req.DeadNationEventID,
			NumberOfTickets: req.NumberOfTickets,
			BookingId:       req.BookingID,
		},
	)
	if err != nil {
		return fmt.Errorf("sending booking %s to DeadNation: %w", req.BookingID, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("DeadNation ticket booking %s status code: %v", req.BookingID, resp.StatusCode())
	}
	return nil
}
