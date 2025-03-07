package main

import (
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	redis "github.com/redis/go-redis/v9"
)

func main() {

	logger := watermill.NewStdLogger(false, false)

	rdb := redis.NewClient(&redis.Options{ // Update the package name to "redis"
		Addr: os.Getenv("REDIS_ADDR"),
	})

	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}

	msg := message.NewMessage(watermill.NewUUID(), []byte("50"))
	publisher.Publish("progress", msg)
	msg = message.NewMessage(watermill.NewUUID(), []byte("100"))
	publisher.Publish("progress", msg)

}
