package command

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

type PaymentRefunder interface {
	RefundTicket(ctx context.Context, cmd *entities.RefundTicket) error
}

type ReceiptVoider interface {
	VoidReceipt(ctx context.Context, cmd *entities.RefundTicket) error
}

type Handler struct {
	paymentRefunder PaymentRefunder
	receiptVoider   ReceiptVoider
}

func NewHandler(paymentRefunder PaymentRefunder, receiptVoider ReceiptVoider) *Handler {
	return &Handler{paymentRefunder: paymentRefunder, receiptVoider: receiptVoider}
}

func (c *Handler) RefundTicket(ctx context.Context, cmd *entities.RefundTicket) error {
	logger := log.FromContext(ctx)
	logger.Infof("refunding ticket %s", cmd.TicketID)
	err := c.paymentRefunder.RefundTicket(ctx, cmd)
	if err != nil {
		return fmt.Errorf("refunding payment for ticket %s: %w", cmd.TicketID, err)
	}
	err = c.receiptVoider.VoidReceipt(ctx, cmd)
	if err != nil {
		return fmt.Errorf("voiding receipt for ticket %s: %w", cmd.TicketID, err)
	}
	return nil
}
