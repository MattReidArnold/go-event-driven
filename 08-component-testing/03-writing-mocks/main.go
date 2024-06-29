package main

import (
	"context"
	"sync"
	"time"
)

type IssueReceiptRequest struct {
	TicketID string `json:"ticket_id"`
	Price    Money  `json:"price"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type IssueReceiptResponse struct {
	ReceiptNumber string    `json:"number"`
	IssuedAt      time.Time `json:"issued_at"`
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request IssueReceiptRequest) (IssueReceiptResponse, error)
}

type ReceiptsServiceMock struct {
	mock           sync.Mutex
	IssuedReceipts []IssueReceiptRequest
}

func (rs *ReceiptsServiceMock) IssueReceipt(ctx context.Context, req IssueReceiptRequest) (IssueReceiptResponse, error) {
	rs.mock.Lock()
	defer rs.mock.Unlock()

	rs.IssuedReceipts = append(rs.IssuedReceipts, req)

	return IssueReceiptResponse{
		ReceiptNumber: "12345",
		IssuedAt:      time.Now(),
	}, nil
}
