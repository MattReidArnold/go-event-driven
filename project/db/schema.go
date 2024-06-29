package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const schema string = `
CREATE TABLE IF NOT EXISTS tickets (
	ticket_id UUID PRIMARY KEY,
	price_amount NUMERIC(10, 2) NOT NULL,
	price_currency CHAR(3) NOT NULL,
	customer_email VARCHAR(255) NOT NULL
	);
CREATE TABLE IF NOT EXISTS shows (
	show_id UUID PRIMARY KEY,
	dead_nation_id VARCHAR(255) NOT NULL,
	number_of_tickets INTEGER NOT NULL,
	start_time TIMESTAMP WITH TIME ZONE NOT NULL,
	title VARCHAR(255) NOT NULL,
	venue VARCHAR(255) NOT NULL
);
CREATE TABLE IF NOT EXISTS bookings (
	booking_id UUID PRIMARY KEY,
	show_id UUID NOT NULL REFERENCES shows(show_id),
	number_of_tickets INTEGER NOT NULL,
	customer_email VARCHAR(255) NOT NULL
);
`

func InitializeSchema(db *sqlx.DB) error {
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("could not initialize database schema: %w", err)
	}
	return nil
}
