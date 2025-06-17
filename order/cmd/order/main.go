package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/rabbit"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/rest"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/services"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/storage"
	"github.com/sunnyyssh/designing-software-cw3/shared/auth"
	"github.com/sunnyyssh/designing-software-cw3/shared/httplib"
	"github.com/sunnyyssh/designing-software-cw3/shared/outbox"
)

const (
	QueueOrderToPayment = "order_to_payment"
	QueuePaymentToOrder = "payment_to_order"
)

func run(ctx context.Context, logger *slog.Logger) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	r := httplib.NewServer()

	db, err := pgxpool.New(ctx, os.Getenv("PG_CONN_STRING"))
	if err != nil {
		return err
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		return err
	}

	for _, migration := range migrations {
		_, err = db.Exec(ctx, migration)
		if err != nil {
			return err
		}
	}

	rabbitMQConnString := os.Getenv("RABBITMQ_CONN_STRING")

	var conn *amqp091.Connection

	for i := 0; i < 10; i++ {
		conn, err = amqp091.Dial(rabbitMQConnString)
		if err == nil {
			break
		}
		logger.Warn("Failed to connect to RabbitMQ, retrying in 2 seconds...", "error", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	logger.Info("Successfully connected to RabbitMQ")

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	qSend, err := ch.QueueDeclare(
		QueueOrderToPayment, // name
		true,                // durable (очередь переживет перезапуск брокера)
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return err
	}

	queuePublisher := rabbit.NewPublisher(ch, &qSend)

	qReceive, err := ch.QueueDeclare(
		QueuePaymentToOrder, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return err
	}

	queueListener := rabbit.NewListener(db, ch, &qReceive, logger)
	go queueListener.Run(ctx)

	st := storage.NewStorage(db)

	service := services.NewOrderService(st)

	handler := rest.NewOrderHandler(service)

	outboxWorker := outbox.NewWorker(
		db,
		queuePublisher,
		&outbox.Config{
			Period:    1 * time.Second,
			BatchSize: 1,
		},
		logger,
	)

	r.Mount("/order").
		Use(auth.MiddlewareUserID).
		GET("", handler.GetOrder).
		GET("/all", handler.ListOrders).
		POST("", handler.CreateOrder)

	go func() {
		if err := outboxWorker.Run(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				logger.Info("outbox worker gracefully stopped")
			} else {
				logger.Error("outbox worker failed and stopped", "error", err)
			}
		}
	}()

	if err := http.ListenAndServe(":8080", r); err != nil {
		return err
	}
	return nil
}

func main() {
	logger := slog.Default()
	ctx := context.Background()

	if err := run(ctx, logger); err != nil {
		logger.ErrorContext(ctx, "run failed", "error", err)
	}
}

var migrations = []string{
	`CREATE TYPE order_status AS ENUM ('new', 'finished', 'cancelled');`,

	`CREATE TABLE orders (
		id UUID PRIMARY KEY,
		user_id UUID NOT NULL,
		description TEXT,
		amount BIGINT NOT NULL,
		status order_status NOT NULL DEFAULT 'new'
	);`,

	`CREATE TABLE outbox (
		id SERIAL,
		message JSONB NOT NULL
	)`,

	`CREATE TABLE inbox (
		id SERIAL,
		message JSONB NOT NULL
	)`,
}
