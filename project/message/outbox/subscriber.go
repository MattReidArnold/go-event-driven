package outbox

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
)

func NewPostgresSubscriber(db *sql.DB, logger watermill.LoggerAdapter) *watermillSQL.Subscriber {
	subscriber, err := watermillSQL.NewSubscriber(
		db,
		watermillSQL.SubscriberConfig{
			PollInterval:   time.Millisecond * 100,
			SchemaAdapter:  watermillSQL.DefaultPostgreSQLSchema{},
			OffsetsAdapter: watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
		},
		logger,
	)
	if err != nil {
		panic(fmt.Errorf("creating watermill sql subscriber: %w", err))
	}

	err = subscriber.SubscribeInitialize(outboxTopic)
	if err != nil {
		panic(fmt.Errorf("subscribing to outbox topic: %w", err))
	}
	return subscriber
}
