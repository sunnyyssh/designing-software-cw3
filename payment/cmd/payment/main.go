package main

import (
	"context"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/rest"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/services"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/storage"
	"github.com/sunnyyssh/designing-software-cw3/shared/auth"
	"github.com/sunnyyssh/designing-software-cw3/shared/httplib"
)

func run() error {
	ctx := context.Background()

	r := httplib.NewServer()

	db, err := pgxpool.New(ctx, os.Getenv("PG_CONN_STRING"))
	if err != nil {
		return err
	}
	if err = db.Ping(ctx); err != nil {
		return err
	}

	st := storage.NewStorage(db)

	service := services.NewPaymentService(st)

	handler := rest.NewPaymentHandler(service)

	r.Mount("/account").
		Use(auth.MiddlewareUserID).
		GET("", handler.GetAccount).
		PUT("", handler.CreateAccount).
		POST("/amount", handler.ReplenishAccount)

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
