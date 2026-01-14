package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/domain/uow"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
)

type GroupUseCaseImpl struct {
	uow    uow.UnitOfWork
	logger *slog.Logger
}

func NewGroupUseCase(uow uow.UnitOfWork, logger *slog.Logger) GroupUseCaseImpl {
	return GroupUseCaseImpl{
		uow:    uow,
		logger: logger,
	}
}

func (u GroupUseCaseImpl) Create(ctx context.Context, userID string, req *CreateGroupRequest) (*GroupResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	uID, err := identifier.ParseID(userID)
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

	order, err := tracking.NewOrderVO(req.Order)
	if err != nil {
		return nil, err
	}

	id, err := identifier.NewID()
	if err != nil {
		return nil, err
	}

	group := tracking.NewGroup(id, uID, name, description, order)

	repo := u.uow.TrackingRepository()
	if err := repo.Save(ctx, *group); err != nil {
		return nil, err
	}

	return u.mapToResponse(*group), nil
}

func (u GroupUseCaseImpl) Update(ctx context.Context, userID string, req *UpdateGroupRequest) (*GroupResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	groupID, err := identifier.ParseID(req.ID)
	if err != nil {
		return nil, err
	}

	repo := u.uow.TrackingRepository()
	group, err := repo.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.UserID != uID {
		return nil, errors.New("unauthorized")
	}

	name, err := tracking.NewNameVO(req.Name)
	if err != nil {
		return nil, err
	}

	description, err := tracking.NewDescriptionVO(req.Description)
	if err != nil {
		return nil, err
	}

	order, err := tracking.NewOrderVO(req.Order)
	if err != nil {
		return nil, err
	}

	group.Name = name
	group.Description = description
	group.Order = order

	if err := repo.Save(ctx, group); err != nil {
		return nil, err
	}

	return u.mapToResponse(group), nil
}

func (u GroupUseCaseImpl) Delete(ctx context.Context, userID string, id string) error {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return err
	}

	groupID, err := identifier.ParseID(id)
	if err != nil {
		return err
	}

	repo := u.uow.TrackingRepository()
	group, err := repo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	if group.UserID != uID {
		return errors.New("unauthorized")
	}

	return repo.Delete(ctx, groupID)
}

func (u GroupUseCaseImpl) Get(ctx context.Context, userID string, id string) (*GroupResponse, error) {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	groupID, err := identifier.ParseID(id)
	if err != nil {
		return nil, err
	}

	repo := u.uow.TrackingRepository()
	group, err := repo.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.UserID != uID {
		return nil, errors.New("unauthorized")
	}

	return u.mapToResponse(group), nil
}

func (u GroupUseCaseImpl) List(ctx context.Context, userID string) ([]*GroupResponse, error) {
	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	repo := u.uow.TrackingRepository()
	groups, err := repo.FindByUserID(ctx, uID)
	if err != nil {
		return nil, err
	}

	responses := make([]*GroupResponse, len(groups))
	for i, group := range groups {
		responses[i] = u.mapToResponse(group)
	}

	return responses, nil
}

func (u GroupUseCaseImpl) mapToResponse(g tracking.Group) *GroupResponse {
	categories := make([]CategoryResponse, len(g.Categories))
	for i, c := range g.Categories {
		categories[i] = CategoryResponse{
			ID:          c.ID.String(),
			Name:        c.Name.Value(),
			Description: c.Description.Value(),
			IsRecurrent: c.IsRecurrent,
			StartMonth:  c.StartMonth.Value(),
			EndMonth:    c.EndMonth.Value(),
			Budget:      c.Budget.Amount(),
		}
	}

	return &GroupResponse{
		ID:          g.ID.String(),
		Name:        g.Name.Value(),
		Description: g.Description.Value(),
		Order:       g.Order.Value(),
		Categories:  categories,
	}
}
