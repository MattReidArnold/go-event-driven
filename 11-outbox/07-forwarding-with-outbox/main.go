package main

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	_ "github.com/lib/pq"
)

func RunForwarder(
	db *sqlx.DB,
	rdb *redis.Client,
	outboxTopic string,
	logger watermill.LoggerAdapter,
) error {
	// your code goes here
	subscriber, err := watermillSQL.NewSubscriber(
		db,
		watermillSQL.SubscriberConfig{
			SchemaAdapter:  watermillSQL.DefaultPostgreSQLSchema{},
			OffsetsAdapter: watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
		},
		logger,
	)
	if err != nil {
		return err
	}

	err = subscriber.SubscribeInitialize(outboxTopic)
	if err != nil {
		return err
	}

	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		return err
	}
	forwarder, err := forwarder.NewForwarder(
		subscriber,
		pub,
		logger,
		forwarder.Config{
			ForwarderTopic: outboxTopic,
			Middlewares: []message.HandlerMiddleware{
				func(h message.HandlerFunc) message.HandlerFunc {
					return func(msg *message.Message) ([]*message.Message, error) {
						fmt.Println("Forwarding message", msg.UUID, string(msg.Payload), msg.Metadata)

						return h(msg)
					}
				},
			},
		},
	)

	go func() {
		err = forwarder.Run(context.Background())
		if err != nil {
			panic(err)
		}
	}()
	<-forwarder.Running()
	return nil
}
