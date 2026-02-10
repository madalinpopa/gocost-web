package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/madalinpopa/gocost-web/internal/domain"
	"github.com/madalinpopa/gocost-web/internal/domain/income"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type IncomeUseCaseImpl struct {
	uow    domain.UnitOfWork
	logger *slog.Logger
}

func NewIncomeUseCase(uow domain.UnitOfWork, logger *slog.Logger) IncomeUseCaseImpl {
	return IncomeUseCaseImpl{
		uow:    uow,
		logger: logger,
	}
}

func (u IncomeUseCaseImpl) Create(ctx context.Context, req *CreateIncomeRequest) (*IncomeResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	uID, err := identifier.ParseID(req.UserID)
	if err != nil {
		return nil, err
	}

	amount, err := money.NewFromFloat(req.Amount, req.Currency)
	if err != nil {
		return nil, err
	}

	source, err := income.NewSourceVO(req.Source)
	if err != nil {
		return nil, err
	}

	id, err := identifier.NewID()
	if err != nil {
		return nil, err
	}

	inc, err := income.NewIncome(id, uID, amount, source, req.ReceivedAt)
	if err != nil {
		return nil, err
	}

	txUOW, err := u.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}

	if err := txUOW.IncomeRepository().Save(ctx, *inc); err != nil {
		_ = txUOW.Rollback()
		return nil, err
	}

	if err := txUOW.Commit(); err != nil {
		_ = txUOW.Rollback()
		return nil, err
	}

	return &IncomeResponse{
		ID:          inc.ID.String(),
		AmountCents: inc.Amount.Cents(),
		Currency:    inc.Amount.Currency(),
		Source:      inc.Source.Value(),
		ReceivedAt:  inc.ReceivedAt,
	}, nil
}

func (u IncomeUseCaseImpl) Update(ctx context.Context, req *UpdateIncomeRequest) (*IncomeResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	uID, err := identifier.ParseID(req.UserID)
	if err != nil {
		return nil, err
	}

	incID, err := identifier.ParseID(req.ID)
	if err != nil {
		return nil, err
	}

	repo := u.uow.IncomeRepository()
	inc, err := repo.FindByID(ctx, incID)
	if err != nil {
		return nil, err
	}

	if inc.UserID != uID {
		return nil, errors.New("unauthorized")
	}

	amount, err := money.NewFromFloat(req.Amount, req.Currency)
	if err != nil {
		return nil, err
	}

	source, err := income.NewSourceVO(req.Source)
	if err != nil {
		return nil, err
	}

	updatedInc, err := income.NewIncome(inc.ID, inc.UserID, amount, source, req.ReceivedAt)
	if err != nil {
		return nil, err
	}

	txUOW, err := u.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}

	if err := txUOW.IncomeRepository().Save(ctx, *updatedInc); err != nil {
		_ = txUOW.Rollback()
		return nil, err
	}

	if err := txUOW.Commit(); err != nil {
		_ = txUOW.Rollback()
		return nil, err
	}

	return &IncomeResponse{
		ID:          updatedInc.ID.String(),
		AmountCents: updatedInc.Amount.Cents(),
		Currency:    updatedInc.Amount.Currency(),
		Source:      updatedInc.Source.Value(),
		ReceivedAt:  updatedInc.ReceivedAt,
	}, nil
}

func (u IncomeUseCaseImpl) Delete(ctx context.Context, userID string, id string) error {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return err
	}

	incID, err := identifier.ParseID(id)
	if err != nil {
		return err
	}

	repo := u.uow.IncomeRepository()
	inc, err := repo.FindByID(ctx, incID)
	if err != nil {
		return err
	}

	if inc.UserID != uID {
		return errors.New("unauthorized")
	}

	txUOW, err := u.uow.Begin(ctx)
	if err != nil {
		return err
	}

	if err := txUOW.IncomeRepository().Delete(ctx, incID); err != nil {
		_ = txUOW.Rollback()
		return err
	}

	if err := txUOW.Commit(); err != nil {
		_ = txUOW.Rollback()
		return err
	}

	return nil
}

func (u IncomeUseCaseImpl) Get(ctx context.Context, userID string, id string) (*IncomeResponse, error) {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	incID, err := identifier.ParseID(id)
	if err != nil {
		return nil, err
	}

	repo := u.uow.IncomeRepository()
	inc, err := repo.FindByID(ctx, incID)
	if err != nil {
		return nil, err
	}

	if inc.UserID != uID {
		return nil, errors.New("unauthorized")
	}

	return &IncomeResponse{
		ID:          inc.ID.String(),
		AmountCents: inc.Amount.Cents(),
		Currency:    inc.Amount.Currency(),
		Source:      inc.Source.Value(),
		ReceivedAt:  inc.ReceivedAt,
	}, nil
}

func (u IncomeUseCaseImpl) List(ctx context.Context, userID string) ([]*IncomeResponse, error) {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	repo := u.uow.IncomeRepository()
	incomes, err := repo.FindByUserID(ctx, uID)
	if err != nil {
		return nil, err
	}

	responses := make([]*IncomeResponse, len(incomes))
	for i, inc := range incomes {
		responses[i] = &IncomeResponse{
			ID:          inc.ID.String(),
			AmountCents: inc.Amount.Cents(),
			Currency:    inc.Amount.Currency(),
			Source:      inc.Source.Value(),
			ReceivedAt:  inc.ReceivedAt,
		}
	}

	return responses, nil
}

func (u IncomeUseCaseImpl) ListByMonth(ctx context.Context, userID string, month string) ([]*IncomeResponse, error) {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	repo := u.uow.IncomeRepository()
	incomes, err := repo.FindByUserIDAndMonth(ctx, uID, month)
	if err != nil {
		return nil, err
	}

	responses := make([]*IncomeResponse, 0, len(incomes))
	for _, inc := range incomes {
		responses = append(responses, &IncomeResponse{
			ID:          inc.ID.String(),
			AmountCents: inc.Amount.Cents(),
			Currency:    inc.Amount.Currency(),
			Source:      inc.Source.Value(),
			ReceivedAt:  inc.ReceivedAt,
		})
	}

	return responses, nil
}

func (u IncomeUseCaseImpl) Total(ctx context.Context, userID string, month string) (float64, error) {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return 0, err
	}

	repo := u.uow.IncomeRepository()
	total, err := repo.TotalByUserIDAndMonth(ctx, uID, month)
	if err != nil {
		return 0, err
	}

	return total.Amount(), nil
}
