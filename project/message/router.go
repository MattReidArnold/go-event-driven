package message

import (
	"fmt"
	"tickets/message/command"
	"tickets/message/event"
	"tickets/message/outbox"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewRouter(
	postgresSubscriber message.Subscriber,
	publisher message.Publisher,
	eventProcessorConfig cqrs.EventProcessorConfig,
	eventHandler *event.Handler,
	commandProcessorConfig cqrs.CommandProcessorConfig,
	commandHandler *command.Handler,
	watermillLogger watermill.LoggerAdapter,

) *message.Router {
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(fmt.Errorf("failed to create new router: %w", err))
	}

	useMiddlewares(router, watermillLogger)

	outbox.AddForwarderHandler(postgresSubscriber, publisher, router, watermillLogger)

	ep, err := cqrs.NewEventProcessorWithConfig(
		router,
		eventProcessorConfig,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create new event processor: %w", err))
	}

	err = ep.AddHandlers(
		cqrs.NewEventHandler(
			"StoreTicket",
			eventHandler.StoreTicket,
		),
		cqrs.NewEventHandler(
			"IssueReceipt",
			eventHandler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			"AppendToTicketsToPrint",
			eventHandler.AppendToTicketsToPrintSpreadsheet,
		),
		cqrs.NewEventHandler(
			"PrintTicket",
			eventHandler.PrintTicket,
		),
		cqrs.NewEventHandler(
			"AppendToTicketsToRefund",
			eventHandler.AppendToTicketsToRefundSpreadsheet,
		),
		cqrs.NewEventHandler(
			"RemoveTicket",
			eventHandler.RemoveTicket,
		),
		cqrs.NewEventHandler(
			"MakeDeadNationBooking",
			eventHandler.MakeDeadNationBooking,
		),
	)
	if err != nil {
		panic(fmt.Errorf("adding event handlers: %w", err))
	}

	command.RegisterCommandProcessor(commandProcessorConfig, router, commandHandler)
	return router
}
