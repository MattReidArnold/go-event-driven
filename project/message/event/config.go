package event

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

var marshaler = cqrs.JSONMarshaler{
	GenerateName: cqrs.StructName,
}

func NewProcessorConfig(redisClient *redis.Client, watermillLogger watermill.LoggerAdapter) cqrs.EventProcessorConfig {
	return cqrs.EventProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return "events." + params.EventName, nil
		},
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return redisstream.NewSubscriber(redisstream.SubscriberConfig{
				Client:        redisClient,
				ConsumerGroup: "svc-tickets." + params.HandlerName,
			}, watermillLogger)
		},
		Marshaler: marshaler,
		Logger:    watermillLogger,
	}
}

func NewBusConfig(watermillLogger watermill.LoggerAdapter) cqrs.EventBusConfig {
	return cqrs.EventBusConfig{
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.StructName,
		},
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return "events." + params.EventName, nil
		},
		Logger: watermillLogger,
	}
}
