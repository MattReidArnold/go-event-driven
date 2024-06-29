package entities

import "github.com/google/uuid"

type Booking struct {
	BookingID       string `db:"booking_id"`
	ShowID          string `db:"show_id"`
	NumberOfTickets int    `db:"number_of_tickets"`
	CustomerEmail   string `db:"customer_email"`
}

type DeadNationBookingRequest struct {
	CustomerEmail     string
	DeadNationEventID uuid.UUID
	NumberOfTickets   int
	BookingID         uuid.UUID
}
