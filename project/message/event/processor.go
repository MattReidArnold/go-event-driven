package event

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func RegisterEventHandlers(
	router *message.Router,
	config cqrs.EventProcessorConfig,
	handler *Handler,
	logger watermill.LoggerAdapter,
) {
	ep, err := cqrs.NewEventProcessorWithConfig(router, config)
	if err != nil {
		panic(fmt.Errorf("creating new event processor: %w", err))
	}
	err = ep.AddHandlers([]cqrs.EventHandler{
		cqrs.NewEventHandler(
			"StoreTicket",
			handler.StoreTicket,
		),
		cqrs.NewEventHandler(
			"IssueReceipt",
			handler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			"AppendToTicketsToPrint",
			handler.AppendToTicketsToPrintSpreadsheet,
		),
		cqrs.NewEventHandler(
			"PrintTicket",
			handler.PrintTicket,
		),
		cqrs.NewEventHandler(
			"AppendToTicketsToRefund",
			handler.AppendToTicketsToRefundSpreadsheet,
		),
		cqrs.NewEventHandler(
			"RemoveTicket",
			handler.RemoveTicket,
		),
		cqrs.NewEventHandler(
			"MakeDeadNationBooking",
			handler.MakeDeadNationBooking,
		),
	}...)
	if err != nil {
		panic(fmt.Errorf("adding event handlers: %w", err))
	}
}
