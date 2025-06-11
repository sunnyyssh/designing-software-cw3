package model

import "github.com/gofrs/uuid"

type Account struct {
	UserID uuid.UUID `json:"user_id"`
	Amount int64     `json:"amount"`
}
