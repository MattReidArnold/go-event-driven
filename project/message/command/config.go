package command

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewProcessorConfig(
	redisClient *redis.Client,
	watermillLogger watermill.LoggerAdapter,
) cqrs.CommandProcessorConfig {
	return cqrs.CommandProcessorConfig{
		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return redisstream.NewSubscriber(
				redisstream.SubscriberConfig{
					Client:        redisClient,
					ConsumerGroup: "svc-tickets.commands." + params.HandlerName,
				},
				watermillLogger,
			)
		},
		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
			return "commands." + params.CommandName, nil
		},
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.StructName,
		},
		Logger: watermillLogger,
	}
}

func NewBusConfig(watermillLogger watermill.LoggerAdapter) cqrs.CommandBusConfig {
	return cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return "commands." + params.CommandName, nil
		},
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.StructName,
		},
		Logger: watermillLogger,
	}
}
