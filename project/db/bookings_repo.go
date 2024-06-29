package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tickets/entities"
	"tickets/http"
	"tickets/message/event"
	"tickets/message/outbox"

	"github.com/jmoiron/sqlx"
)

type BookingsRepository struct {
	db *sqlx.DB
}

func NewBookingsRepository(db *sqlx.DB) BookingsRepository {
	if db == nil {
		panic("nil db")
	}

	return BookingsRepository{db: db}
}

func (b BookingsRepository) addBookingTxx(ctx context.Context, booking entities.Booking, tx *sqlx.Tx) error {
	availableSeats := 0
	err := tx.GetContext(ctx, &availableSeats, `
		SELECT
		    number_of_tickets AS available_seats
		FROM
		    shows
		WHERE
		    show_id = $1
	`, booking.ShowID)
	if err != nil {
		return fmt.Errorf("finding available seats: %w", err)
	}

	alreadyBookedSeats := 0
	err = tx.GetContext(ctx, &alreadyBookedSeats, `
		SELECT
		    coalesce(SUM(number_of_tickets), 0) AS already_booked_seats
		FROM
		    bookings
		WHERE
		    show_id = $1
	`, booking.ShowID)
	if err != nil {
		return fmt.Errorf("finding booked seats: %w", err)
	}

	if availableSeats-alreadyBookedSeats < booking.NumberOfTickets {
		return http.ErrInsufficientSeats
	}

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO 
		    bookings (booking_id, show_id, number_of_tickets, customer_email) 
		VALUES (:booking_id, :show_id, :number_of_tickets, :customer_email)
		`, booking)
	if err != nil {
		return fmt.Errorf("adding booking: %w", err)
	}

	outboxPublisher, err := outbox.NewPublisherForDb(ctx, tx)
	if err != nil {
		return fmt.Errorf("creating event bus: %w", err)
	}

	err = event.NewEventBus(outboxPublisher).Publish(ctx, entities.BookingMade{
		Header:          entities.NewEventHeader(),
		BookingID:       booking.BookingID,
		NumberOfTickets: booking.NumberOfTickets,
		CustomerEmail:   booking.CustomerEmail,
		ShowID:          booking.ShowID,
	})
	if err != nil {
		return fmt.Errorf("publishing BookingMade event: %w", err)
	}

	return nil
}

func (b BookingsRepository) AddBooking(ctx context.Context, booking entities.Booking) error {
	var err error
	tx, err := b.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	err = b.addBookingTxx(ctx, booking, tx)
	if err != nil {
		rollbackErr := tx.Rollback()
		err = errors.Join(err, rollbackErr)
		// var pqErr *pq.Error
		// // if errors.As(err, &pqErr) && pqErr.Code == "40001" { // serialization_failure
		// // 	// Sleep for a random interval between 5 and 25 milliseconds
		// // 	time.Sleep(time.Duration(math.Pow(2, float64(i+1))*10) * time.Millisecond)

		// // }
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return err
}
