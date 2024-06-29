package command

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (c *Handler) RefundTicket(ctx context.Context, cmd *entities.RefundTicket) error {
	logger := log.FromContext(ctx)
	logger.Infof("refunding ticket %s", cmd.TicketID)
	// handle command
	return nil
}
