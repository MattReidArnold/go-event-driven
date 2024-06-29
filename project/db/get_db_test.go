package db_test

import (
	"context"
	"os"
	"sync"
	"tickets/db"
	"tickets/message/outbox"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

var _db *sqlx.DB
var getDbOnce sync.Once

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		panic(err)
	}
}

func getDb() *sqlx.DB {
	getDbOnce.Do(func() {
		watermillLogger := log.NewWatermill(log.FromContext(context.Background()))

		var err error
		_db, err = sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
		if err != nil {
			panic(err)
		}
		err = db.InitializeSchema(_db)
		if err != nil {
			panic(err)
		}
		outbox.NewPostgresSubscriber(_db.DB, watermillLogger)

	})
	return _db
}
