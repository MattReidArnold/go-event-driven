package outbox

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

func AddForwarderHandler(
	postgresSubscriber message.Subscriber,
	publisher message.Publisher,
	router *message.Router,
	logger watermill.LoggerAdapter,
) error {
	// your code goes here
	// subscriber, err := watermillSQL.NewSubscriber(
	// 	db,
	// 	watermillSQL.SubscriberConfig{
	// 		SchemaAdapter:  watermillSQL.DefaultPostgreSQLSchema{},
	// 		OffsetsAdapter: watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
	// 	},
	// 	logger,
	// )
	// if err != nil {
	// 	return fmt.Errorf("failed to create subscriber: %w", err)
	// }

	// err = subscriber.SubscribeInitialize(outboxTopic)
	// if err != nil {
	// 	return fmt.Errorf("failed to initialize subscriber: %w", err)
	// }

	// pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
	// 	Client: rdb,
	// }, logger)
	// if err != nil {
	// 	return fmt.Errorf("failed to create publisher: %w", err)
	// }
	forward, err := forwarder.NewForwarder(
		postgresSubscriber,
		publisher,
		logger,
		forwarder.Config{
			ForwarderTopic: outboxTopic,
			Middlewares: []message.HandlerMiddleware{
				func(h message.HandlerFunc) message.HandlerFunc {
					return func(msg *message.Message) ([]*message.Message, error) {
						log.FromContext(msg.Context()).WithFields(logrus.Fields{
							"message_id": msg.UUID,
							"payload":    string(msg.Payload),
							"metadata":   msg.Metadata,
						}).Info("Forwarding message")
						return h(msg)
					}
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("creating forwarder: %w", err)
	}
	go func() {
		err := forward.Run(context.Background())
		if err != nil {
			panic(fmt.Errorf("running forwarder: %w", err))
		}
	}()
	<-forward.Running()

	return nil
}
