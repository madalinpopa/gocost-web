package usecase

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/shared/identifier"
	"github.com/madalinpopa/gocost-web/internal/shared/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestExpenseUseCase(trackingRepo *MockGroupRepository, expenseRepo *MockExpenseRepository) ExpenseUseCaseImpl {
	if trackingRepo == nil {
		trackingRepo = &MockGroupRepository{}
	}
	if expenseRepo == nil {
		expenseRepo = &MockExpenseRepository{}
	}
	return NewExpenseUseCase(
		&MockUnitOfWork{TrackingRepo: trackingRepo, ExpenseRepo: expenseRepo},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func newTestExpense(t *testing.T, categoryID identifier.ID) *expense.Expense {
	t.Helper()
	id, _ := identifier.NewID()
	amount, _ := money.NewFromFloat(100.0)
	desc, _ := expense.NewExpenseDescriptionVO("Test Expense")
	payment := expense.NewUnpaidStatus()

	exp, err := expense.NewExpense(id, categoryID, amount, desc, time.Now(), payment)
	require.NoError(t, err)
	return exp
}

func TestExpenseUseCase_Create(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)

	// Create a category on the group to link expense
	catID, _ := identifier.NewID()
	name, _ := tracking.NewNameVO("Category")
	desc, _ := tracking.NewDescriptionVO("Desc")
	startMonth, _ := tracking.ParseMonth("2023-01")
	_, _ = group.CreateCategory(catID, name, desc, false, startMonth, tracking.Month{}, money.Money{})

	validReq := &CreateExpenseRequest{
		CategoryID:  catID.String(),
		Amount:      50.0,
		Description: "Lunch",
		SpentAt:     time.Now(),
		IsPaid:      false,
	}

	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestExpenseUseCase(nil, nil)
		resp, err := usecase.Create(context.Background(), validUserID.String(), nil)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "request cannot be nil")
	})

	t.Run("returns error for invalid user ID", func(t *testing.T) {
		usecase := newTestExpenseUseCase(nil, nil)
		resp, err := usecase.Create(context.Background(), "invalid", validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error when category group not found", func(t *testing.T) {
		repo := &MockGroupRepository{}
		repo.On("FindGroupByCategoryID", mock.Anything, mock.Anything).Return(tracking.Group{}, tracking.ErrGroupNotFound)

		usecase := newTestExpenseUseCase(repo, nil)
		resp, err := usecase.Create(context.Background(), validUserID.String(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})

	t.Run("returns unauthorized when group belongs to another user", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		otherGroup := newTestGroup(t, otherUserID)
		repo := &MockGroupRepository{}
		repo.On("FindGroupByCategoryID", mock.Anything, mock.Anything).Return(*otherGroup, nil)

		usecase := newTestExpenseUseCase(repo, nil)
		resp, err := usecase.Create(context.Background(), validUserID.String(), validReq)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "unauthorized")
	})

	t.Run("creates expense successfully", func(t *testing.T) {
		var savedExpense expense.Expense
		groupRepo := &MockGroupRepository{}
		groupRepo.On("FindGroupByCategoryID", mock.Anything, mock.Anything).Return(*group, nil)

		expenseRepo := &MockExpenseRepository{}
		expenseRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedExpense = args.Get(1).(expense.Expense)
		})

		usecase := newTestExpenseUseCase(groupRepo, expenseRepo)

		resp, err := usecase.Create(context.Background(), validUserID.String(), validReq)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validReq.Amount, resp.Amount)
		assert.Equal(t, validReq.Description, resp.Description)
		assert.Equal(t, validReq.CategoryID, resp.CategoryID)

		assert.Equal(t, validReq.Amount, savedExpense.Amount.Amount())
	})
}

