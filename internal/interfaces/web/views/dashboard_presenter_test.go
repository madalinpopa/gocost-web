package views

import (
	"testing"

	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardPresenter_Present_ProgressBar(t *testing.T) {
	presenter, err := NewDashboardPresenter("USD")
	require.NoError(t, err)

	data := &usecase.DashboardResponse{
		TotalIncomeCents:   100000,
		TotalExpensesCents: 13000,
		TotalBudgetedCents: 15000,
		PaidExpensesCents:  11000,
		Groups: []usecase.DashboardGroupResponse{
			{
				ID: "g1",
				Categories: []usecase.DashboardCategoryResponse{
					{
						ID:             "c1",
						Name:           "Food",
						StartMonth:     "2024-01",
						BudgetCents:    10000,
						SpentCents:     7000,
						PaidSpentCents: 5000,
					},
					{
						ID:             "c2",
						Name:           "Rent",
						StartMonth:     "2024-01",
						BudgetCents:    5000,
						SpentCents:     6000,
						PaidSpentCents: 6000,
					},
				},
			},
		},
	}

	view, err := presenter.Present(data)
	require.NoError(t, err)

	// Verify c1 (Food)
	c1 := view.Groups[0].Categories[0]
	assert.Equal(t, "c1", c1.ID)
	assert.Equal(t, 70.0, c1.Spent.Amount())
	assert.Equal(t, 50.0, c1.PaidSpent.Amount())
	assert.Equal(t, 20.0, c1.UnpaidSpent.Amount())
	assert.Equal(t, 50.0, c1.PaidPercentage)
	assert.Equal(t, 20.0, c1.UnpaidPercentage)
	assert.False(t, c1.IsOverBudget)
	assert.Equal(t, 30.0, c1.RemainingBudget.Amount())
	assert.False(t, c1.IsNearBudget)

	// Verify c2 (Rent)
	c2 := view.Groups[0].Categories[1]
	assert.Equal(t, "c2", c2.ID)
	assert.Equal(t, 60.0, c2.Spent.Amount())
	assert.Equal(t, 60.0, c2.PaidSpent.Amount())
	assert.Equal(t, 0.0, c2.UnpaidSpent.Amount())
	assert.Equal(t, 100.0, c2.PaidPercentage) // Capped at 100
	assert.Equal(t, 0.0, c2.UnpaidPercentage)
	assert.True(t, c2.IsOverBudget)
	assert.Equal(t, 10.0, c2.OverBudgetAmount.Amount())
	assert.False(t, c2.IsNearBudget)
}

func TestDashboardPresenter_Present_TotalIncome(t *testing.T) {
	// Case 1: Total income preserved
	presenter1, err := NewDashboardPresenter("USD")
	require.NoError(t, err)
	view1, err := presenter1.Present(&usecase.DashboardResponse{TotalIncomeCents: 10000})
	require.NoError(t, err)
	assert.Equal(t, 100.0, view1.TotalIncome.Amount())

	// Case 2: Total income unaffected by expenses
	presenter2, err := NewDashboardPresenter("USD")
	require.NoError(t, err)
	view2, err := presenter2.Present(&usecase.DashboardResponse{TotalIncomeCents: 4000})
	require.NoError(t, err)
	assert.Equal(t, 40.0, view2.TotalIncome.Amount())
}

func TestDashboardPresenter_Present_TotalBudgetedStatus(t *testing.T) {
	t.Run("under", func(t *testing.T) {
		presenter, err := NewDashboardPresenter("USD")
		require.NoError(t, err)
		view, err := presenter.Present(&usecase.DashboardResponse{TotalIncomeCents: 10000, TotalBudgetedCents: 5000})
		require.NoError(t, err)
		assert.Equal(t, BudgetStatusUnder, view.TotalBudgetedStatus)
	})

	t.Run("equal", func(t *testing.T) {
		presenter, err := NewDashboardPresenter("USD")
		require.NoError(t, err)
		view, err := presenter.Present(&usecase.DashboardResponse{TotalIncomeCents: 10000, TotalBudgetedCents: 10000})
		require.NoError(t, err)
		assert.Equal(t, BudgetStatusEqual, view.TotalBudgetedStatus)
	})

	t.Run("over", func(t *testing.T) {
		presenter, err := NewDashboardPresenter("USD")
		require.NoError(t, err)
		view, err := presenter.Present(&usecase.DashboardResponse{TotalIncomeCents: 10000, TotalBudgetedCents: 15000})
		require.NoError(t, err)
		assert.Equal(t, BudgetStatusOver, view.TotalBudgetedStatus)
	})
}

func TestDashboardPresenter_Present_TotalBudgetedSubtractsPaidExpensesOnly(t *testing.T) {
	presenter, err := NewDashboardPresenter("USD")
	require.NoError(t, err)

	view, err := presenter.Present(&usecase.DashboardResponse{
		TotalIncomeCents:   50000,
		TotalExpensesCents: 5000,
		TotalBudgetedCents: 20000,
		PaidExpensesCents:  3000,
	})
	require.NoError(t, err)

	assert.Equal(t, 170.0, view.TotalBudgeted.Amount())
}
