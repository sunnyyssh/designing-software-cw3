package model

import "github.com/gofrs/uuid"

type OrderStatus string

const (
	StatusNew       OrderStatus = "new"
	StatusFinished  OrderStatus = "finished"
	StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	Description string      `json:"description"`
	Amount      int64       `json:"amount"`
	Status      OrderStatus `json:"order_status"`
}

type OrderMessage struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Amount int64     `json:"amount"`
}

type OrderServedMessage struct {
	ID     uuid.UUID   `json:"id"`
	Status OrderStatus `json:"status"`
}
