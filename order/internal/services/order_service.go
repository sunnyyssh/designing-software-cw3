package services

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/model"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/storage"
)

type OrderService struct {
	storage *storage.Storage
}

func NewOrderService(storage *storage.Storage) *OrderService {
	return &OrderService{storage}
}

func (s *OrderService) GetOrder(ctx context.Context, orderID uuid.UUID) (_ *model.Order, err error) {
	repo, endTx, err := s.storage.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer endTx(ctx, &err)

	return repo.Order().Get(ctx, orderID)
}

func (s *OrderService) ListOrders(ctx context.Context) (_ []model.Order, err error) {
	repo, endTx, err := s.storage.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer endTx(ctx, &err)

	return repo.Order().List(ctx)
}

func (s *OrderService) CreateOrder(
	ctx context.Context, userID uuid.UUID, amount int64, description string,
) (_ *model.Order, err error) {
	repo, endTx, err := s.storage.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer endTx(ctx, &err)

	order := &model.Order{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      userID,
		Description: description,
		Amount:      amount,
		Status:      model.StatusNew,
	}

	if err := repo.Order().Create(ctx, order); err != nil {
		return nil, err
	}

	err = repo.Outbox().Add(ctx, model.OrderMessage{
		ID:     order.ID,
		UserID: order.UserID,
		Amount: order.Amount,
	})
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) SetOrderStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) (err error) {
	repo, endTx, err := s.storage.Begin(ctx)
	if err != nil {
		return err
	}
	defer endTx(ctx, &err)

	order, err := repo.Order().Get(ctx, id)
	if err != nil {
		return err
	}

	order.Status = status

	if err = repo.Order().Update(ctx, order); err != nil {
		return err
	}

	return nil
}
