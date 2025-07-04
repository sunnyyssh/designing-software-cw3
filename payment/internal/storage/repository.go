package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/model"
	"github.com/sunnyyssh/designing-software-cw3/shared/errs"
)

type Repository interface {
	Account() AccountRepository
	Outbox() Outbox
}

type repository struct {
	db pgx.Tx
}

func (r *repository) Account() AccountRepository {
	return &accountRepository{r.db}
}

func (r *repository) Outbox() Outbox {
	return &outbox{r.db}
}

type AccountRepository interface {
	GetAccount(ctx context.Context, userID uuid.UUID) (*model.Account, error)
	CreateAccount(context.Context, *model.Account) error
	UpdateAccount(context.Context, *model.Account) error
}

type accountRepository struct {
	db pgx.Tx
}

func (r *accountRepository) GetAccount(ctx context.Context, userID uuid.UUID) (*model.Account, error) {
	account := &model.Account{
		UserID: userID,
	}

	row := r.db.QueryRow(ctx, `SELECT amount FROM accounts WHERE user_id = $1`, userID)
	if err := row.Scan(&account.Amount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("account with user_id %s not found", userID)
		}
		return nil, err
	}

	return account, nil
}

func (r *accountRepository) CreateAccount(ctx context.Context, account *model.Account) error {
	_, err := r.db.Exec(ctx, `INSERT INTO accounts (user_id, amount) VALUES ($1, $2)`, account.UserID, account.Amount)
	return err
}

func (r *accountRepository) UpdateAccount(ctx context.Context, account *model.Account) error {
	_, err := r.db.Exec(ctx, `UPDATE accounts SET amount = $1 WHERE user_id = $2`, account.Amount, account.UserID)
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

	if _, err := o.db.Exec(ctx, `INSERT INTO outbox (message) VALUES ($1)`, data); err != nil {
		return err
	}
	return nil
}
