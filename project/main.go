package main

import (
	"context"
	"os"
	"os/signal"
	"tickets/clients"
	"tickets/message"
	"tickets/service"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	log.Init(logrus.InfoLevel)

	godotenv.Load()

	rdb := message.NewRedisClient(os.Getenv("REDIS_ADDR"))
	defer rdb.Close()

	c, err := clients.NewClients(os.Getenv("GATEWAY_ADDR"))
	if err != nil {
		panic(err)
	}
	db, err := sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	receiptsClient := clients.NewReceiptsClient(c)
	spreadsheetsClient := clients.NewSpreadsheetsClient(c)
	filesClient := clients.NewFilesClient(c)
	deadNationClient := clients.NewDeadNationClient(c)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	app := service.New(db, rdb, receiptsClient, spreadsheetsClient, filesClient, deadNationClient)

	err = app.Run(ctx)
	if err != nil {
		panic(err)
	}
}
