package txcontext

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type key int

func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, key(0), tx)
}

func FromContext(ctx context.Context) (pgx.Tx, bool) {
	val := ctx.Value(key(0))
	if val == nil {
		return nil, false
	}
	return val.(pgx.Tx), true
}
