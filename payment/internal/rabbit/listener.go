package rabbit

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
)

type Listener struct {
	db     *pgxpool.Pool
	ch     *amqp091.Channel
	q      *amqp091.Queue
	logger *slog.Logger
}

func NewListener(db *pgxpool.Pool, ch *amqp091.Channel, q *amqp091.Queue, logger *slog.Logger) *Listener {
	return &Listener{
		db:     db,
		ch:     ch,
		q:      q,
		logger: logger,
	}
}

func (l *Listener) Run(ctx context.Context) error {
	logger := l.logger.With("queue", l.q.Name)

	msgs, err := l.ch.Consume(
		l.q.Name, // queue
		"",       // consumer
		true,     // auto-ack (для простоты примера, в проде лучше false)
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "consuming queue")

LOOP:
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				logger.Info("stopping")
				break LOOP
			}

			logger.Info("message received", "timestamp", msg.Timestamp)
			if err = l.handleMessage(ctx, msg, logger); err != nil {
				return err
			}

		case <-ctx.Done():
			break LOOP
		}
	}

	return nil
}

func (l *Listener) handleMessage(ctx context.Context, msg amqp091.Delivery, logger *slog.Logger) (err error) {
	tx, err := l.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `INSERT INTO inbox (message) VALUES ($1)`, msg.Body)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	logger.Info("message appended to inbox table")
	return nil
}
