package db_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"tickets/db"
	"tickets/entities"
	tixhttp "tickets/http"

	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookingsRepo(t *testing.T) {

	t.Run("AddBooking", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		showsRepo := db.NewShowsRepo(getDb())
		bookingsRepo := db.NewBookingsRepository(getDb())

		show := entities.Show{
			ShowID:          uuid.NewString(),
			DeadNationID:    uuid.NewString(),
			NumberOfTickets: 4,
			StartTime:       time.Now(),
			Title:           gofakeit.HackerPhrase(),
			Venue:           gofakeit.Address().Address,
		}

		err := showsRepo.AddShow(ctx, show)
		require.NoError(t, err)

		booking := entities.Booking{
			BookingID:       uuid.NewString(),
			ShowID:          show.ShowID,
			NumberOfTickets: 4,
			CustomerEmail:   gofakeit.Email(),
		}

		err = bookingsRepo.AddBooking(ctx, booking)
		require.NoError(t, err)

		var bookingFromDb entities.Booking
		err = getDb().GetContext(ctx, &bookingFromDb, `SELECT * FROM bookings WHERE booking_id = $1`, booking.BookingID)
		assert.NoError(t, err)

		assert.Equal(t, booking.BookingID, bookingFromDb.BookingID)
		assert.Equal(t, booking.ShowID, bookingFromDb.ShowID)
		assert.Equal(t, booking.NumberOfTickets, bookingFromDb.NumberOfTickets)
		assert.Equal(t, booking.CustomerEmail, bookingFromDb.CustomerEmail)
	})

	t.Run("overbooking", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		showsRepo := db.NewShowsRepo(getDb())
		bookingsRepo := db.NewBookingsRepository(getDb())

		capacity := 10

		show := entities.Show{
			ShowID:          uuid.NewString(),
			DeadNationID:    uuid.NewString(),
			NumberOfTickets: capacity,
			StartTime:       time.Now(),
			Title:           gofakeit.HackerPhrase(),
			Venue:           gofakeit.Address().Address,
		}
		err := showsRepo.AddShow(ctx, show)
		require.NoError(t, err)

		for i := 0; i < capacity; i++ {
			booking := entities.Booking{
				BookingID:       uuid.NewString(),
				ShowID:          show.ShowID,
				NumberOfTickets: 1,
				CustomerEmail:   gofakeit.Email(),
			}
			err := bookingsRepo.AddBooking(ctx, booking)
			require.NoError(t, err)
		}

		overbooking := entities.Booking{
			BookingID:       uuid.NewString(),
			ShowID:          show.ShowID,
			NumberOfTickets: 1,
			CustomerEmail:   gofakeit.Email(),
		}

		err = bookingsRepo.AddBooking(ctx, overbooking)
		require.ErrorIs(t, err, tixhttp.ErrInsufficientSeats)
	})

	t.Run("concurrent booking", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		showsRepo := db.NewShowsRepo(getDb())
		bookingsRepo := db.NewBookingsRepository(getDb())

		capacity := 1

		show := entities.Show{
			ShowID:          uuid.NewString(),
			DeadNationID:    uuid.NewString(),
			NumberOfTickets: capacity,
			StartTime:       time.Now(),
			Title:           gofakeit.HackerPhrase(),
			Venue:           gofakeit.Address().Address,
		}
		err := showsRepo.AddShow(ctx, show)
		require.NoError(t, err)

		workersCount := 50

		workersErrs := make(chan error, workersCount)
		unlock := make(chan struct{}, 1)
		wg := sync.WaitGroup{}
		wg.Add(workersCount)
		for i := 0; i < workersCount; i++ {
			go func() {
				defer wg.Done()

				booking := entities.Booking{
					BookingID:       uuid.NewString(),
					ShowID:          show.ShowID,
					NumberOfTickets: 1,
					CustomerEmail:   gofakeit.Email(),
				}

				<-unlock

				workersErrs <- bookingsRepo.AddBooking(ctx, booking)
			}()
		}
		close(unlock)
		wg.Wait()
		close(workersErrs)

		bookedCount, rejectedCount := 0, 0
		for err := range workersErrs {
			if err == nil {
				bookedCount++
			} else {
				assert.ErrorIs(t, err, tixhttp.ErrInsufficientSeats)
				rejectedCount++
			}
		}
		assert.Equal(t, capacity, bookedCount)
		assert.Equal(t, workersCount-bookedCount, rejectedCount)
	})
}
