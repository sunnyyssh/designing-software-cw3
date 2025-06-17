package handlers

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/model"
	"github.com/sunnyyssh/designing-software-cw3/shared/inbox"
	"github.com/sunnyyssh/designing-software-cw3/shared/txcontext"
)

type PaymentService interface {
	ServeOrder(context.Context, *model.OrderMessage) error
}

func NewInboxHandler(service PaymentService) inbox.HandlerFunc {
	return func(ctx context.Context, tx pgx.Tx, messages ...json.RawMessage) error {
		ctx = txcontext.WithTx(ctx, tx)

		for _, msg := range messages {
			var orderMsg model.OrderMessage
			if err := json.Unmarshal([]byte(msg), &orderMsg); err != nil {
				return err
			}

			if err := service.ServeOrder(ctx, &orderMsg); err != nil {
				return err
			}
		}

		return nil
	}
}
