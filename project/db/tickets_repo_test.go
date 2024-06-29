package db_test

import (
	"context"
	"testing"
	tixDB "tickets/db"
	"tickets/entities"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"
)

func TestTicketsRepositoryAddTicketIsIdempotent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticket := entities.Ticket{
		TicketID:      uuid.NewString(),
		CustomerEmail: uuid.NewString() + "@" + uuid.NewString() + ".com",
		Price: entities.Money{
			Amount:   "2000.00",
			Currency: "USD",
		},
	}
	repo := tixDB.NewTicketsRepo(getDb())

	for i := 0; i < 5; i++ {
		assert.NoError(t, repo.AddTicket(ctx, ticket))
	}

	var tickets []entities.Ticket
	err := getDb().SelectContext(
		ctx,
		&tickets,
		`SELECT 
			ticket_id, 
			customer_email, 
			price_amount as "price.amount", 
			price_currency as "price.currency" 
		FROM tickets
		WHERE ticket_id = $1`,
		ticket.TicketID,
	)
	assert.NoError(t, err)

	assert.Len(t, tickets, 1, "wrong number of ticket %s found", ticket.TicketID)

}
