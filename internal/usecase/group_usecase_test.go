package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestGroupUseCase(repo *MockGroupRepository) GroupUseCaseImpl {
	if repo == nil {
		repo = &MockGroupRepository{}
	}
	return NewGroupUseCase(
		&MockUnitOfWork{TrackingRepo: repo},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func newTestGroup(t *testing.T, userID identifier.ID) *tracking.Group {
	t.Helper()

	id, err := identifier.NewID()
	require.NoError(t, err)

	name, err := tracking.NewNameVO("Test Group")
	require.NoError(t, err)

	desc, err := tracking.NewDescriptionVO("Test Description")
	require.NoError(t, err)

	order, err := tracking.NewOrderVO(0)
	require.NoError(t, err)

	return tracking.NewGroup(id, userID, name, desc, order)
}

func TestGroupUseCase_Create(t *testing.T) {
	validUserID, _ := identifier.NewID()
	validReq := &CreateGroupRequest{
		UserID:      validUserID.String(),
		Name:        "Test Group",
		Description: "Test Description",
	}

	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestGroupUseCase(nil)
		resp, err := usecase.Create(context.Background(), nil)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "request cannot be nil")
	})

	t.Run("returns error for invalid user ID", func(t *testing.T) {
		usecase := newTestGroupUseCase(nil)
		req := *validReq
		req.UserID = "invalid-id"
		resp, err := usecase.Create(context.Background(), &req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error for invalid name", func(t *testing.T) {
		usecase := newTestGroupUseCase(nil)
		req := &CreateGroupRequest{
			UserID:      validUserID.String(),
			Name:        "",
			Description: "Test Description",
		}
		resp, err := usecase.Create(context.Background(), req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrEmptyName)
	})

	t.Run("saves group and returns response", func(t *testing.T) {
		var savedGroup tracking.Group
		repo := &MockGroupRepository{}
		txRepo := &MockGroupRepository{}
		txRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedGroup = args.Get(1).(tracking.Group)
		})

		txUOW := &MockUnitOfWork{TrackingRepo: txRepo}
		baseUOW := &MockUnitOfWork{TrackingRepo: repo}
		baseUOW.On("Begin", mock.Anything).Return(txUOW, nil)
		txUOW.On("Commit").Return(nil)

		usecase := NewGroupUseCase(
			baseUOW,
			slog.New(slog.NewTextHandler(io.Discard, nil)),
		)

		resp, err := usecase.Create(context.Background(), validReq)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, validReq.Name, resp.Name)
		assert.Equal(t, validReq.Description, resp.Description)
		assert.NotEmpty(t, resp.ID)

		assert.Equal(t, validReq.Name, savedGroup.Name.Value())
		assert.Equal(t, validReq.Description, savedGroup.Description.Value())
		assert.Equal(t, validUserID, savedGroup.UserID)
	})
}

func TestGroupUseCase_Update(t *testing.T) {
	validUserID, _ := identifier.NewID()
	existingGroup := newTestGroup(t, validUserID)
	validReq := &UpdateGroupRequest{
		ID:          existingGroup.ID.String(),
		UserID:      validUserID.String(),
		Name:        "Updated Group",
		Description: "Updated Description",
	}

	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestGroupUseCase(nil)
		resp, err := usecase.Update(context.Background(), nil)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "request cannot be nil")
	})

	t.Run("returns error for invalid group ID", func(t *testing.T) {
		usecase := newTestGroupUseCase(nil)
		req := &UpdateGroupRequest{ID: "invalid", UserID: validUserID.String()}
		resp, err := usecase.Update(context.Background(), req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error when group not found", func(t *testing.T) {
		expectedErr := errors.New("not found")
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(tracking.Group{}, expectedErr)

		usecase := newTestGroupUseCase(repo)
		resp, err := usecase.Update(context.Background(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns error when unauthorized", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		otherUserGroup := newTestGroup(t, otherUserID)
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*otherUserGroup, nil)

		usecase := newTestGroupUseCase(repo)

		req := &UpdateGroupRequest{ID: otherUserGroup.ID.String(), UserID: validUserID.String(), Name: "Test", Description: "Test"}
		resp, err := usecase.Update(context.Background(), req)

		assert.Nil(t, resp)
		assert.EqualError(t, err, "unauthorized")
	})

	t.Run("updates group and returns response", func(t *testing.T) {
		var savedGroup tracking.Group
		repo := &MockGroupRepository{}
		txRepo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*existingGroup, nil)
		txRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedGroup = args.Get(1).(tracking.Group)
		})

		txUOW := &MockUnitOfWork{TrackingRepo: txRepo}
		baseUOW := &MockUnitOfWork{TrackingRepo: repo}
		baseUOW.On("Begin", mock.Anything).Return(txUOW, nil)
		txUOW.On("Commit").Return(nil)

		usecase := NewGroupUseCase(
			baseUOW,
			slog.New(slog.NewTextHandler(io.Discard, nil)),
		)

		resp, err := usecase.Update(context.Background(), validReq)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, validReq.Name, resp.Name)
		assert.Equal(t, validReq.Description, resp.Description)
		assert.Equal(t, existingGroup.ID.String(), resp.ID)

		assert.Equal(t, validReq.Name, savedGroup.Name.Value())
		assert.Equal(t, validReq.Description, savedGroup.Description.Value())
	})
}

