package rest

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/model"
	"github.com/sunnyyssh/designing-software-cw3/shared/errs"
)

type OrderService interface {
	GetOrder(ctx context.Context, orderID uuid.UUID) (*model.Order, error)
	ListOrders(ctx context.Context) ([]model.Order, error)
	CreateOrder(ctx context.Context, userID uuid.UUID, amount int64) (*model.Order, error)
}

type OrderHandler struct {
	service OrderService
}

func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service}
}

func (h *OrderHandler) GetOrder(req *http.Request) (any, error) {
	orderID, err := uuid.FromString(req.PathValue("orderId"))
	if err != nil {
		return nil, errs.BadRequest("orderId UUID path value must be specified: %s", err)
	}

	return h.service.GetOrder(req.Context(), orderID)
}

func (h *OrderHandler) ListOrders(req *http.Request) (any, error) {
	return h.service.ListOrders(req.Context())
}
