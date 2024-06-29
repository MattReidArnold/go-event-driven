package event

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/jmoiron/sqlx"
)

type Handler struct {
	ticketsRepository TicketsRepository
	showsRepository   ShowsRepository
	receiptService    ReceiptService
	spreadsheetsAPI   SpreadsheetsAPI
	deadNationService DeadNationService
	filesService      FilesService
	eventBus          *cqrs.EventBus
}

func NewEventHandler(
	ticketsRepository TicketsRepository,
	showsRepository ShowsRepository,
	receiptService ReceiptService,
	spreadsheetsAPI SpreadsheetsAPI,
	deadNationService DeadNationService,
	filesService FilesService,
	eventBus *cqrs.EventBus,
) *Handler {
	return &Handler{
		ticketsRepository: ticketsRepository,
		showsRepository:   showsRepository,
		receiptService:    receiptService,
		spreadsheetsAPI:   spreadsheetsAPI,
		deadNationService: deadNationService,
		filesService:      filesService,
		eventBus:          eventBus,
	}
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
	FindByID(ctx context.Context, showID string) (entities.Show, error)
}
type BookingsRepository interface {
	AddBooking(ctx context.Context, booking entities.Booking) error
	AddBookingInTx(ctx context.Context, booking entities.Booking, tx *sqlx.Tx) error
}

type DeadNationService interface {
	BookTickets(ctx context.Context, req entities.DeadNationBookingRequest) error
}
