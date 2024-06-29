package main

import (
	"fmt"
	"log"
	"time"
)

type User struct {
	Email string
}

type UserRepository interface {
	CreateUserAccount(u User) error
}

type NotificationsClient interface {
	SendNotification(u User) error
}

type NewsletterClient interface {
	AddToNewsletter(u User) error
}

type Handler struct {
	repository          UserRepository
	newsletterClient    NewsletterClient
	notificationsClient NotificationsClient
}

func NewHandler(
	repository UserRepository,
	newsletterClient NewsletterClient,
	notificationsClient NotificationsClient,
) Handler {
	return Handler{
		repository:          repository,
		newsletterClient:    newsletterClient,
		notificationsClient: notificationsClient,
	}
}

const (
	maxAttempts  = 5
	waitDuration = 200 * time.Millisecond
)

func (h Handler) SignUp(u User) error {
	if err := h.repository.CreateUserAccount(u); err != nil {
		return err
	}

	go retry("add user to newsletter", func() error {
		return h.newsletterClient.AddToNewsletter(u)
	})

	go retry("send notification", func() error {
		return h.notificationsClient.SendNotification(u)
	})

	return nil
}
func retry(operation string, action func() error) {
	for i := 0; i < maxAttempts-1; i++ {
		err := action()
		if err == nil {
			return
		}
		log.Printf("failed to %v: %v\n", operation, err)

		time.Sleep(waitDuration)
	}
	log.Printf("failed to %v: %v\n", operation, fmt.Errorf("out of retries"))
}
