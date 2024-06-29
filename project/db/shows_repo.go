package db

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type showsRepo struct {
	db *sqlx.DB
}

func NewShowsRepo(dbConn *sqlx.DB) *showsRepo {
	return &showsRepo{
		db: dbConn,
	}
}

func (r *showsRepo) AddShow(ctx context.Context, show entities.Show) error {
	res, err := r.db.NamedExecContext(ctx, `
		INSERT INTO 
			shows (show_id, dead_nation_id, number_of_tickets, start_time, title, venue) 
		VALUES (:show_id, :dead_nation_id, :number_of_tickets, :start_time, :title, :venue )
		ON CONFLICT (show_id) DO NOTHING`,
		show,
	)
	if err != nil {
		return fmt.Errorf("inserting show: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 1 {
		logrus.Info("show inserted ", show.ShowID)
	}

	return nil
}

func (r *showsRepo) FindByID(ctx context.Context, showID string) (entities.Show, error) {
	var show entities.Show
	err := r.db.GetContext(ctx, &show, `
		SELECT 
			show_id, dead_nation_id, number_of_tickets, start_time, title, venue 
		FROM 
			shows 
		WHERE show_id = $1
		`, showID)
	if err != nil {
		return entities.Show{}, fmt.Errorf("selecting show by id %s: %w", showID, err)
	}
	return show, nil
}
