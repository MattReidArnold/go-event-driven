package main

import "context"

type PaymentTaken struct {
	PaymentID string
	Amount    int
}

type PaymentsHandler struct {
	repo *PaymentsRepository
}

func NewPaymentsHandler(repo *PaymentsRepository) *PaymentsHandler {
	return &PaymentsHandler{repo: repo}
}

func (p *PaymentsHandler) HandlePaymentTaken(ctx context.Context, event *PaymentTaken) error {
	return p.repo.SavePaymentTaken(ctx, event)
}

type PaymentsRepository struct {
	payments map[string]PaymentTaken
}

func (p *PaymentsRepository) Payments() []PaymentTaken {
	payments := make([]PaymentTaken, 0, len(p.payments))
	for _, payment := range p.payments {
		payments = append(payments, payment)
	}
	return payments
}

func NewPaymentsRepository() *PaymentsRepository {
	return &PaymentsRepository{}
}

func (p *PaymentsRepository) SavePaymentTaken(ctx context.Context, event *PaymentTaken) error {
	if p.payments == nil {
		p.payments = make(map[string]PaymentTaken)
	}
	p.payments[event.PaymentID] = *event
	return nil
}
