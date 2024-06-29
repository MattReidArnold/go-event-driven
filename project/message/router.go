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

	event.RegisterEventHandlers(router, eventProcessorConfig, eventHandler, watermillLogger)
	command.RegisterCommandHandler(router, commandProcessorConfig, commandHandler)

	return router
}
