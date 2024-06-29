package http

import (
	"fmt"
	"net/http"
	"tickets/entities"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h Handler) PostShows(c echo.Context) error {
	show := entities.Show{}

	if err := c.Bind(&show); err != nil {
		return fmt.Errorf("binding show: %w", err)
	}

	show.ShowID = uuid.NewString()

	err := h.showsRepository.AddShow(c.Request().Context(), show)
	if err != nil {
		return fmt.Errorf("adding show: %w", err)
	}

	return c.JSON(http.StatusCreated, show)
}
