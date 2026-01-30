package views

import (
	"math"
	"time"

	"github.com/madalinpopa/gocost-web/internal/usecase"
)

type DashboardPresenter struct {
	Currency string
}

func budgetStatus(totalBudgeted, balance float64) BudgetStatus {
	switch {
	case totalBudgeted < balance:
		return BudgetStatusUnder
	case totalBudgeted > balance:
		return BudgetStatusOver
	default:
		return BudgetStatusEqual
	}
}

func NewDashboardPresenter(currency string) *DashboardPresenter {
	return &DashboardPresenter{
		Currency: currency,
	}
}

func (p *DashboardPresenter) Present(
	totalIncome, totalExpenses float64,
	groups []*usecase.GroupResponse,
	expenses []*usecase.ExpenseResponse,
	date time.Time,
) DashboardView {
	monthStr := date.Format("2006-01")
	// Create a map of expenses by category for easier lookup
	expensesByCategory := make(map[string][]ExpenseView)
	categorySpent := make(map[string]float64)
	totalBudgeted := 0.0
	paidExpensesTotal := 0.0

	for _, exp := range expenses {
		status := StatusUnpaid
		paidAt := ""
		if exp.IsPaid {
			status = StatusPaid
			paidExpensesTotal += exp.Amount
			if exp.PaidAt != nil {
				paidAt = exp.PaidAt.Format("2006-01-02")
			}
		}
		expensesByCategory[exp.CategoryID] = append(expensesByCategory[exp.CategoryID], ExpenseView{
			ID:          exp.ID,
			Amount:      exp.Amount,
			Currency:    p.Currency,
			Description: exp.Description,
			Status:      status,
			SpentAt:     exp.SpentAt.Format("2006-01-02"),
			PaidAt:      paidAt,
		})
		categorySpent[exp.CategoryID] += exp.Amount
	}

	// Round spent amounts to avoid floating point precision issues
	for id, spent := range categorySpent {
		categorySpent[id] = math.Round(spent*100) / 100
	}
	paidExpensesTotal = math.Round(paidExpensesTotal*100) / 100

	// Map to views
	var groupViews []GroupView
	for _, grp := range groups {
		var categoryViews []CategoryView

		for _, cat := range grp.Categories {
			catType := TypeMonthly
			if cat.IsRecurrent {
				catType = TypeRecurrent
			}

			// Check if category is active for this month
			start, _ := time.Parse("2006-01", cat.StartMonth)
			var end time.Time
			if cat.EndMonth != "" {
				end, _ = time.Parse("2006-01", cat.EndMonth)
			}

			showCategory := false
			currentMonthStart, _ := time.Parse("2006-01", monthStr)

			if cat.IsRecurrent {
				if !currentMonthStart.Before(start) {
					if cat.EndMonth == "" || !currentMonthStart.After(end) {
						showCategory = true
					}
				}
			} else {
				if cat.StartMonth == monthStr {
					showCategory = true
				}
			}

			if showCategory {
				totalBudgeted += cat.Budget
				spent := categorySpent[cat.ID]
				catExpenses := expensesByCategory[cat.ID]

				// Calculate progress bar data
				paidSpent := 0.0
				for _, exp := range catExpenses {
					if exp.Status == StatusPaid {
						paidSpent += exp.Amount
					}
				}
				unpaidSpent := spent - paidSpent

				// Percentages
				percentage := 0.0
				paidPercentage := 0.0
				unpaidPercentage := 0.0

				if cat.Budget > 0 {
					percentage = (spent / cat.Budget) * 100
					if percentage > 100 {
						percentage = 100
					}

					paidPercentage = (paidSpent / cat.Budget) * 100
					if paidPercentage > 100 {
						paidPercentage = 100
					}

					unpaidPercentage = (unpaidSpent / cat.Budget) * 100
					if paidPercentage+unpaidPercentage > 100 {
						unpaidPercentage = 100 - paidPercentage
					}
				}

				// Budget calculations
				isOverBudget := spent > cat.Budget
				isNearBudget := !isOverBudget && percentage > 85

				overBudgetAmount := 0.0
				remainingBudget := 0.0

				if isOverBudget {
					overBudgetAmount = spent - cat.Budget
				} else {
					remainingBudget = cat.Budget - spent
				}

				categoryViews = append(categoryViews, CategoryView{
					ID:          cat.ID,
					Name:        cat.Name,
					Type:        catType,
					Description: cat.Description,
					StartMonth:  cat.StartMonth,
					EndMonth:    cat.EndMonth,
					Budget:      cat.Budget,
					Spent:       spent,
					Currency:    p.Currency,
					Expenses:    catExpenses,
					// Progress Bar Fields
					PaidSpent:        paidSpent,
					UnpaidSpent:      unpaidSpent,
					PaidPercentage:   paidPercentage,
					UnpaidPercentage: unpaidPercentage,
					IsNearBudget:     isNearBudget,
					IsOverBudget:     isOverBudget,
					OverBudgetAmount: overBudgetAmount,
					RemainingBudget:  remainingBudget,
				})
			}
		}

		groupViews = append(groupViews, GroupView{
			ID:          grp.ID,
			Name:        grp.Name,
			Description: grp.Description,
			Order:       grp.Order,
			Categories:  categoryViews,
		})
	}

	status := budgetStatus(totalBudgeted, totalIncome)
	displayBudget := totalBudgeted - paidExpensesTotal
	displayBudget = math.Round(displayBudget*100) / 100

	return DashboardView{
		TotalIncome:         totalIncome,
		TotalExpenses:       totalExpenses,
		TotalBudgeted:       displayBudget,
		TotalBudgetedStatus: status,
		Currency:            p.Currency,
		Groups:              groupViews,
	}
}
