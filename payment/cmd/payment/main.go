package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/handlers"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/rest"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/services"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/storage"
	"github.com/sunnyyssh/designing-software-cw3/shared/auth"
	"github.com/sunnyyssh/designing-software-cw3/shared/httplib"
	"github.com/sunnyyssh/designing-software-cw3/shared/inbox"
)

func run() error {
	ctx := context.Background()
	logger := slog.Default()

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

	service := services.NewPaymentService(st)

	inboxWorker := inbox.NewWorker(
		db,
		handlers.NewInboxHandler(service),
		&inbox.Config{
			Period:    1 * time.Second,
			BatchSize: 1,
		},
		logger,
	)

	handler := rest.NewPaymentHandler(service)

	r.Mount("/account").
		Use(auth.MiddlewareUserID).
		GET("", handler.GetAccount).
		PUT("", handler.CreateAccount).
		POST("/amount", handler.ReplenishAccount)

	go func() {
		if err := inboxWorker.Run(ctx); err != nil {
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
