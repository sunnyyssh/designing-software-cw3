package services

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/model"
	"github.com/sunnyyssh/designing-software-cw3/payment/internal/storage"
	"github.com/sunnyyssh/designing-software-cw3/shared/errs"
)

type PaymentService struct {
	storage *storage.Storage
}

func NewPaymentService(storage *storage.Storage) *PaymentService {
	return &PaymentService{storage}
}

func (s *PaymentService) GetAccount(ctx context.Context, userID uuid.UUID) (_ *model.Account, err error) {
	repo, endTx, err := s.storage.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer endTx(ctx, &err)

	return repo.Account().GetAccount(ctx, userID)
}

func (s *PaymentService) CreateAccount(ctx context.Context, userID uuid.UUID) (_ *model.Account, err error) {
	repo, endTx, err := s.storage.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer endTx(ctx, &err)

	acc := &model.Account{
		UserID: userID,
		Amount: 0,
	}

	if err := repo.Account().CreateAccount(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *PaymentService) ReplenishAccount(ctx context.Context, userID uuid.UUID, amount int64) (_ *model.Account, err error) {
	repo, endTx, err := s.storage.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer endTx(ctx, &err)

	acc, err := repo.Account().GetAccount(ctx, userID)
	if err != nil {
		return nil, err
	}

	if acc.Amount+amount < 0 {
		return nil, errs.BadRequest("not enough money on the account")
	}
	acc.Amount += amount

	if err := repo.Account().UpdateAccount(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}
