package inbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
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

			if cnt == 0 {
				w.logger.DebugContext(ctx, "serving inbox", "cnt", cnt)
			} else {
				w.logger.InfoContext(ctx, "serving inbox", "cnt", cnt)
			}
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
	if err != nil {
		return 0, err
	}
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

	if len(messages) == 0 {
		return 0, nil
	}

	events := make([]json.RawMessage, 0, len(messages))
	for _, msg := range messages {
		events = append(events, msg.Message)
	}

	if err := w.handler(ctx, tx, events...); err != nil {
		return 0, err
	}

	ids := make([]int, 0, len(messages))
	for _, msg := range messages {
		ids = append(ids, msg.ID)
	}

	idsQuery, idsArgs := sqlArgs(ids, 1)

	if _, err := tx.Exec(ctx, fmt.Sprintf(`DELETE FROM inbox WHERE id IN (%s)`, idsQuery), idsArgs...); err != nil {
		return 0, err
	}

	return len(messages), nil
}

func sqlArgs[T any](args []T, start int) (string, []any) {
	var q strings.Builder
	anyArgs := make([]any, 0, len(args))

	for i, arg := range args {
		anyArgs = append(anyArgs, arg)

		q.WriteString("$")
		q.WriteString(strconv.Itoa(start + i))
		if i != len(args)-1 {
			q.WriteString(", ")
		}
	}

	return q.String(), anyArgs
}
