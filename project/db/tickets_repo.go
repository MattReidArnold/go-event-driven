package db

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type ticketsRepo struct {
	db *sqlx.DB
}

func NewTicketsRepo(dbConn *sqlx.DB) *ticketsRepo {
	return &ticketsRepo{
		db: dbConn,
	}
}

func (r *ticketsRepo) AddTicket(ctx context.Context, ticket entities.Ticket) error {
	res, err := r.db.NamedExecContext(
		ctx,
		`INSERT INTO tickets (ticket_id, price_amount, price_currency, customer_email) 
		VALUES ( :ticket_id, :price.amount, :price.currency, :customer_email )
		ON CONFLICT (ticket_id) DO NOTHING`,
		ticket,
	)
	if err != nil {
		return fmt.Errorf("could not insert ticket: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 1 {
		logrus.Info("Ticket stored ", ticket.TicketID)
	}

	return nil
}

func (r *ticketsRepo) RemoveTicket(ctx context.Context, ticketID string) error {
	res, err := r.db.ExecContext(
		ctx,
		`DELETE FROM tickets WHERE ticket_id = $1`,
		ticketID,
	)
	if err != nil {
		return fmt.Errorf("could not delete ticket: %w", err)
	}
	logrus.Info("ticket removed", res)
	return nil
}

func (r *ticketsRepo) FindAll(ctx context.Context) ([]entities.Ticket, error) {
	var tickets []entities.Ticket
	err := r.db.SelectContext(
		ctx,
		&tickets,
		`SELECT 
			ticket_id, 
			customer_email, 
			price_amount as "price.amount", 
			price_currency as "price.currency" 
		FROM tickets`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to select from db: %w", err)
	}

	return tickets, nil
}