func TestExpenseUseCase_Update(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)

	catID, _ := identifier.NewID()
	name, _ := tracking.NewNameVO("Category")
	desc, _ := tracking.NewDescriptionVO("Desc")
	startMonth, _ := tracking.ParseMonth("2023-01")
	_, _ = group.CreateCategory(catID, name, desc, false, startMonth, tracking.Month{}, money.Money{})

	exp := newTestExpense(t, catID)

	validReq := &UpdateExpenseRequest{
		ID:          exp.ID.String(),
		CategoryID:  catID.String(),
		Amount:      75.0,
		Description: "Updated Lunch",
		SpentAt:     time.Now(),
		IsPaid:      true,
		PaidAt:      &time.Time{},
	}
	now := time.Now()
	validReq.PaidAt = &now

	t.Run("returns error when expense not found", func(t *testing.T) {
		repo := &MockExpenseRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(expense.Expense{}, expense.ErrExpenseNotFound)

		usecase := newTestExpenseUseCase(nil, repo)
		resp, err := usecase.Update(context.Background(), validUserID.String(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expense.ErrExpenseNotFound)
	})

	t.Run("returns unauthorized when expense belongs to another user", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		otherGroup := newTestGroup(t, otherUserID)

		expenseRepo := &MockExpenseRepository{}
		expenseRepo.On("FindByID", mock.Anything, mock.Anything).Return(*exp, nil)

		groupRepo := &MockGroupRepository{}
		groupRepo.On("FindGroupByCategoryID", mock.Anything, mock.Anything).Return(*otherGroup, nil)

		usecase := newTestExpenseUseCase(groupRepo, expenseRepo)

		resp, err := usecase.Update(context.Background(), validUserID.String(), validReq)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "unauthorized")
	})

	t.Run("updates expense successfully", func(t *testing.T) {
		var savedExpense expense.Expense
		expenseRepo := &MockExpenseRepository{}
		expenseRepo.On("FindByID", mock.Anything, mock.Anything).Return(*exp, nil)
		expenseRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedExpense = args.Get(1).(expense.Expense)
		})

		groupRepo := &MockGroupRepository{}
		groupRepo.On("FindGroupByCategoryID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestExpenseUseCase(groupRepo, expenseRepo)

		resp, err := usecase.Update(context.Background(), validUserID.String(), validReq)
		require.NoError(t, err)
		assert.Equal(t, validReq.Amount, resp.Amount)
		assert.Equal(t, validReq.Description, resp.Description)
		assert.True(t, resp.IsPaid)

		assert.Equal(t, validReq.Amount, savedExpense.Amount.Amount())
		assert.Equal(t, validReq.Description, savedExpense.Description.Value())
	})

	t.Run("verifies new category ownership on change", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		otherGroup := newTestGroup(t, otherUserID)
		otherCatID, _ := identifier.NewID()

		expenseRepo := &MockExpenseRepository{}
		expenseRepo.On("FindByID", mock.Anything, mock.Anything).Return(*exp, nil)

		groupRepo := &MockGroupRepository{}
		// First call with catID
		groupRepo.On("FindGroupByCategoryID", mock.Anything, catID).Return(*group, nil)
		// Second call with otherCatID
		groupRepo.On("FindGroupByCategoryID", mock.Anything, otherCatID).Return(*otherGroup, nil)

		usecase := newTestExpenseUseCase(groupRepo, expenseRepo)

		req := *validReq
		req.CategoryID = otherCatID.String()

		resp, err := usecase.Update(context.Background(), validUserID.String(), &req)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "unauthorized")
	})
}

func TestExpenseUseCase_Delete(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)

	catID, _ := identifier.NewID()
	name, _ := tracking.NewNameVO("Category")
	desc, _ := tracking.NewDescriptionVO("Desc")
	startMonth, _ := tracking.ParseMonth("2023-01")
	_, _ = group.CreateCategory(catID, name, desc, false, startMonth, tracking.Month{}, money.Money{})

	exp := newTestExpense(t, catID)

	t.Run("deletes expense successfully", func(t *testing.T) {
		var deletedID identifier.ID
		expenseRepo := &MockExpenseRepository{}
		expenseRepo.On("FindByID", mock.Anything, mock.Anything).Return(*exp, nil)
		expenseRepo.On("Delete", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			deletedID = args.Get(1).(identifier.ID)
		})

		groupRepo := &MockGroupRepository{}
		groupRepo.On("FindGroupByCategoryID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestExpenseUseCase(groupRepo, expenseRepo)

		err := usecase.Delete(context.Background(), validUserID.String(), exp.ID.String())
		require.NoError(t, err)
		assert.Equal(t, exp.ID, deletedID)
	})

	t.Run("returns unauthorized", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		otherGroup := newTestGroup(t, otherUserID)

		expenseRepo := &MockExpenseRepository{}
		expenseRepo.On("FindByID", mock.Anything, mock.Anything).Return(*exp, nil)

		groupRepo := &MockGroupRepository{}
		groupRepo.On("FindGroupByCategoryID", mock.Anything, mock.Anything).Return(*otherGroup, nil)

		usecase := newTestExpenseUseCase(groupRepo, expenseRepo)

		err := usecase.Delete(context.Background(), validUserID.String(), exp.ID.String())
		assert.EqualError(t, err, "unauthorized")
	})
}

