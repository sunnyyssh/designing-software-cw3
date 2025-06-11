package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/model"
	"github.com/sunnyyssh/designing-software-cw3/shared/errs"
)

type Repository interface {
	Order() OrderRepository
	Outbox() Outbox
}

type repository struct {
	db pgx.Tx
}

func (r *repository) Order() OrderRepository {
	return &orderRepository{r.db}
}

func (r *repository) Outbox() Outbox {
	return &outbox{r.db}
}

type OrderRepository interface {
	Get(context.Context, uuid.UUID) (*model.Order, error)
	List(context.Context) ([]model.Order, error)
	Create(context.Context, *model.Order) error
	Update(context.Context, *model.Order) error
}

type orderRepository struct {
	db pgx.Tx
}

func (r *orderRepository) Get(ctx context.Context, orderID uuid.UUID) (*model.Order, error) {
	q := `SELECT user_id, description, amount, status FROM orders WHERE id = $1`
	order := &model.Order{
		ID: orderID,
	}

	err := r.db.QueryRow(ctx, q, orderID).Scan(&order.UserID, &order.Description, &order.Amount, &order.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("order with id %s not found", orderID)
		}
		return nil, err
	}
	return order, nil
}

func (r *orderRepository) List(ctx context.Context) ([]model.Order, error) {
	q := `SELECT id, user_id, description, amount, status FROM orders`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]model.Order, 0)

	for rows.Next() {
		var order model.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.Description, &order.Amount, &order.Status)
		if err != nil {
			return nil, err
		}
		res = append(res, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	q := `INSERT INTO orders (id, user_id, description, amount, status) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, q, order.ID, order.UserID, order.Description, order.Amount, order.Status)
	return err
}

func (r *orderRepository) Update(ctx context.Context, order *model.Order) error {
	q := `UPDATE orders SET user_id = $2, description = $3, amount = $4, status = $5 WHERE id = $1 `
	_, err := r.db.Exec(ctx, q, order.ID, order.UserID, order.Description, order.Amount, order.Status)
	return err
}

type Outbox interface {
	Add(context.Context, any) error
}

type outbox struct {
	db pgx.Tx
}

func (o *outbox) Add(ctx context.Context, msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if _, err := o.db.Exec(ctx, `INSERT INTO outbox (message) VALUES $1`, data); err != nil {
		return err
	}
	return nil
}
