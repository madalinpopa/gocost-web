package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/madalinpopa/gocost-web/internal/domain"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type CategoryUseCaseImpl struct {
	uow    domain.UnitOfWork
	logger *slog.Logger
}

func NewCategoryUseCase(uow domain.UnitOfWork, logger *slog.Logger) CategoryUseCaseImpl {
	return CategoryUseCaseImpl{
		uow:    uow,
		logger: logger,
	}
}

func (u CategoryUseCaseImpl) Create(ctx context.Context, req *CreateCategoryRequest) (*CategoryResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	group, err := u.verifyGroupOwnership(ctx, req.UserID, req.GroupID)
	if err != nil {
		return nil, err
	}

	name, err := tracking.NewNameVO(req.Name)
	if err != nil {
		return nil, err
	}

	description, err := tracking.NewDescriptionVO(req.Description)
	if err != nil {
		return nil, err
	}

	startMonth, err := tracking.ParseMonth(req.StartMonth)
	if err != nil {
		return nil, err
	}

	var endMonth tracking.Month
	if req.EndMonth != "" {
		endMonth, err = tracking.ParseMonth(req.EndMonth)
		if err != nil {
			return nil, err
		}
	}

	id, err := identifier.NewID()
	if err != nil {
		return nil, err
	}

	budget, err := money.NewFromFloat(req.Budget, req.Currency)
	if err != nil {
		return nil, err
	}

	category, err := group.CreateCategory(id, name, description, req.IsRecurrent, startMonth, endMonth, budget)
	if err != nil {
		return nil, err
	}

	txUOW, err := u.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}

	if err := txUOW.TrackingRepository().Save(ctx, *group); err != nil {
		_ = txUOW.Rollback()
		return nil, err
	}

	if err := txUOW.Commit(); err != nil {
		_ = txUOW.Rollback()
		return nil, err
	}

	return u.mapToResponse(category), nil
}

func (u CategoryUseCaseImpl) Update(ctx context.Context, req *UpdateCategoryRequest) (*CategoryResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	group, err := u.verifyGroupOwnership(ctx, req.UserID, req.GroupID)
	if err != nil {
		return nil, err
	}

	cID, err := identifier.ParseID(req.ID)
	if err != nil {
		return nil, err
	}

	// Find existing category to check properties
	var existingCategory *tracking.Category
	for _, c := range group.Categories {
		if c.ID == cID {
			existingCategory = c
			break
		}
	}
	if existingCategory == nil {
		return nil, tracking.ErrCategoryNotFound
	}

	name, err := tracking.NewNameVO(req.Name)
	if err != nil {
		return nil, err
	}

	description, err := tracking.NewDescriptionVO(req.Description)
	if err != nil {
		return nil, err
	}

	startMonth, err := tracking.ParseMonth(req.StartMonth)
	if err != nil {
		return nil, err
	}

	var endMonth tracking.Month
	if req.EndMonth != "" {
		endMonth, err = tracking.ParseMonth(req.EndMonth)
		if err != nil {
			return nil, err
		}
	}

	budget, err := money.NewFromFloat(req.Budget, req.Currency)
	if err != nil {
		return nil, err
	}

	// Fork Logic
	shouldFork := false
	var viewMonth tracking.Month
	if req.CurrentMonth != "" {
		viewMonth, err = tracking.ParseMonth(req.CurrentMonth)
		if err == nil && !viewMonth.IsZero() {
			// If viewMonth > StartMonth AND user didn't manually change StartMonth
			if existingCategory.StartMonth.Before(viewMonth) && req.StartMonth == existingCategory.StartMonth.Value() {
				shouldFork = true
			}
		}
	}

	var category *tracking.Category

	if shouldFork && existingCategory.IsRecurrent {
		// 1. Terminate old category
		cutoff := viewMonth.Previous()
		_, err = group.UpdateCategory(existingCategory.ID, existingCategory.Name, existingCategory.Description, existingCategory.IsRecurrent, existingCategory.StartMonth, cutoff, existingCategory.Budget)
		if err != nil {
			return nil, err
		}

		// 2. Create new category
		newID, err := identifier.NewID()
		if err != nil {
			return nil, err
		}

		// New category starts at viewMonth (where the user is)
		category, err = group.CreateCategory(newID, name, description, req.IsRecurrent, viewMonth, endMonth, budget)
		if err != nil {
			return nil, err
		}
	} else {
		// Standard Update
		category, err = group.UpdateCategory(cID, name, description, req.IsRecurrent, startMonth, endMonth, budget)
		if err != nil {
			return nil, err
		}
	}

	repo := u.uow.TrackingRepository()
	if err := repo.Save(ctx, *group); err != nil {
		return nil, err
	}

	if shouldFork && existingCategory.IsRecurrent {
		expenseRepo := u.uow.ExpenseRepository()
		if err := expenseRepo.ReassignCategoryFromMonth(ctx, group.UserID, existingCategory.ID, category.ID, viewMonth.Value()); err != nil {
			return nil, err
		}
	}

	return u.mapToResponse(category), nil
}

func (u CategoryUseCaseImpl) Delete(ctx context.Context, userID string, groupID string, id string) error {
	group, err := u.verifyGroupOwnership(ctx, userID, groupID)
	if err != nil {
		return err
	}

	cID, err := identifier.ParseID(id)
	if err != nil {
		return err
	}

	repo := u.uow.TrackingRepository()
	if err := group.RemoveCategory(cID); err != nil {
		return err
	}

	return repo.DeleteCategory(ctx, cID)
}

func (u CategoryUseCaseImpl) Get(ctx context.Context, userID string, groupID string, id string) (*CategoryResponse, error) {
	group, err := u.verifyGroupOwnership(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}

	cID, err := identifier.ParseID(id)
	if err != nil {
		return nil, err
	}

	for _, c := range group.Categories {
		if c.ID == cID {
			return u.mapToResponse(c), nil
		}
	}

	return nil, tracking.ErrCategoryNotFound
}

func (u CategoryUseCaseImpl) List(ctx context.Context, userID string, groupID string) ([]CategoryResponse, error) {
	group, err := u.verifyGroupOwnership(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}

	responses := make([]CategoryResponse, len(group.Categories))
	for i, c := range group.Categories {
		responses[i] = *u.mapToResponse(c)
	}

	return responses, nil
}

func (u CategoryUseCaseImpl) verifyGroupOwnership(ctx context.Context, userID string, groupID string) (*tracking.Group, error) {
	gID, err := identifier.ParseID(groupID)
	if err != nil {
		return nil, err
	}

	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	repo := u.uow.TrackingRepository()
	group, err := repo.FindByID(ctx, gID)
	if err != nil {
		return nil, err
	}

	if group.UserID != uID {
		return nil, tracking.ErrGroupNotFound // Return NotFound to prevent ID enumeration/probing
	}

	return &group, nil
}

func (u CategoryUseCaseImpl) mapToResponse(c *tracking.Category) *CategoryResponse {
	return &CategoryResponse{
		ID:          c.ID.String(),
		Name:        c.Name.Value(),
		Description: c.Description.Value(),
		IsRecurrent: c.IsRecurrent,
		StartMonth:  c.StartMonth.Value(),
		EndMonth:    c.EndMonth.Value(),
		Budget:      c.Budget.Amount(),
	}
}
