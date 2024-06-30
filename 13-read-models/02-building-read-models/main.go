package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/shopspring/decimal"
)

type InvoiceIssued struct {
	InvoiceID    string
	CustomerName string
	Amount       decimal.Decimal
	IssuedAt     time.Time
}

type InvoicePaymentReceived struct {
	PaymentID  string
	InvoiceID  string
	PaidAmount decimal.Decimal
	PaidAt     time.Time

	FullyPaid bool
}

type InvoiceVoided struct {
	InvoiceID string
	VoidedAt  time.Time
}

type InvoiceReadModel struct {
	InvoiceID    string
	CustomerName string
	Amount       decimal.Decimal
	IssuedAt     time.Time

	FullyPaid     bool
	PaidAmount    decimal.Decimal
	LastPaymentAt time.Time

	Voided   bool
	VoidedAt time.Time
}

type InvoiceReadModelStorage struct {
	invoices map[string]InvoiceReadModel

	payments map[string]struct{}
}

func NewInvoiceReadModelStorage() *InvoiceReadModelStorage {
	return &InvoiceReadModelStorage{
		invoices: make(map[string]InvoiceReadModel),
	}
}

func (s *InvoiceReadModelStorage) Invoices() []InvoiceReadModel {
	invoices := make([]InvoiceReadModel, 0, len(s.invoices))
	for _, invoice := range s.invoices {
		invoices = append(invoices, invoice)
	}
	return invoices
}

func (s *InvoiceReadModelStorage) InvoiceByID(id string) (InvoiceReadModel, bool) {
	invoice, ok := s.invoices[id]
	return invoice, ok
}

func (s *InvoiceReadModelStorage) OnInvoiceIssued(ctx context.Context, event *InvoiceIssued) error {
	_, exists := s.InvoiceByID(event.InvoiceID)
	if exists {
		return nil
	}
	model := InvoiceReadModel{
		InvoiceID:    event.InvoiceID,
		CustomerName: event.CustomerName,
		Amount:       event.Amount,
		IssuedAt:     event.IssuedAt,
	}
	s.invoices[event.InvoiceID] = model
	return nil
}

func (s *InvoiceReadModelStorage) OnInvoicePaymentReceived(ctx context.Context, event *InvoicePaymentReceived) error {
	model, modelExists := s.InvoiceByID(event.InvoiceID)
	if !modelExists {
		return fmt.Errorf("invoice %s not found", event.InvoiceID)
	}

	_, paymentExists := s.payments[event.PaymentID]
	if paymentExists {
		return nil
	}

	model.PaidAmount = model.PaidAmount.Add(event.PaidAmount)
	model.FullyPaid = model.PaidAmount.Equal(model.Amount)
	model.LastPaymentAt = event.PaidAt

	s.invoices[event.InvoiceID] = model
	s.payments[event.PaymentID] = struct{}{}

	return nil
}

func (s *InvoiceReadModelStorage) OnInvoiceVoided(ctx context.Context, event *InvoiceVoided) error {
	model, exists := s.InvoiceByID(event.InvoiceID)
	if !exists {
		return fmt.Errorf("invoice %s not found", event.InvoiceID)
	}

	model.Voided = true
	model.VoidedAt = event.VoidedAt

	s.invoices[event.InvoiceID] = model

	return nil
}

func NewRouter(storage *InvoiceReadModelStorage, eventProcessorConfig cqrs.EventProcessorConfig, watermillLogger watermill.LoggerAdapter) (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		return nil, fmt.Errorf("could not create router: %w", err)
	}

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, eventProcessorConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create command processor: %w", err)
	}

	err = eventProcessor.AddHandlers(
		// TODO: add event handlers
		cqrs.NewEventHandler("OnInvoiceIssued", storage.OnInvoiceIssued),
		cqrs.NewEventHandler("OnInvoicePaymentReceived", storage.OnInvoicePaymentReceived),
		cqrs.NewEventHandler("OnInvoiceVoided", storage.OnInvoiceVoided),
	)
	if err != nil {
		return nil, fmt.Errorf("could not add event handlers: %w", err)
	}

	return router, nil
}
