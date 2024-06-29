package http

import (
	"errors"
	"fmt"
	"net/http"
	"tickets/entities"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type BookTicketRequest struct {
	CustomerEmail   string `json:"customer_email"`
	NumberOfTickets int    `json:"number_of_tickets"`
	ShowID          string `json:"show_id"`
}

type BookTicketResponse struct {
	BookingId string `json:"booking_id"`
}

func (h Handler) PostBookTickets(c echo.Context) error {
	req := BookTicketRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	if req.NumberOfTickets < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "number of tickets must be greater than 0")
	}

	bookingID := uuid.NewString()

	err := h.bookingsRepository.AddBooking(c.Request().Context(), entities.Booking{
		BookingID:       bookingID,
		CustomerEmail:   req.CustomerEmail,
		NumberOfTickets: req.NumberOfTickets,
		ShowID:          req.ShowID,
	})
	if errors.Is(err, ErrInsufficientSeats) {
		return echo.NewHTTPError(http.StatusBadRequest, "insufficient seats")

	}
	if err != nil {
		return fmt.Errorf("failed to add booking: %w", err)
	}

	return c.JSON(
		http.StatusCreated,
		BookTicketResponse{
			BookingId: bookingID,
		},
	)
}
