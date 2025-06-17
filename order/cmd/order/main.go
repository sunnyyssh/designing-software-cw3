package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/rest"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/services"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/storage"
	"github.com/sunnyyssh/designing-software-cw3/shared/auth"
	"github.com/sunnyyssh/designing-software-cw3/shared/httplib"
	"github.com/sunnyyssh/designing-software-cw3/shared/outbox"
)

func run() error {
	logger := slog.Default()
	ctx := context.Background()

	r := httplib.NewServer()

	db, err := pgxpool.New(ctx, os.Getenv("PG_CONN_STRING"))
	if err != nil {
		return err
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		return err
	}

	st := storage.NewStorage(db)

	service := services.NewOrderService(st)

	handler := rest.NewOrderHandler(service)

	outboxWorker := outbox.NewWorker(
		db,
		nil, // TODO:
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
	if err := run(); err != nil {
		panic(err)
	}
}
