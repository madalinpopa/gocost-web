package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestDashboardUseCase(trackingRepo *MockGroupRepository, incomeRepo *MockIncomeRepository, expenseRepo *MockExpenseRepository) DashboardUseCaseImpl {
	if trackingRepo == nil {
		trackingRepo = &MockGroupRepository{}
	}
	if incomeRepo == nil {
		incomeRepo = &MockIncomeRepository{}
	}
	if expenseRepo == nil {
		expenseRepo = &MockExpenseRepository{}
	}

	return NewDashboardUseCase(
		&MockUnitOfWork{
			TrackingRepo: trackingRepo,
			IncomeRepo:   incomeRepo,
			ExpenseRepo:  expenseRepo,
		},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func newDashboardGroup(t *testing.T, userID identifier.ID, name string, order int) *tracking.Group {
	t.Helper()

	id, err := identifier.NewID()
	require.NoError(t, err)

	nameVO, err := tracking.NewNameVO(name)
	require.NoError(t, err)

	descVO, err := tracking.NewDescriptionVO("Description for " + name)
	require.NoError(t, err)

	orderVO, err := tracking.NewOrderVO(order)
	require.NoError(t, err)

	return tracking.NewGroup(id, userID, nameVO, descVO, orderVO)
}

func addDashboardCategory(t *testing.T, group *tracking.Group, name string, budgetCents int64) *tracking.Category {
	t.Helper()

	id, err := identifier.NewID()
	require.NoError(t, err)

	nameVO, err := tracking.NewNameVO(name)
	require.NoError(t, err)

	descVO, err := tracking.NewDescriptionVO("Description for " + name)
	require.NoError(t, err)

	startMonth, err := tracking.ParseMonth("2024-01")
	require.NoError(t, err)

	budget, err := money.New(budgetCents, "USD")
	require.NoError(t, err)

	category, err := group.CreateCategory(id, nameVO, descVO, true, startMonth, tracking.Month{}, budget)
	require.NoError(t, err)

	return category
}

func newDashboardExpense(t *testing.T, categoryID identifier.ID, amount float64, description string, spentAt time.Time, payment expense.PaymentStatus) expense.Expense {
	t.Helper()

	id, err := identifier.NewID()
	require.NoError(t, err)

	m, err := money.NewFromFloat(amount, "USD")
	require.NoError(t, err)

	descVO, err := expense.NewExpenseDescriptionVO(description)
	require.NoError(t, err)

	exp, err := expense.NewExpense(id, categoryID, m, descVO, spentAt, payment)
	require.NoError(t, err)

	return *exp
}

func mustMoneyFromFloat(t *testing.T, amount float64) money.Money {
	t.Helper()

	m, err := money.NewFromFloat(amount, "USD")
	require.NoError(t, err)
	return m
}

func TestDashboardUseCase_Get_Validation(t *testing.T) {
	usecase := newTestDashboardUseCase(nil, nil, nil)

	t.Run("returns error for nil request", func(t *testing.T) {
		resp, err := usecase.Get(context.Background(), nil)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "request cannot be nil")
	})

	t.Run("returns error for empty user id", func(t *testing.T) {
		resp, err := usecase.Get(context.Background(), &DashboardRequest{UserID: "", Month: "2024-01"})
		assert.Nil(t, resp)
		assert.EqualError(t, err, "user id cannot be empty")
	})

	t.Run("returns error for empty month", func(t *testing.T) {
		resp, err := usecase.Get(context.Background(), &DashboardRequest{UserID: "user-id", Month: ""})
		assert.Nil(t, resp)
		assert.EqualError(t, err, "month cannot be empty")
	})

	t.Run("returns error for invalid user id", func(t *testing.T) {
		resp, err := usecase.Get(context.Background(), &DashboardRequest{UserID: "invalid", Month: "2024-01"})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})
}

func TestDashboardUseCase_Get_RepositoryErrors(t *testing.T) {
	userID, _ := identifier.NewID()
	month := "2024-02"

	incomeTotal := mustMoneyFromFloat(t, 0)
	expenseTotal := mustMoneyFromFloat(t, 0)

	trackingErr := errors.New("tracking error")
	incomeErr := errors.New("income error")
	expenseTotalErr := errors.New("expense total error")
	categoryTotalsErr := errors.New("category totals error")
	expenseListErr := errors.New("expense list error")

	tests := []struct {
		name        string
		setup       func(*MockGroupRepository, *MockIncomeRepository, *MockExpenseRepository)
		expectedErr error
	}{
		{
			name: "tracking repository error",
			setup: func(trackingRepo *MockGroupRepository, _ *MockIncomeRepository, _ *MockExpenseRepository) {
				trackingRepo.On("FindByUserIDAndMonth", mock.Anything, userID, month).Return(nil, trackingErr)
			},
			expectedErr: trackingErr,
		},
		{
			name: "income repository error",
			setup: func(trackingRepo *MockGroupRepository, incomeRepo *MockIncomeRepository, _ *MockExpenseRepository) {
				trackingRepo.On("FindByUserIDAndMonth", mock.Anything, userID, month).Return([]tracking.Group{}, nil)
				incomeRepo.On("TotalByUserIDAndMonth", mock.Anything, userID, month).Return(money.Money{}, incomeErr)
			},
			expectedErr: incomeErr,
		},
		{
			name: "expense total error",
			setup: func(trackingRepo *MockGroupRepository, incomeRepo *MockIncomeRepository, expenseRepo *MockExpenseRepository) {
				trackingRepo.On("FindByUserIDAndMonth", mock.Anything, userID, month).Return([]tracking.Group{}, nil)
				incomeRepo.On("TotalByUserIDAndMonth", mock.Anything, userID, month).Return(incomeTotal, nil)
				expenseRepo.On("Total", mock.Anything, userID, month).Return(money.Money{}, expenseTotalErr)
			},
			expectedErr: expenseTotalErr,
		},
		{
			name: "expense category totals error",
			setup: func(trackingRepo *MockGroupRepository, incomeRepo *MockIncomeRepository, expenseRepo *MockExpenseRepository) {
				trackingRepo.On("FindByUserIDAndMonth", mock.Anything, userID, month).Return([]tracking.Group{}, nil)
				incomeRepo.On("TotalByUserIDAndMonth", mock.Anything, userID, month).Return(incomeTotal, nil)
				expenseRepo.On("Total", mock.Anything, userID, month).Return(expenseTotal, nil)
				expenseRepo.On("TotalsByCategoryAndMonth", mock.Anything, userID, month).Return(nil, categoryTotalsErr)
			},
			expectedErr: categoryTotalsErr,
		},
		{
			name: "expense list error",
			setup: func(trackingRepo *MockGroupRepository, incomeRepo *MockIncomeRepository, expenseRepo *MockExpenseRepository) {
				trackingRepo.On("FindByUserIDAndMonth", mock.Anything, userID, month).Return([]tracking.Group{}, nil)
				incomeRepo.On("TotalByUserIDAndMonth", mock.Anything, userID, month).Return(incomeTotal, nil)
				expenseRepo.On("Total", mock.Anything, userID, month).Return(expenseTotal, nil)
				expenseRepo.On("TotalsByCategoryAndMonth", mock.Anything, userID, month).Return([]expense.CategoryTotals{}, nil)
				expenseRepo.On("FindByUserIDAndMonth", mock.Anything, userID, month).Return(nil, expenseListErr)
			},
			expectedErr: expenseListErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingRepo := &MockGroupRepository{}
			incomeRepo := &MockIncomeRepository{}
			expenseRepo := &MockExpenseRepository{}
			tt.setup(trackingRepo, incomeRepo, expenseRepo)

			usecase := newTestDashboardUseCase(trackingRepo, incomeRepo, expenseRepo)
			resp, err := usecase.Get(context.Background(), &DashboardRequest{
				UserID: userID.String(),
				Month:  month,
			})

			assert.Nil(t, resp)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestDashboardUseCase_Get_Success(t *testing.T) {
	userID, _ := identifier.NewID()
	month := "2024-02"

	groupA := newDashboardGroup(t, userID, "Group A", 0)
	groupB := newDashboardGroup(t, userID, "Group B", 1)

	categoryA := addDashboardCategory(t, groupA, "Food", 10000)
	categoryB := addDashboardCategory(t, groupA, "Transport", 20000)
	categoryC := addDashboardCategory(t, groupB, "Utilities", 5000)

	spentAt := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	paidAt := time.Date(2024, 2, 12, 0, 0, 0, 0, time.UTC)
	paidStatus, err := expense.NewPaidStatus(paidAt)
	require.NoError(t, err)

	expenseA := newDashboardExpense(t, categoryA.ID, 50.0, "Groceries", spentAt, paidStatus)
	expenseB := newDashboardExpense(t, categoryA.ID, 25.0, "Snacks", spentAt, expense.NewUnpaidStatus())
	expenseC := newDashboardExpense(t, categoryB.ID, 20.0, "Bus", spentAt, paidStatus)
	otherCategoryID, _ := identifier.NewID()
	expenseD := newDashboardExpense(t, otherCategoryID, 10.0, "Other", spentAt, paidStatus)

	incomeTotal := mustMoneyFromFloat(t, 150.0)
	expenseTotal := mustMoneyFromFloat(t, 95.0)

	categoryTotals := []expense.CategoryTotals{
		{
			CategoryID: categoryA.ID,
			Total:      mustMoneyFromFloat(t, 75.0),
			PaidTotal:  mustMoneyFromFloat(t, 50.0),
		},
		{
			CategoryID: categoryB.ID,
			Total:      mustMoneyFromFloat(t, 20.0),
			PaidTotal:  mustMoneyFromFloat(t, 20.0),
		},
	}

	trackingRepo := &MockGroupRepository{}
	trackingRepo.On("FindByUserIDAndMonth", mock.Anything, userID, month).Return([]tracking.Group{*groupA, *groupB}, nil)

	incomeRepo := &MockIncomeRepository{}
	incomeRepo.On("TotalByUserIDAndMonth", mock.Anything, userID, month).Return(incomeTotal, nil)

	expenseRepo := &MockExpenseRepository{}
	expenseRepo.On("Total", mock.Anything, userID, month).Return(expenseTotal, nil)
	expenseRepo.On("TotalsByCategoryAndMonth", mock.Anything, userID, month).Return(categoryTotals, nil)
	expenseRepo.On("FindByUserIDAndMonth", mock.Anything, userID, month).Return([]expense.Expense{expenseA, expenseB, expenseC, expenseD}, nil)

	usecase := newTestDashboardUseCase(trackingRepo, incomeRepo, expenseRepo)
	resp, err := usecase.Get(context.Background(), &DashboardRequest{
		UserID: userID.String(),
		Month:  month,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, incomeTotal.Cents(), resp.TotalIncomeCents)
	assert.Equal(t, expenseTotal.Cents(), resp.TotalExpensesCents)
	assert.Equal(t, int64(35000), resp.TotalBudgetedCents)
	assert.Equal(t, int64(7000), resp.PaidExpensesCents)

	groupByID := make(map[string]DashboardGroupResponse)
	for _, group := range resp.Groups {
		groupByID[group.ID] = group
	}

	groupAResp, ok := groupByID[groupA.ID.String()]
	require.True(t, ok)
	assert.Equal(t, groupA.Name.Value(), groupAResp.Name)
	assert.Equal(t, groupA.Description.Value(), groupAResp.Description)
	assert.Equal(t, groupA.Order.Value(), groupAResp.Order)
	require.Len(t, groupAResp.Categories, 2)

	groupBResp, ok := groupByID[groupB.ID.String()]
	require.True(t, ok)
	require.Len(t, groupBResp.Categories, 1)

	var categoryAResp DashboardCategoryResponse
	var categoryAFound bool
	for _, category := range groupAResp.Categories {
		if category.ID == categoryA.ID.String() {
			categoryAResp = category
			categoryAFound = true
			break
		}
	}
	require.True(t, categoryAFound)
	require.Equal(t, categoryA.ID.String(), categoryAResp.ID)
	assert.Equal(t, categoryA.Name.Value(), categoryAResp.Name)
	assert.Equal(t, categoryA.Description.Value(), categoryAResp.Description)
	assert.Equal(t, categoryA.Budget.Cents(), categoryAResp.BudgetCents)
	assert.Equal(t, mustMoneyFromFloat(t, 75.0).Cents(), categoryAResp.SpentCents)
	assert.Equal(t, mustMoneyFromFloat(t, 50.0).Cents(), categoryAResp.PaidSpentCents)
	require.Len(t, categoryAResp.Expenses, 2)
	assert.Equal(t, expenseA.Description.Value(), categoryAResp.Expenses[0].Description)
	assert.Equal(t, expenseB.Description.Value(), categoryAResp.Expenses[1].Description)
	require.NotNil(t, categoryAResp.Expenses[0].PaidAt)
	assert.True(t, categoryAResp.Expenses[0].IsPaid)
	assert.WithinDuration(t, paidAt, *categoryAResp.Expenses[0].PaidAt, time.Second)

	var categoryCResp DashboardCategoryResponse
	var categoryCFound bool
	for _, category := range groupBResp.Categories {
		if category.ID == categoryC.ID.String() {
			categoryCResp = category
			categoryCFound = true
			break
		}
	}
	require.True(t, categoryCFound)
	require.Equal(t, categoryC.ID.String(), categoryCResp.ID)
	assert.Equal(t, int64(0), categoryCResp.SpentCents)
	assert.Equal(t, int64(0), categoryCResp.PaidSpentCents)
	require.Empty(t, categoryCResp.Expenses)
}
