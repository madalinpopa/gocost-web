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

func TestDashboardPresenter_Present_Balance(t *testing.T) {
	presenter := NewDashboardPresenter("$")
	date, _ := time.Parse("2006-01", "2024-01")

	// Case 1: Positive Balance
	view1 := presenter.Present(100, 40, nil, nil, date)
	assert.Equal(t, 60.0, view1.Balance)
	assert.Equal(t, 60.0, view1.BalanceAbs)

	// Case 2: Negative Balance
	view2 := presenter.Present(40, 100, nil, nil, date)
	assert.Equal(t, -60.0, view2.Balance)
	assert.Equal(t, 60.0, view2.BalanceAbs)
}