func TestGroupUseCase_Delete(t *testing.T) {
	validUserID, _ := identifier.NewID()
	existingGroup := newTestGroup(t, validUserID)

	t.Run("returns error for invalid group ID", func(t *testing.T) {
		usecase := newTestGroupUseCase(nil)
		err := usecase.Delete(context.Background(), validUserID.String(), "invalid")
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error when group not found", func(t *testing.T) {
		expectedErr := errors.New("not found")
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(tracking.Group{}, expectedErr)

		usecase := newTestGroupUseCase(repo)
		err := usecase.Delete(context.Background(), validUserID.String(), existingGroup.ID.String())
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns error when unauthorized", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		otherUserGroup := newTestGroup(t, otherUserID)
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*otherUserGroup, nil)

		usecase := newTestGroupUseCase(repo)
		err := usecase.Delete(context.Background(), validUserID.String(), otherUserGroup.ID.String())
		assert.EqualError(t, err, "unauthorized")
	})

	t.Run("deletes group successfully", func(t *testing.T) {
		var deletedID identifier.ID
		repo := &MockGroupRepository{}
		txRepo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*existingGroup, nil)
		txRepo.On("Delete", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			deletedID = args.Get(1).(identifier.ID)
		})

		txUOW := &MockUnitOfWork{TrackingRepo: txRepo}
		baseUOW := &MockUnitOfWork{TrackingRepo: repo}
		baseUOW.On("Begin", mock.Anything).Return(txUOW, nil)
		txUOW.On("Commit").Return(nil)

		usecase := NewGroupUseCase(
			baseUOW,
			slog.New(slog.NewTextHandler(io.Discard, nil)),
		)

		err := usecase.Delete(context.Background(), validUserID.String(), existingGroup.ID.String())
		require.NoError(t, err)
		assert.Equal(t, existingGroup.ID, deletedID)
	})
}

func TestGroupUseCase_Get(t *testing.T) {
	validUserID, _ := identifier.NewID()
	existingGroup := newTestGroup(t, validUserID)

	t.Run("returns group successfully", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*existingGroup, nil)

		usecase := newTestGroupUseCase(repo)

		resp, err := usecase.Get(context.Background(), validUserID.String(), existingGroup.ID.String())
		require.NoError(t, err)
		assert.Equal(t, existingGroup.ID.String(), resp.ID)
		assert.Equal(t, existingGroup.Name.Value(), resp.Name)
	})

	t.Run("returns unauthorized for different user", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*existingGroup, nil)

		usecase := newTestGroupUseCase(repo)

		resp, err := usecase.Get(context.Background(), otherUserID.String(), existingGroup.ID.String())
		assert.Nil(t, resp)
		assert.EqualError(t, err, "unauthorized")
	})
}

func TestGroupUseCase_List(t *testing.T) {
	validUserID, _ := identifier.NewID()
	grp1 := newTestGroup(t, validUserID)
	grp2 := newTestGroup(t, validUserID)

	t.Run("returns list of groups", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByUserID", mock.Anything, mock.Anything).Return([]tracking.Group{*grp1, *grp2}, nil)

		usecase := newTestGroupUseCase(repo)

		resps, err := usecase.List(context.Background(), validUserID.String())
		require.NoError(t, err)
		assert.Len(t, resps, 2)
		assert.Equal(t, grp1.ID.String(), resps[0].ID)
		assert.Equal(t, grp2.ID.String(), resps[1].ID)
	})

	t.Run("returns empty list if none found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByUserID", mock.Anything, mock.Anything).Return([]tracking.Group{}, nil)

		usecase := newTestGroupUseCase(repo)

		resps, err := usecase.List(context.Background(), validUserID.String())
		require.NoError(t, err)
		assert.Empty(t, resps)
	})
}

func TestGroupUseCase_List_Mapping(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)

	catID, _ := identifier.NewID()
	catName, _ := tracking.NewNameVO("Test Category")
	catDesc, _ := tracking.NewDescriptionVO("Test Desc")
	startMonth, _ := tracking.ParseMonth("2023-01")
	budget, _ := money.NewFromFloat(123.45, "USD")

	category, _ := tracking.NewCategory(
		catID,
		group.ID,
		catName,
		catDesc,
		false,
		startMonth,
		tracking.Month{},
		budget,
	)
	_ = group.AddCategory(category)

	t.Run("verifies budget mapping", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByUserID", mock.Anything, mock.Anything).Return([]tracking.Group{*group}, nil)

		usecase := newTestGroupUseCase(repo)

		resps, err := usecase.List(context.Background(), validUserID.String())
		require.NoError(t, err)
		assert.Len(t, resps, 1)
		assert.Len(t, resps[0].Categories, 1)
		assert.Equal(t, 123.45, resps[0].Categories[0].Budget)
	})
}
