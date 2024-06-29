package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type AlarmClient interface {
	StartAlarm() error
	StopAlarm() error
}

func ConsumeMessages(sub message.Subscriber, alarmClient AlarmClient) {
	messages, err := sub.Subscribe(context.Background(), "smoke_sensor")
	if err != nil {
		panic(err)
	}

	for msg := range messages {
		detection := string(msg.Payload)
		var err error
		switch detection {
		case "0": //no smoke detected
			err = alarmClient.StopAlarm()
		case "1": //smoke detected
			err = alarmClient.StartAlarm()
		}
		if err != nil {
			msg.Nack()
			continue
		}

		msg.Ack()
	}
}
