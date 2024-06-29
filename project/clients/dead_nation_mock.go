package clients

import (
	"context"
	"sync"
	"tickets/entities"
)

type DeadNationClientMock struct {
	lock sync.Mutex

	DeadNationBookings []entities.DeadNationBookingRequest
}

func (c *DeadNationClientMock) BookTickets(ctx context.Context, req entities.DeadNationBookingRequest) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.DeadNationBookings = append(c.DeadNationBookings, req)

	return nil
}
