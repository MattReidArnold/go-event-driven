package message

import (
	"fmt"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewRedisPublisher(rdb *redis.Client, watermillLogger watermill.LoggerAdapter) message.Publisher {
	var pub message.Publisher
	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, watermillLogger)
	if err != nil {
		panic(err)
	}
	pub = log.CorrelationPublisherDecorator{pub}

	return pub
}

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func NewRedisSubscriber(rdb *redis.Client, watermillLogger watermill.LoggerAdapter, consumerGroup string) *redisstream.Subscriber {
	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		ConsumerGroup: consumerGroup,
		Client:        rdb,
	}, watermillLogger)
	if err != nil {
		panic(fmt.Errorf("creating redis subscriber: %w ", err))
	}

	return sub
}
