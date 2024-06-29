package command

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func RegisterCommandProcessor(config cqrs.CommandProcessorConfig, router *message.Router, handler *Handler) error {
	// commandProcessor, err := cqrs.NewCommandProcessorWithConfig(
	// 	router,
	// 	cqrs.CommandProcessorConfig{
	// 		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
	// 			return "commands", nil
	// 		},
	// 		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
	// 			return sub, nil
	// 		},
	// 		Marshaler: cqrs.JSONMarshaler{
	// 			GenerateName: cqrs.StructName,
	// 		},
	// 		Logger: watermillLogger,
	// 	},
	// )
	commandProcessor, err := cqrs.NewCommandProcessorWithConfig(router, config)
	if err != nil {
		return fmt.Errorf("creating new command processor: %w", err)
	}

	err = commandProcessor.AddHandlers(
		cqrs.NewCommandHandler("RefundTicket", handler.RefundTicket),
	)
	if err != nil {
		return fmt.Errorf("adding command handlers: %w", err)
	}
	return nil
}
