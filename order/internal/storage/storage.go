package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sunnyyssh/designing-software-cw3/shared/txcontext"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(db *pgxpool.Pool) *Storage { return &Storage{db} }

func (s *Storage) Begin(ctx context.Context) (Repository, func(context.Context, *error) error, error) {
	if ctxTx, ok := txcontext.FromContext(ctx); ok {
		emptyEndTx := func(ctx context.Context, err *error) error { return nil }
		return &repository{ctxTx}, emptyEndTx, nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Should be used only in `defer` block because it has recover() call
	endTx := func(ctx context.Context, err *error) error {
		if p := recover(); p != nil {
			if *err != nil {
				*err = fmt.Errorf(`[PANIC] panic recovered: %+v. (error overwritten: "%w")`, p, *err)
			} else {
				*err = fmt.Errorf(`[PANIC] panic recovered: %+v`, p)
			}
		}

		if *err != nil {
			return tx.Rollback(ctx)
		} else {
			return tx.Commit(ctx)
		}
	}

	return &repository{tx}, endTx, nil
}