func TestExpenseUseCase_Get(t *testing.T) {
	validUserID, _ := identifier.NewID()
	group := newTestGroup(t, validUserID)

	catID, _ := identifier.NewID()
	name, _ := tracking.NewNameVO("Category")
	desc, _ := tracking.NewDescriptionVO("Desc")
	startMonth, _ := tracking.ParseMonth("2023-01")
	_, _ = group.CreateCategory(catID, name, desc, false, startMonth, tracking.Month{}, money.Money{})

	exp := newTestExpense(t, catID)

	t.Run("returns expense successfully", func(t *testing.T) {
		expenseRepo := &MockExpenseRepository{}
		expenseRepo.On("FindByID", mock.Anything, mock.Anything).Return(*exp, nil)

		groupRepo := &MockGroupRepository{}
		groupRepo.On("FindGroupByCategoryID", mock.Anything, mock.Anything).Return(*group, nil)

		usecase := newTestExpenseUseCase(groupRepo, expenseRepo)

		resp, err := usecase.Get(context.Background(), validUserID.String(), exp.ID.String())
		require.NoError(t, err)
		assert.Equal(t, exp.ID.String(), resp.ID)
	})
}

func TestExpenseUseCase_List(t *testing.T) {
	validUserID, _ := identifier.NewID()
	id1, _ := identifier.NewID()
	exp1 := newTestExpense(t, id1)
	id2, _ := identifier.NewID()
	exp2 := newTestExpense(t, id2)

	t.Run("returns list of expenses", func(t *testing.T) {
		expenseRepo := &MockExpenseRepository{}
		expenseRepo.On("FindByUserID", mock.Anything, mock.Anything).Return([]expense.Expense{*exp1, *exp2}, nil)

		usecase := newTestExpenseUseCase(nil, expenseRepo)

		resps, err := usecase.List(context.Background(), validUserID.String())
		require.NoError(t, err)
		assert.Len(t, resps, 2)
	})
}

func TestExpenseUseCase_ListByMonth(t *testing.T) {
	validUserID, _ := identifier.NewID()
	id1, _ := identifier.NewID()
	exp1 := newTestExpense(t, id1)
	id2, _ := identifier.NewID()
	exp2 := newTestExpense(t, id2)

	t.Run("returns list of expenses by month", func(t *testing.T) {
		expenseRepo := &MockExpenseRepository{}
		expenseRepo.On("FindByUserIDAndMonth", mock.Anything, mock.Anything, "2023-10").Return([]expense.Expense{*exp1, *exp2}, nil)

		usecase := newTestExpenseUseCase(nil, expenseRepo)

		resps, err := usecase.ListByMonth(context.Background(), validUserID.String(), "2023-10")
		require.NoError(t, err)
		assert.Len(t, resps, 2)
	})
}

func TestExpenseUseCase_Total(t *testing.T) {
	validUserID, _ := identifier.NewID()

	t.Run("returns total amount", func(t *testing.T) {
		expenseRepo := &MockExpenseRepository{}
		expenseRepo.On("Total", mock.Anything, validUserID, "2023-10").Return(100.0, nil)

		usecase := newTestExpenseUseCase(nil, expenseRepo)

		total, err := usecase.Total(context.Background(), validUserID.String(), "2023-10")
		require.NoError(t, err)
		assert.Equal(t, 100.0, total)
	})

	t.Run("returns error on invalid user id", func(t *testing.T) {
		usecase := newTestExpenseUseCase(nil, nil)
		_, err := usecase.Total(context.Background(), "invalid", "2023-10")
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})
}
