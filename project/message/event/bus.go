package event

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewEventBus(pub message.Publisher, config cqrs.EventBusConfig) *cqrs.EventBus {
	bus, err := cqrs.NewEventBusWithConfig(pub, config)
	if err != nil {
		panic(fmt.Errorf("creating event bus: %w", err))
	}
	return bus
}
