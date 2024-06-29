package http

import (
	"context"
	"errors"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	eventBus           *cqrs.EventBus
	commandBus         *cqrs.CommandBus
	spreadsheetsAPI    SpreadsheetsAPI
	ticketsRepository  TicketsRepository
	showsRepository    ShowsRepository
	bookingsRepository BookingsRepository
}

type ReceiptService interface {
	IssueReceipt(ctx context.Context, req entities.IssueReceiptRequest) error
	// IssueReceipt(ctx context.Context, req entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error)
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type FilesService interface {
	UploadFile(ctx context.Context, fileID string, fileContent string) error
}

type TicketsRepository interface {
	AddTicket(ctx context.Context, ticket entities.Ticket) error
	RemoveTicket(ctx context.Context, ticketID string) error
	FindAll(ctx context.Context) ([]entities.Ticket, error)
}

type ShowsRepository interface {
	AddShow(ctx context.Context, show entities.Show) error
}

var ErrInsufficientSeats = errors.New("db: insufficient seats")

type BookingsRepository interface {
	AddBooking(ctx context.Context, booking entities.Booking) error
}
