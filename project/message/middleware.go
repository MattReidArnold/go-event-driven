package message

import (
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/sirupsen/logrus"
)

func useMiddlewares(router *message.Router, watermillLogger watermill.LoggerAdapter) {
	router.AddMiddleware(PurgeMessageMiddleware)
	router.AddMiddleware(LoggingMiddleware)
	router.AddMiddleware(CorrelationIDMiddleware)
	router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          watermillLogger,
	}.Middleware)
}

func LoggingMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		logger := log.FromContext(msg.Context())
		logger = logger.WithField("message_uuid", msg.UUID)

		logger.Info("Handling a message")

		msgs, err := next(msg)
		if err != nil {
			logger.WithError(err).Info("Message handling error")
		}
		return msgs, err
	}
}

func PurgeMessageMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		if msg.UUID == "2beaf5bc-d5e4-4653-b075-2b36bbf28949" {
			return nil, nil
		}
		return next(msg)
	}
}

func CorrelationIDMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		correlationID := msg.Metadata.Get("correlation_id")
		if correlationID == "" {
			correlationID = watermill.NewShortUUID()
			msg.Metadata.Set("correlation_id", correlationID)
		}
		ctx := log.ContextWithCorrelationID(msg.Context(), correlationID)
		ctx = log.ToContext(ctx, logrus.WithFields(logrus.Fields{"correlation_id": correlationID}))
		msg.SetContext(ctx)
		return next(msg)
	}
}

func RequireTypeMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		msgType := msg.Metadata.Get("type")
		if msgType == "" {
			logger := log.FromContext(msg.Context())
			logger.Info("Ignoring message because metadata 'type' is empty")
			return nil, nil
		}
		return next(msg)
	}
}
