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

func newTestCategoryUseCase(groupRepo *MockGroupRepository, userRepo *MockUserRepository, expenseRepo *MockExpenseRepository) CategoryUseCaseImpl {
	if groupRepo == nil {
		groupRepo = &MockGroupRepository{}
	}
	if userRepo == nil {
		userRepo = &MockUserRepository{}
	}
	if expenseRepo == nil {
		expenseRepo = &MockExpenseRepository{}
	}
	return NewCategoryUseCase(
		&MockUnitOfWork{TrackingRepo: groupRepo, UserRepo: userRepo, ExpenseRepo: expenseRepo},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func TestCategoryUseCase_Create(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)
	validReq := &CreateCategoryRequest{
		UserID:      validUserID.String(),
		GroupID:     group.ID.String(),
		Currency:    "USD",
		Name:        "Test Category",
		Description: "Test Description",
		StartMonth:  "2023-01",
		IsRecurrent: false,
		Budget:      100.0,
	}

	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestCategoryUseCase(nil, nil, nil)
		resp, err := usecase.Create(context.Background(), nil)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "request cannot be nil")
	})

	t.Run("returns error for invalid group ID", func(t *testing.T) {
		usecase := newTestCategoryUseCase(nil, nil, nil)
		req := *validReq
		req.GroupID = "invalid-id"
		resp, err := usecase.Create(context.Background(), &req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error when group not found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(tracking.Group{}, tracking.ErrGroupNotFound)

		usecase := newTestCategoryUseCase(repo, nil, nil)
		resp, err := usecase.Create(context.Background(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})

	t.Run("returns error when user does not own group", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo, nil, nil)
		otherUserID, _ := identifier.NewID()
		req := *validReq
		req.UserID = otherUserID.String()

		resp, err := usecase.Create(context.Background(), &req)
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

		usecase := newTestCategoryUseCase(repo, nil, nil)

		resp, err := usecase.Create(context.Background(), validReq)

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
		UserID:      validUserID.String(),
		GroupID:     group.ID.String(),
		Currency:    "USD",
		Name:        "New Name",
		Description: "New Description",
		StartMonth:  "2023-02",
		IsRecurrent: false,
		Budget:      100.0,
	}

	t.Run("returns error when group not found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(tracking.Group{}, tracking.ErrGroupNotFound)

		usecase := newTestCategoryUseCase(repo, nil, nil)
		resp, err := usecase.Update(context.Background(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})

	t.Run("returns error when user does not own group", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo, nil, nil)
		otherUserID, _ := identifier.NewID()
		req := *validReq
		req.UserID = otherUserID.String()

		resp, err := usecase.Update(context.Background(), &req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})

	t.Run("returns error when category not found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo, nil, nil)
		req := *validReq

		newID, _ := identifier.NewID()
		req.ID = newID.String() // Different ID

		resp, err := usecase.Update(context.Background(), &req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrCategoryNotFound)
	})

	t.Run("forks recurrent category when updating in future month", func(t *testing.T) {
		var savedGroup tracking.Group
		repo := &MockGroupRepository{}
		expenseRepo := &MockExpenseRepository{}

		// Setup clean group and category
		forkGroup := newTestGroup(t, validUserID)
		fCatID, _ := identifier.NewID()
		fName, _ := tracking.NewNameVO("Recurrent Cat")
		fDesc, _ := tracking.NewDescriptionVO("Desc")
		fStart, _ := tracking.ParseMonth("2023-01")
		fBudget, _ := money.NewFromFloat(100.0, "USD")
		_, _ = forkGroup.CreateCategory(fCatID, fName, fDesc, true, fStart, tracking.Month{}, fBudget)

		repo.On("FindByID", mock.Anything, mock.Anything).Return(*forkGroup, nil)
		repo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedGroup = args.Get(1).(tracking.Group)
		})
		expenseRepo.On("ReassignCategoryFromMonth", mock.Anything, fCatID, mock.Anything, "2023-03").Return(nil)

		usecase := newTestCategoryUseCase(repo, nil, expenseRepo)

		req := &UpdateCategoryRequest{
			ID:           fCatID.String(),
			UserID:       validUserID.String(),
			GroupID:      forkGroup.ID.String(),
			Currency:     "USD",
			Name:         "Forked Cat",
			Description:  "New Desc",
			StartMonth:   "2023-01", // Original Start
			CurrentMonth: "2023-03", // Future View
			IsRecurrent:  true,
			Budget:       200.0,
		}

		resp, err := usecase.Update(context.Background(), req)

		require.NoError(t, err)
		assert.NotEqual(t, fCatID.String(), resp.ID) // Response should be the NEW category
		assert.Equal(t, "Forked Cat", resp.Name)
		assert.Equal(t, "2023-03", resp.StartMonth)
		assert.Equal(t, 200.0, resp.Budget)

		// Check saved group state
		require.Len(t, savedGroup.Categories, 2)

		// Find original category
		var original, forked *tracking.Category
		for _, c := range savedGroup.Categories {
			if c.ID == fCatID {
				original = c
			} else {
				forked = c
			}
		}

		require.NotNil(t, original)
		require.NotNil(t, forked)

		// Original should end in Feb
		assert.Equal(t, "2023-02", original.EndMonth.Value())
		assert.Equal(t, "Recurrent Cat", original.Name.Value()) // Should preserve old name

		// Forked should start in March
		assert.Equal(t, "2023-03", forked.StartMonth.Value())
		assert.Equal(t, "Forked Cat", forked.Name.Value())
		assert.Equal(t, 200.0, forked.Budget.Amount())
		expenseRepo.AssertExpectations(t)
	})

	t.Run("updates category and saves group", func(t *testing.T) {
		var savedGroup tracking.Group
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)
		repo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedGroup = args.Get(1).(tracking.Group)
		})

		usecase := newTestCategoryUseCase(repo, nil, nil)

		resp, err := usecase.Update(context.Background(), validReq)

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

		usecase := newTestCategoryUseCase(repo, nil, nil)

		newID, _ := identifier.NewID()
		err := usecase.Delete(context.Background(), validUserID.String(), group.ID.String(), newID.String())
		assert.ErrorIs(t, err, tracking.ErrCategoryNotFound)
	})

	t.Run("returns error when user does not own group", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo, nil, nil)
		otherUserID, _ := identifier.NewID()

		err := usecase.Delete(context.Background(), otherUserID.String(), group.ID.String(), catID.String())
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})

	t.Run("deletes category and saves group", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)
		repo.On("DeleteCategory", mock.Anything, catID).Return(nil)

		usecase := newTestCategoryUseCase(repo, nil, nil)

		err := usecase.Delete(context.Background(), validUserID.String(), group.ID.String(), catID.String())
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

		usecase := newTestCategoryUseCase(repo, nil, nil)

		resp, err := usecase.Get(context.Background(), validUserID.String(), group.ID.String(), catID.String())
		require.NoError(t, err)
		assert.Equal(t, catID.String(), resp.ID)
		assert.Equal(t, "Category", resp.Name)
	})

	t.Run("returns error when user does not own group", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo, nil, nil)
		otherUserID, _ := identifier.NewID()

		resp, err := usecase.Get(context.Background(), otherUserID.String(), group.ID.String(), catID.String())
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})

	t.Run("returns error when category not found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo, nil, nil)

		newID, _ := identifier.NewID()
		resp, err := usecase.Get(context.Background(), validUserID.String(), group.ID.String(), newID.String())
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

		usecase := newTestCategoryUseCase(repo, nil, nil)

		resps, err := usecase.List(context.Background(), validUserID.String(), group.ID.String())
		require.NoError(t, err)
		assert.Len(t, resps, 2)
		assert.Equal(t, catID1.String(), resps[0].ID)
		assert.Equal(t, catID2.String(), resps[1].ID)
	})

	t.Run("returns error when user does not own group", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestCategoryUseCase(repo, nil, nil)
		otherUserID, _ := identifier.NewID()

		resps, err := usecase.List(context.Background(), otherUserID.String(), group.ID.String())
		assert.Nil(t, resps)
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})
}
