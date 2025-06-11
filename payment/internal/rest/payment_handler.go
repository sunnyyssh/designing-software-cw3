package rest

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/model"
	"github.com/sunnyyssh/designing-software-cw3/shared/auth"
	"github.com/sunnyyssh/designing-software-cw3/shared/errs"
	"github.com/sunnyyssh/designing-software-cw3/shared/httplib"
)

type PaymentService interface {
	GetAccount(ctx context.Context, userID uuid.UUID) (*model.Account, error)
	CreateAccount(ctx context.Context, userID uuid.UUID) (*model.Account, error)
	ReplenishAccount(ctx context.Context, userID uuid.UUID, amount int64) (*model.Account, error)
}

type PaymentHandler struct {
	service PaymentService
}

func NewPaymentHandler(service PaymentService) *PaymentHandler {
	return &PaymentHandler{service}
}

func (h *PaymentHandler) GetAccount(req *http.Request) (any, error) {
	ctx := req.Context()

	userID := auth.MustUserIDFromContext(ctx)

	return h.service.GetAccount(ctx, userID)
}

func (h *PaymentHandler) CreateAccount(req *http.Request) (any, error) {
	ctx := req.Context()

	userID := auth.MustUserIDFromContext(ctx)

	return h.service.CreateAccount(ctx, userID)
}

func (h *PaymentHandler) ReplenishAccount(req *http.Request) (any, error) {
	ctx := req.Context()

	userID := auth.MustUserIDFromContext(ctx)

	type Request struct {
		Amount int64 `json:"amount"`
	}
	reqBody, err := httplib.UnmarshalBody[Request](req)
	if err != nil {
		return nil, errs.BadRequest("invalid format of body")
	}

	return h.service.ReplenishAccount(ctx, userID, reqBody.Amount)
}
