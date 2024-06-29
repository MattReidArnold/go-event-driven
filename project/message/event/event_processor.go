package event

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

type PubSubAdapter interface {
	NewSubscriber(consumerGroup string) (message.Subscriber, error)
}

func RegisterEventHandlers(
	pubSubAdapter PubSubAdapter,
	router *message.Router,
	logger watermill.LoggerAdapter,
	handler *Handler,
) error {
	ep, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return pubSubAdapter.NewSubscriber("svc-tickets." + params.HandlerName)
			},
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: cqrs.JSONMarshaler{
				GenerateName: cqrs.StructName,
			},
			Logger: logger,
		},
	)
	if err != nil {
		return err
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
	return err
}
