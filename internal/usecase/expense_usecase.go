package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/madalinpopa/gocost-web/internal/domain"
	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type ExpenseUseCaseImpl struct {
	uow    domain.UnitOfWork
	logger *slog.Logger
}

func NewExpenseUseCase(uow domain.UnitOfWork, logger *slog.Logger) ExpenseUseCaseImpl {
	return ExpenseUseCaseImpl{
		uow:    uow,
		logger: logger,
	}
}

func (u ExpenseUseCaseImpl) Create(ctx context.Context, req *CreateExpenseRequest) (*ExpenseResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	uID, err := identifier.ParseID(req.UserID)
	if err != nil {
		return nil, err
	}

	catID, err := identifier.ParseID(req.CategoryID)
	if err != nil {
		return nil, err
	}

	// Verify category belongs to user
	groupRepo := u.uow.TrackingRepository()
	group, err := groupRepo.FindGroupByCategoryID(ctx, catID)
	if err != nil {
		return nil, err
	}
	if group.UserID != uID {
		return nil, errors.New("unauthorized")
	}

	amount, err := money.NewFromFloat(req.Amount, req.Currency)
	if err != nil {
		return nil, err
	}

	description, err := expense.NewExpenseDescriptionVO(req.Description)
	if err != nil {
		return nil, err
	}

	payment, err := expense.NewPaymentStatus(req.IsPaid, req.PaidAt)
	if err != nil {
		return nil, err
	}

	id, err := identifier.NewID()
	if err != nil {
		return nil, err
	}

	exp, err := expense.NewExpense(id, catID, amount, description, req.SpentAt, payment)
	if err != nil {
		return nil, err
	}

	if err := u.uow.ExpenseRepository().Save(ctx, *exp); err != nil {
		return nil, err
	}

	return u.mapToResponse(exp), nil
}

func (u ExpenseUseCaseImpl) Update(ctx context.Context, req *UpdateExpenseRequest) (*ExpenseResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	uID, err := identifier.ParseID(req.UserID)
	if err != nil {
		return nil, err
	}

	expID, err := identifier.ParseID(req.ID)
	if err != nil {
		return nil, err
	}

	expenseRepo := u.uow.ExpenseRepository()
	exp, err := expenseRepo.FindByID(ctx, expID)
	if err != nil {
		return nil, err
	}

	// Verify existing expense ownership
	groupRepo := u.uow.TrackingRepository()
	group, err := groupRepo.FindGroupByCategoryID(ctx, exp.CategoryID)
	if err != nil {
		return nil, err
	}
	if group.UserID != uID {
		return nil, errors.New("unauthorized")
	}

	// If category changed, verify new category
	if req.CategoryID != exp.CategoryID.String() {
		newCatID, err := identifier.ParseID(req.CategoryID)
		if err != nil {
			return nil, err
		}
		newGroup, err := groupRepo.FindGroupByCategoryID(ctx, newCatID)
		if err != nil {
			return nil, err
		}
		if newGroup.UserID != uID {
			return nil, errors.New("unauthorized")
		}
		exp.CategoryID = newCatID
	}

	amount, err := money.NewFromFloat(req.Amount, req.Currency)
	if err != nil {
		return nil, err
	}

	description, err := expense.NewExpenseDescriptionVO(req.Description)
	if err != nil {
		return nil, err
	}

	payment, err := expense.NewPaymentStatus(req.IsPaid, req.PaidAt)
	if err != nil {
		return nil, err
	}

	exp.Amount = amount
	exp.Description = description
	exp.SpentAt = req.SpentAt
	exp.Payment = payment

	if err := expenseRepo.Save(ctx, exp); err != nil {
		return nil, err
	}

	return u.mapToResponse(&exp), nil
}

func (u ExpenseUseCaseImpl) Delete(ctx context.Context, userID string, id string) error {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return err
	}

	expID, err := identifier.ParseID(id)
	if err != nil {
		return err
	}

	expenseRepo := u.uow.ExpenseRepository()
	exp, err := expenseRepo.FindByID(ctx, expID)
	if err != nil {
		return err
	}

	groupRepo := u.uow.TrackingRepository()
	group, err := groupRepo.FindGroupByCategoryID(ctx, exp.CategoryID)
	if err != nil {
		return err
	}
	if group.UserID != uID {
		return errors.New("unauthorized")
	}

	return expenseRepo.Delete(ctx, expID)
}

func (u ExpenseUseCaseImpl) Get(ctx context.Context, userID string, id string) (*ExpenseResponse, error) {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	expID, err := identifier.ParseID(id)
	if err != nil {
		return nil, err
	}

	expenseRepo := u.uow.ExpenseRepository()
	exp, err := expenseRepo.FindByID(ctx, expID)
	if err != nil {
		return nil, err
	}

	groupRepo := u.uow.TrackingRepository()
	group, err := groupRepo.FindGroupByCategoryID(ctx, exp.CategoryID)
	if err != nil {
		return nil, err
	}
	if group.UserID != uID {
		return nil, errors.New("unauthorized")
	}

	return u.mapToResponse(&exp), nil
}

func (u ExpenseUseCaseImpl) List(ctx context.Context, userID string) ([]*ExpenseResponse, error) {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	expenseRepo := u.uow.ExpenseRepository()
	expenses, err := expenseRepo.FindByUserID(ctx, uID)
	if err != nil {
		return nil, err
	}

	responses := make([]*ExpenseResponse, len(expenses))
	for i, e := range expenses {
		responses[i] = u.mapToResponse(&e)
	}

	return responses, nil
}

func (u ExpenseUseCaseImpl) ListByMonth(ctx context.Context, userID string, month string) ([]*ExpenseResponse, error) {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	expenseRepo := u.uow.ExpenseRepository()
	expenses, err := expenseRepo.FindByUserIDAndMonth(ctx, uID, month)
	if err != nil {
		return nil, err
	}

	responses := make([]*ExpenseResponse, len(expenses))
	for i, e := range expenses {
		responses[i] = u.mapToResponse(&e)
	}

	return responses, nil
}

func (u ExpenseUseCaseImpl) Total(ctx context.Context, userID string, month string) (float64, error) {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return 0, err
	}

	totalMoney, err := u.uow.ExpenseRepository().Total(ctx, uID, month)
	if err != nil {
		return 0, err
	}
	return totalMoney.Amount(), nil
}

func (u ExpenseUseCaseImpl) mapToResponse(e *expense.Expense) *ExpenseResponse {
	return &ExpenseResponse{
		ID:          e.ID.String(),
		CategoryID:  e.CategoryID.String(),
		AmountCents: e.Amount.Cents(),
		Currency:    e.Amount.Currency(),
		Description: e.Description.Value(),
		SpentAt:     e.SpentAt,
		IsPaid:      e.Payment.IsPaid(),
		PaidAt:      e.Payment.PaidAt(),
	}
}
