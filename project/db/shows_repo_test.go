package db_test

import (
	"context"
	"testing"
	tixdb "tickets/db"
	"tickets/entities"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowsRepoAddsAShow(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := getDb()

	showID := uuid.NewString()

	repo := tixdb.NewShowsRepo(db)
	show := entities.Show{
		ShowID:          showID,
		DeadNationID:    uuid.NewString(),
		NumberOfTickets: 200,
		StartTime:       time.Now().UTC(),
		Title:           gofakeit.StreetName(),
		Venue:           gofakeit.City(),
	}
	err := repo.AddShow(ctx, show)
	assert.NoError(t, err)

	showFromDb := entities.Show{}
	err = db.Get(&showFromDb, `SELECT * FROM shows WHERE show_id = $1`, showID)
	require.NoError(t, err)

	assert.Equal(t, show.ShowID, showFromDb.ShowID)
	assert.Equal(t, show.DeadNationID, showFromDb.DeadNationID)
	assert.Equal(t, show.NumberOfTickets, showFromDb.NumberOfTickets)
	assert.Equal(t, show.StartTime, showFromDb.StartTime)
	assert.Equal(t, show.Title, showFromDb.Title)
	assert.Equal(t, show.Venue, showFromDb.Venue)
}
