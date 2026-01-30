package views

import (
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestDashboardPresenter_Present_ProgressBar(t *testing.T) {
	presenter := NewDashboardPresenter("$")

	// Setup data
	date, _ := time.Parse("2006-01", "2024-01")

	groups := []*usecase.GroupResponse{
		{
			ID: "g1",
			Categories: []usecase.CategoryResponse{
				{
					ID:         "c1",
					Name:       "Food",
					StartMonth: "2024-01",
					Budget:     100.0,
				},
				{
					ID:         "c2",
					Name:       "Rent",
					StartMonth: "2024-01",
					Budget:     50.0,
				},
			},
		},
	}

	// c1: 50 paid, 20 unpaid. Total 70.
	// c2: 60 paid. Total 60. (Over budget)
	paidAt := time.Now()
	expenses := []*usecase.ExpenseResponse{
		{CategoryID: "c1", Amount: 50.0, IsPaid: true, PaidAt: &paidAt, SpentAt: date},
		{CategoryID: "c1", Amount: 20.0, IsPaid: false, SpentAt: date},
		{CategoryID: "c2", Amount: 60.0, IsPaid: true, PaidAt: &paidAt, SpentAt: date},
	}

	view := presenter.Present(1000, 130, groups, expenses, date)

	// Verify c1 (Food)
	c1 := view.Groups[0].Categories[0]
	assert.Equal(t, "c1", c1.ID)
	assert.Equal(t, 70.0, c1.Spent)
	assert.Equal(t, 50.0, c1.PaidSpent)
	assert.Equal(t, 20.0, c1.UnpaidSpent)
	assert.Equal(t, 50.0, c1.PaidPercentage)
	assert.Equal(t, 20.0, c1.UnpaidPercentage)
	assert.False(t, c1.IsOverBudget)
	assert.Equal(t, 30.0, c1.RemainingBudget)
	assert.False(t, c1.IsNearBudget)

	// Verify c2 (Rent)
	c2 := view.Groups[0].Categories[1]
	assert.Equal(t, "c2", c2.ID)
	assert.Equal(t, 60.0, c2.Spent)
	assert.Equal(t, 60.0, c2.PaidSpent)
	assert.Equal(t, 0.0, c2.UnpaidSpent)
	assert.Equal(t, 100.0, c2.PaidPercentage) // Capped at 100
	assert.Equal(t, 0.0, c2.UnpaidPercentage)
	assert.True(t, c2.IsOverBudget)
	assert.Equal(t, 10.0, c2.OverBudgetAmount)
	assert.False(t, c2.IsNearBudget)
}

func TestDashboardPresenter_Present_TotalIncome(t *testing.T) {
	presenter := NewDashboardPresenter("$")
	date, _ := time.Parse("2006-01", "2024-01")

	// Case 1: Total income preserved
	view1 := presenter.Present(100, 40, nil, nil, date)
	assert.Equal(t, 100.0, view1.TotalIncome)

	// Case 2: Total income unaffected by expenses
	view2 := presenter.Present(40, 100, nil, nil, date)
	assert.Equal(t, 40.0, view2.TotalIncome)
}

func TestDashboardPresenter_Present_TotalBudgetedStatus(t *testing.T) {
	presenter := NewDashboardPresenter("$")
	date, _ := time.Parse("2006-01", "2024-01")

	makeGroups := func(budget float64) []*usecase.GroupResponse {
		return []*usecase.GroupResponse{
			{
				ID: "g1",
				Categories: []usecase.CategoryResponse{
					{
						ID:         "c1",
						Name:       "Food",
						StartMonth: "2024-01",
						Budget:     budget,
					},
				},
			},
		}
	}

	t.Run("under", func(t *testing.T) {
		view := presenter.Present(100, 0, makeGroups(50), nil, date)
		assert.Equal(t, BudgetStatusUnder, view.TotalBudgetedStatus)
	})

	t.Run("equal", func(t *testing.T) {
		view := presenter.Present(100, 0, makeGroups(100), nil, date)
		assert.Equal(t, BudgetStatusEqual, view.TotalBudgetedStatus)
	})

	t.Run("over", func(t *testing.T) {
		view := presenter.Present(100, 0, makeGroups(150), nil, date)
		assert.Equal(t, BudgetStatusOver, view.TotalBudgetedStatus)
	})
}

func TestDashboardPresenter_Present_TotalBudgetedSubtractsPaidExpensesOnly(t *testing.T) {
	presenter := NewDashboardPresenter("$")
	date, _ := time.Parse("2006-01", "2024-01")

	groups := []*usecase.GroupResponse{
		{
			ID: "g1",
			Categories: []usecase.CategoryResponse{
				{ID: "c1", Name: "Food", StartMonth: "2024-01", Budget: 120.0},
				{ID: "c2", Name: "Rent", StartMonth: "2024-01", Budget: 80.0},
			},
		},
	}

	paidAt := time.Now()
	expenses := []*usecase.ExpenseResponse{
		{CategoryID: "c1", Amount: 30.0, IsPaid: true, PaidAt: &paidAt, SpentAt: date},
		{CategoryID: "c2", Amount: 20.0, IsPaid: false, SpentAt: date},
	}

	view := presenter.Present(500, 50, groups, expenses, date)

	assert.Equal(t, 170.0, view.TotalBudgeted)
}
