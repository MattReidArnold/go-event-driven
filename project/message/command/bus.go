package command

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewCommandBus(pub message.Publisher, config cqrs.CommandBusConfig) *cqrs.CommandBus {
	bus, err := cqrs.NewCommandBusWithConfig(pub, config)
	if err != nil {
		panic(fmt.Errorf("creating command bus: %w", err))
	}
	return bus
}
