package entities

import (
	"time"
)

type Show struct {
	ShowID          string    `db:"show_id" json:"show_id"`
	DeadNationID    string    `db:"dead_nation_id" json:"dead_nation_id"`
	NumberOfTickets int       `db:"number_of_tickets" json:"number_of_tickets"`
	StartTime       time.Time `db:"start_time" json:"start_time"`
	Title           string    `db:"title" json:"title"`
	Venue           string    `db:"venue" json:"venue"`
}
