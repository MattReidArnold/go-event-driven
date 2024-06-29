package command

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func RegisterCommandHandler(router *message.Router, config cqrs.CommandProcessorConfig, handler *Handler) error {
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
