package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/domain/uow"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type CategoryUseCaseImpl struct {
	uow    uow.UnitOfWork
	logger *slog.Logger
}

func NewCategoryUseCase(uow uow.UnitOfWork, logger *slog.Logger) CategoryUseCaseImpl {
	return CategoryUseCaseImpl{
		uow:    uow,
		logger: logger,
	}
}

func (u CategoryUseCaseImpl) Create(ctx context.Context, groupID string, req *CreateCategoryRequest) (*CategoryResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	gID, err := identifier.ParseID(groupID)
	if err != nil {
		return nil, err
	}

	repo := u.uow.TrackingRepository()
	group, err := repo.FindByID(ctx, gID)
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

	budget, err := money.NewFromFloat(req.Budget)
	if err != nil {
		return nil, err
	}

	category, err := group.CreateCategory(id, name, description, req.IsRecurrent, startMonth, endMonth, budget)
	if err != nil {
		return nil, err
	}

	if err := repo.Save(ctx, group); err != nil {
		return nil, err
	}

	return u.mapToResponse(category), nil
}

func (u CategoryUseCaseImpl) Update(ctx context.Context, groupID string, req *UpdateCategoryRequest) (*CategoryResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	gID, err := identifier.ParseID(groupID)
	if err != nil {
		return nil, err
	}

	cID, err := identifier.ParseID(req.ID)
	if err != nil {
		return nil, err
	}

	repo := u.uow.TrackingRepository()
	group, err := repo.FindByID(ctx, gID)
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

	budget, err := money.NewFromFloat(req.Budget)
	if err != nil {
		return nil, err
	}

	category, err := group.UpdateCategory(cID, name, description, req.IsRecurrent, startMonth, endMonth, budget)
	if err != nil {
		return nil, err
	}

	if err := repo.Save(ctx, group); err != nil {
		return nil, err
	}

	return u.mapToResponse(category), nil
}

func (u CategoryUseCaseImpl) Delete(ctx context.Context, groupID string, id string) error {
	gID, err := identifier.ParseID(groupID)
	if err != nil {
		return err
	}

	cID, err := identifier.ParseID(id)
	if err != nil {
		return err
	}

	repo := u.uow.TrackingRepository()
	group, err := repo.FindByID(ctx, gID)
	if err != nil {
		return err
	}

	if err := group.RemoveCategory(cID); err != nil {
		return err
	}

	return repo.DeleteCategory(ctx, cID)
}

func (u CategoryUseCaseImpl) Get(ctx context.Context, groupID string, id string) (*CategoryResponse, error) {
	gID, err := identifier.ParseID(groupID)
	if err != nil {
		return nil, err
	}

	cID, err := identifier.ParseID(id)
	if err != nil {
		return nil, err
	}

	repo := u.uow.TrackingRepository()
	group, err := repo.FindByID(ctx, gID)
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

func (u CategoryUseCaseImpl) List(ctx context.Context, groupID string) ([]CategoryResponse, error) {
	gID, err := identifier.ParseID(groupID)
	if err != nil {
		return nil, err
	}

	repo := u.uow.TrackingRepository()
	group, err := repo.FindByID(ctx, gID)
	if err != nil {
		return nil, err
	}

	responses := make([]CategoryResponse, len(group.Categories))
	for i, c := range group.Categories {
		responses[i] = *u.mapToResponse(c)
	}

	return responses, nil
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
