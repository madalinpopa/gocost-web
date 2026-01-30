package views

import (
	"time"

	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/madalinpopa/gocost-web/internal/usecase"
)

type DashboardPresenter struct {
	Currency string
}

func budgetStatus(totalBudgeted, balance money.Money) BudgetStatus {
	less, _ := totalBudgeted.LessThan(balance)
	greater, _ := totalBudgeted.GreaterThan(balance)
	switch {
	case less:
		return BudgetStatusUnder
	case greater:
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
	totalIncomeFloat, totalExpensesFloat float64,
	groups []*usecase.GroupResponse,
	expenses []*usecase.ExpenseResponse,
	date time.Time,
) DashboardView {
	monthStr := date.Format("2006-01")

	// Initialize Money objects
	totalIncome, _ := money.NewFromFloat(totalIncomeFloat, p.Currency)
	totalExpenses, _ := money.NewFromFloat(totalExpensesFloat, p.Currency)
	
	// Create a map of expenses by category for easier lookup
	expensesByCategory := make(map[string][]ExpenseView)
	// categorySpent stores Money
	categorySpent := make(map[string]money.Money)
	
	totalBudgeted, _ := money.New(0, p.Currency)
	paidExpensesTotal, _ := money.New(0, p.Currency)

	for _, exp := range expenses {
		status := StatusUnpaid
		paidAt := ""
		expAmount, _ := money.NewFromFloat(exp.Amount, p.Currency)

		if exp.IsPaid {
			status = StatusPaid
			paidExpensesTotal, _ = paidExpensesTotal.Add(expAmount)
			if exp.PaidAt != nil {
				paidAt = exp.PaidAt.Format("2006-01-02")
			}
		}
		
		expensesByCategory[exp.CategoryID] = append(expensesByCategory[exp.CategoryID], ExpenseView{
			ID:          exp.ID,
			Amount:      expAmount,
			Currency:    p.Currency,
			Description: exp.Description,
			Status:      status,
			SpentAt:     exp.SpentAt.Format("2006-01-02"),
			PaidAt:      paidAt,
		})
		
		currentSpent := categorySpent[exp.CategoryID]
		isZero, err := currentSpent.IsZero()
		if err != nil || isZero {
			// Initialize if not present (handling potentially nil map value, though map returns zero value for struct?)
			// money.Money zero value has nil pointer, so methods might fail if not handled.
			// My wrapper IsZero checks nil.
			currentSpent, _ = money.New(0, p.Currency)
		}
		newSpent, _ := currentSpent.Add(expAmount)
		categorySpent[exp.CategoryID] = newSpent
	}

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
				catBudget, _ := money.NewFromFloat(cat.Budget, p.Currency)
				totalBudgeted, _ = totalBudgeted.Add(catBudget)
				
				spent, ok := categorySpent[cat.ID]
				if !ok {
					spent, _ = money.New(0, p.Currency)
				}
				
				catExpenses := expensesByCategory[cat.ID]

				// Calculate progress bar data
				paidSpent, _ := money.New(0, p.Currency)
				for _, exp := range catExpenses {
					if exp.Status == StatusPaid {
						paidSpent, _ = paidSpent.Add(exp.Amount)
					}
				}
				unpaidSpent, _ := spent.Subtract(paidSpent)

				// Percentages
				percentage := 0.0
				paidPercentage := 0.0
				unpaidPercentage := 0.0
				
				catBudgetAmount := catBudget.Amount()
				spentAmount := spent.Amount()
				paidSpentAmount := paidSpent.Amount()
				unpaidSpentAmount := unpaidSpent.Amount()

				if catBudgetAmount > 0 {
					percentage = (spentAmount / catBudgetAmount) * 100
					if percentage > 100 {
						percentage = 100
					}

					paidPercentage = (paidSpentAmount / catBudgetAmount) * 100
					if paidPercentage > 100 {
						paidPercentage = 100
					}

					unpaidPercentage = (unpaidSpentAmount / catBudgetAmount) * 100
					if paidPercentage+unpaidPercentage > 100 {
						unpaidPercentage = 100 - paidPercentage
					}
				}

				// Budget calculations
				isOverBudget, _ := spent.GreaterThan(catBudget)
				isNearBudget := !isOverBudget && percentage > 85

				overBudgetAmount, _ := money.New(0, p.Currency)
				remainingBudget, _ := money.New(0, p.Currency)

				if isOverBudget {
					overBudgetAmount, _ = spent.Subtract(catBudget)
				} else {
					remainingBudget, _ = catBudget.Subtract(spent)
				}

				isBudgetPositive, _ := catBudget.IsPositive()

				categoryViews = append(categoryViews, CategoryView{
					ID:               cat.ID,
					Name:             cat.Name,
					Type:             catType,
					Description:      cat.Description,
					StartMonth:       cat.StartMonth,
					EndMonth:         cat.EndMonth,
					Budget:           catBudget,
					IsBudgetPositive: isBudgetPositive,
					Spent:            spent,
					Currency:         p.Currency,
					Expenses:         catExpenses,
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
	displayBudget, _ := totalBudgeted.Subtract(paidExpensesTotal)

	isTotalBudgetedNegative, _ := displayBudget.IsNegative()

	return DashboardView{
		TotalIncome:             totalIncome,
		TotalExpenses:           totalExpenses,
		TotalBudgeted:           displayBudget,
		TotalBudgetedStatus:     status,
		IsTotalBudgetedNegative: isTotalBudgetedNegative,
		Currency:                p.Currency,
		Groups:                  groupViews,
	}
}