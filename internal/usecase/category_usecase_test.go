package usecase

import (
	"context"
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

func newTestCategoryUseCase(repo *MockGroupRepository) CategoryUseCaseImpl {
	if repo == nil {
		repo = &MockGroupRepository{}
	}
	return NewCategoryUseCase(
		&MockUnitOfWork{TrackingRepo: repo},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func TestCategoryUseCase_Create(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)
	validReq := &CreateCategoryRequest{
		Name:        "Test Category",
		Description: "Test Description",
		StartMonth:  "2023-01",
		IsRecurrent: false,
	}

	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestCategoryUseCase(nil)
		resp, err := usecase.Create(context.Background(), group.ID.String(), nil)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "request cannot be nil")
	})

	t.Run("returns error for invalid group ID", func(t *testing.T) {
		usecase := newTestCategoryUseCase(nil)
		resp, err := usecase.Create(context.Background(), "invalid-id", validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error when group not found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(tracking.Group{}, tracking.ErrGroupNotFound)

		usecase := newTestCategoryUseCase(repo)
		resp, err := usecase.Create(context.Background(), group.ID.String(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})

	t.Run("creates category and saves group", func(t *testing.T) {
		var savedGroup tracking.Group
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)
		repo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedGroup = args.Get(1).(tracking.Group)
		})

		usecase := newTestCategoryUseCase(repo)

		resp, err := usecase.Create(context.Background(), group.ID.String(), validReq)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, validReq.Name, resp.Name)
		assert.NotEmpty(t, resp.ID)
		assert.Len(t, savedGroup.Categories, 1)
		assert.Equal(t, validReq.Name, savedGroup.Categories[0].Name.Value())
	})
}

func TestCategoryUseCase_Update(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)

	// Create a category on the group
	catID, _ := identifier.NewID()
	name, _ := tracking.NewNameVO("Old Name")
	desc, _ := tracking.NewDescriptionVO("Old Description")
	startMonth, _ := tracking.ParseMonth("2023-01")
	_, _ = group.CreateCategory(catID, name, desc, false, startMonth, tracking.Month{}, money.Money{})

	validReq := &UpdateCategoryRequest{
		ID:          catID.String(),
		Name:        "New Name",
		Description: "New Description",
		StartMonth:  "2023-02",
		IsRecurrent: false,
	}

	t.Run("returns error when group not found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(tracking.Group{}, tracking.ErrGroupNotFound)

		usecase := newTestCategoryUseCase(repo)
		resp, err := usecase.Update(context.Background(), group.ID.String(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})

	t.Run("returns error when category not found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo)
		req := *validReq

		newID, _ := identifier.NewID()
		req.ID = newID.String() // Different ID

		resp, err := usecase.Update(context.Background(), group.ID.String(), &req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrCategoryNotFound)
	})

	t.Run("updates category and saves group", func(t *testing.T) {
		var savedGroup tracking.Group
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)
		repo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedGroup = args.Get(1).(tracking.Group)
		})

		usecase := newTestCategoryUseCase(repo)

		resp, err := usecase.Update(context.Background(), group.ID.String(), validReq)

		require.NoError(t, err)
		assert.Equal(t, "New Name", resp.Name)
		assert.Equal(t, "New Description", resp.Description)

		require.Len(t, savedGroup.Categories, 1)
		assert.Equal(t, "New Name", savedGroup.Categories[0].Name.Value())
	})
}

func TestCategoryUseCase_Delete(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)

	// Create a category on the group
	catID, _ := identifier.NewID()
	name, _ := tracking.NewNameVO("To Delete")
	desc, _ := tracking.NewDescriptionVO("Desc")
	startMonth, _ := tracking.ParseMonth("2023-01")
	_, _ = group.CreateCategory(catID, name, desc, false, startMonth, tracking.Month{}, money.Money{})

	t.Run("returns error when category not found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo)

		newID, _ := identifier.NewID()
		err := usecase.Delete(context.Background(), group.ID.String(), newID.String())
		assert.ErrorIs(t, err, tracking.ErrCategoryNotFound)
	})

	t.Run("deletes category and saves group", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)
		repo.On("DeleteCategory", mock.Anything, catID).Return(nil)

		usecase := newTestCategoryUseCase(repo)

		err := usecase.Delete(context.Background(), group.ID.String(), catID.String())
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestCategoryUseCase_Get(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)

	catID, _ := identifier.NewID()
	name, _ := tracking.NewNameVO("Category")
	desc, _ := tracking.NewDescriptionVO("Desc")
	startMonth, _ := tracking.ParseMonth("2023-01")
	_, _ = group.CreateCategory(catID, name, desc, false, startMonth, tracking.Month{}, money.Money{})

	t.Run("returns category successfully", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo)

		resp, err := usecase.Get(context.Background(), group.ID.String(), catID.String())
		require.NoError(t, err)
		assert.Equal(t, catID.String(), resp.ID)
		assert.Equal(t, "Category", resp.Name)
	})

	t.Run("returns error when category not found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo)

		newID, _ := identifier.NewID()
		resp, err := usecase.Get(context.Background(), group.ID.String(), newID.String())
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrCategoryNotFound)
	})
}

func TestCategoryUseCase_List(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)

	catID1, _ := identifier.NewID()
	name1, _ := tracking.NewNameVO("Cat 1")
	desc1, _ := tracking.NewDescriptionVO("Desc 1")
	startMonth, _ := tracking.ParseMonth("2023-01")
	_, _ = group.CreateCategory(catID1, name1, desc1, false, startMonth, tracking.Month{}, money.Money{})

	catID2, _ := identifier.NewID()
	name2, _ := tracking.NewNameVO("Cat 2")
	desc2, _ := tracking.NewDescriptionVO("Desc 2")
	_, _ = group.CreateCategory(catID2, name2, desc2, false, startMonth, tracking.Month{}, money.Money{})

	t.Run("returns list of categories", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo)

		resps, err := usecase.List(context.Background(), group.ID.String())
		require.NoError(t, err)
		assert.Len(t, resps, 2)
		assert.Equal(t, catID1.String(), resps[0].ID)
		assert.Equal(t, catID2.String(), resps[1].ID)
	})
}
