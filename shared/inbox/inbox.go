package inbox

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HandlerFunc func(context.Context, pgx.Tx, ...json.RawMessage) error

type Message struct {
	ID      int
	Message json.RawMessage
}

type Config struct {
	Period    time.Duration
	BatchSize int
}

type Worker struct {
	db      *pgxpool.Pool
	handler HandlerFunc
	cfg     *Config
	logger  *slog.Logger
}

func NewWorker(
	db *pgxpool.Pool,
	handler HandlerFunc,
	cfg *Config,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		db:      db,
		handler: handler,
		cfg:     cfg,
		logger:  logger,
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
				w.logger.ErrorContext(ctx, "serving inbox failed", "error", err)
				continue
			}

			w.logger.InfoContext(ctx, "serving inbox", "cnt", cnt)
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

	rows, err := tx.Query(ctx, `SELECT id, message FROM inbox LIMIT $1`, w.cfg.BatchSize)
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

	events := make([]json.RawMessage, 0, len(messages))
	for _, msg := range messages {
		events = append(events, msg.Message)
	}

	if err := w.handler(ctx, tx, events...); err != nil {
		return 0, err
	}

	ids := make([]int, 0, len(messages))

	if _, err := tx.Exec(ctx, `DELETE FROM inbox WHERE id IN $1`, ids); err != nil {
		return 0, err
	}

	return len(messages), nil
}
