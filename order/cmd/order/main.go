package main

import (
	"context"
	"errors"
	"fmt"
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

	for range 10 {
		if err = db.Ping(ctx); err == nil {
			break
		}
		logger.Warn("Failed to connect to PostgreSQL, retrying in 2 seconds...", "error", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		_, err = db.Exec(ctx, migration)
		if err != nil {
			return fmt.Errorf("failed to apply migration \"%s\": %w", migration, err)
		}
	}

	rabbitMQConnString := os.Getenv("RABBITMQ_CONN_STRING")

	var conn *amqp091.Connection

	for range 10 {
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

	st := storage.NewStorage(db)

	service := services.NewOrderService(st)

	queueListener := rabbit.NewListener(service, ch, &qReceive, logger)
	go func() {
		if err := queueListener.Run(ctx); err != nil {
			logger.ErrorContext(ctx, "listening queue failed", "error", err)
		}
	}()

	outboxWorker := outbox.NewWorker(
		db,
		queuePublisher,
		&outbox.Config{
			Period:    1 * time.Second,
			BatchSize: 1,
		},
		logger,
	)

	handler := rest.NewOrderHandler(service)

	r.Mount("/order").
		GET("/{orderId}", handler.GetOrder).
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
		os.Exit(1)
	}
}

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS orders (
		id UUID PRIMARY KEY,
		user_id UUID NOT NULL,
		description TEXT,
		amount BIGINT NOT NULL,
		status VARCHAR(255) NOT NULL DEFAULT 'new'
	);`,

	`CREATE TABLE IF NOT EXISTS outbox (
		id SERIAL,
		message JSONB NOT NULL
	)`,

	`CREATE TABLE IF NOT EXISTS outbox (
		id SERIAL,
		message JSONB NOT NULL
	)`,
}
