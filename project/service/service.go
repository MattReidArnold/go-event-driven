package service

import (
	"context"
	"fmt"
	"net/http"

	"tickets/db"
	ticketsHTTP "tickets/http"
	ticketsMessage "tickets/message"
	"tickets/message/command"
	"tickets/message/event"
	"tickets/message/outbox"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	db              *sqlx.DB
	watermillRouter *watermillMessage.Router
	echoRouter      *echo.Echo
}

func New(
	dbConn *sqlx.DB,
	rdb *redis.Client,
	receiptsService event.ReceiptService,
	spreadsheetsAPI event.SpreadsheetsAPI,
	filesService event.FilesService,
	deadNationService event.DeadNationService,
	paymentRefunder command.PaymentRefunder,
	receiptVoider command.ReceiptVoider,
) Service {
	watermillLogger := log.NewWatermill(log.FromContext(context.Background()))

	ticketsRepo := db.NewTicketsRepo(dbConn)
	showsRepo := db.NewShowsRepo(dbConn)
	bookingsRepo := db.NewBookingsRepository(dbConn)

	var redisPublisher watermillMessage.Publisher
	redisPublisher = ticketsMessage.NewRedisPublisher(rdb, watermillLogger)
	redisPublisher = log.CorrelationPublisherDecorator{Publisher: redisPublisher}

	eventBus := event.NewEventBus(redisPublisher)
	commandBus := command.NewCommandBus(redisPublisher, command.NewBusConfig(watermillLogger))

	eventHandler := event.NewEventHandler(
		ticketsRepo,
		showsRepo,
		receiptsService,
		spreadsheetsAPI,
		deadNationService,
		filesService,
		eventBus,
	)
	commandHandler := command.NewHandler(paymentRefunder, receiptVoider)

	postgresSubscriber := outbox.NewPostgresSubscriber(dbConn.DB, watermillLogger)
	eventProcessorConfig := event.NewProcessorConfig(rdb, watermillLogger)
	commandProcessorConfig := command.NewProcessorConfig(rdb, watermillLogger)

	router := ticketsMessage.NewRouter(
		postgresSubscriber,
		redisPublisher,
		eventProcessorConfig,
		eventHandler,
		commandProcessorConfig,
		commandHandler,
		watermillLogger,
	)

	e := ticketsHTTP.NewHttpRouter(
		commandBus,
		eventBus,
		spreadsheetsAPI,
		ticketsRepo,
		showsRepo,
		bookingsRepo,
	)

	return Service{
		db:              dbConn,
		watermillRouter: router,
		echoRouter:      e,
	}
}

func (s Service) Run(ctx context.Context) error {
	if err := db.InitializeSchema(s.db); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.watermillRouter.Run(ctx)
	})

	g.Go(func() error {
		<-s.watermillRouter.Running()

		err := s.echoRouter.Start(":8080")
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		return s.echoRouter.Shutdown(context.Background())
	})

	return g.Wait()
}
