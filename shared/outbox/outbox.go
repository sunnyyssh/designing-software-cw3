package outbox

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	ID      int
	Message json.RawMessage
}

type EventPublisher interface {
	Publish(context.Context, ...any) error
}

type Config struct {
	Period    time.Duration
	BatchSize int
}

type Worker struct {
	db        *pgxpool.Pool
	publisher EventPublisher
	cfg       *Config
	logger    *slog.Logger
}

func NewWorker(db *pgxpool.Pool, publisher EventPublisher, cfg *Config, logger *slog.Logger) *Worker {
	return &Worker{
		db:        db,
		publisher: publisher,
		cfg:       cfg,
		logger:    logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-time.Tick(w.cfg.Period):
			cnt, err := w.singleRun(ctx)
			if err != nil {
				w.logger.ErrorContext(ctx, "serving outbox failed", "error", err)
				continue
			}

			w.logger.InfoContext(ctx, "serving outbox", "cnt", cnt)
		}
	}
}

func (w *Worker) singleRun(ctx context.Context) (cnt int, err error) {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() error {
		if r := recover(); r != nil {
			return tx.Rollback(ctx)
		}
		if err != nil {
			return tx.Rollback(ctx)
		}
		return tx.Commit(ctx)
	}()

	rows, err := tx.Query(ctx, `SELECT id, message FROM outbox LIMIT $1`, w.cfg.BatchSize)
	defer rows.Close()

	messages := make([]Message, 0, w.cfg.BatchSize)

	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Message); err != nil {
			return 0, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}

	events := make([]any, 0, len(messages))
	for _, msg := range messages {
		events = append(events, msg.Message)
	}

	if err := w.publisher.Publish(ctx, events...); err != nil {
		return 0, err
	}

	ids := make([]int, 0, len(messages))

	if _, err := tx.Exec(ctx, `DELETE FROM outbox WHERE id IN $1`, ids); err != nil {
		return 0, err
	}

	return len(messages), nil
}
