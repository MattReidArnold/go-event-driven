package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"

	redis "github.com/redis/go-redis/v9"
)

func main() {
	logger := watermill.NewStdLogger(false, false)

	rdb := redis.NewClient(&redis.Options{ // Update the package name to "redis"
		Addr: os.Getenv("REDIS_ADDR"),
	})
	subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}

	messages, err := subscriber.Subscribe(context.Background(), "progress")
	if err != nil {
		panic(err)
	}
	defer subscriber.Close()

	for msg := range messages {
		progress := string(msg.Payload)
		fmt.Printf("Message ID: %s - %s%%\n", msg.UUID, progress)

		msg.Ack()
	}

}
