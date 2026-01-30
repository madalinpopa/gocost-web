package views

import (
	"errors"
	"fmt"

	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/madalinpopa/gocost-web/internal/usecase"
)

type DashboardPresenter struct {
	Currency string
	zero     money.Money
}

const dateLayout = "2006-01-02"

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

func NewDashboardPresenter(currency string) (*DashboardPresenter, error) {
	zero, err := money.New(0, currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	return &DashboardPresenter{
		Currency: currency,
		zero:     zero,
	}, nil
}

func (p *DashboardPresenter) Present(data *usecase.DashboardResponse) (DashboardView, error) {
	if data == nil {
		return DashboardView{}, errors.New("dashboard data cannot be nil")
	}

	totalIncome, err := p.moneyFromCents(data.TotalIncomeCents)
	if err != nil {
		return DashboardView{}, err
	}

	totalExpenses, err := p.moneyFromCents(data.TotalExpensesCents)
	if err != nil {
		return DashboardView{}, err
	}

	totalBudgeted, err := p.moneyFromCents(data.TotalBudgetedCents)
	if err != nil {
		return DashboardView{}, err
	}

	paidExpensesTotal, err := p.moneyFromCents(data.PaidExpensesCents)
	if err != nil {
		return DashboardView{}, err
	}

	displayBudget, err := totalBudgeted.Subtract(paidExpensesTotal)
	if err != nil {
		return DashboardView{}, err
	}

	status := budgetStatus(totalBudgeted, totalIncome)
	isTotalBudgetedNegative, _ := displayBudget.IsNegative()

	groupViews := make([]GroupView, 0, len(data.Groups))
	for _, grp := range data.Groups {
		categoryViews := make([]CategoryView, 0, len(grp.Categories))
		for _, cat := range grp.Categories {
			catBudget, err := p.moneyFromCents(cat.BudgetCents)
			if err != nil {
				return DashboardView{}, err
			}

			spent, err := p.moneyFromCents(cat.SpentCents)
			if err != nil {
				return DashboardView{}, err
			}

			paidSpent, err := p.moneyFromCents(cat.PaidSpentCents)
			if err != nil {
				return DashboardView{}, err
			}

			unpaidSpent, err := spent.Subtract(paidSpent)
			if err != nil {
				return DashboardView{}, err
			}

			usagePercentage := budgetUsagePercentage(catBudget, spent)
			paidPercentage, unpaidPercentage := budgetSplitPercentages(catBudget, paidSpent, unpaidSpent)

			isOverBudget, _ := spent.GreaterThan(catBudget)
			isNearBudget := !isOverBudget && usagePercentage > 85

			overBudgetAmount := p.zero
			remainingBudget := p.zero
			if isOverBudget {
				overBudgetAmount, err = spent.Subtract(catBudget)
			} else {
				remainingBudget, err = catBudget.Subtract(spent)
			}
			if err != nil {
				return DashboardView{}, err
			}

			expenseViews, err := p.mapExpenseViews(cat.Expenses)
			if err != nil {
				return DashboardView{}, err
			}

			isBudgetPositive, _ := catBudget.IsPositive()

			categoryViews = append(categoryViews, CategoryView{
				ID:               cat.ID,
				Name:             cat.Name,
				Type:             categoryType(cat.IsRecurrent),
				Description:      cat.Description,
				StartMonth:       cat.StartMonth,
				EndMonth:         cat.EndMonth,
				Budget:           catBudget,
				IsBudgetPositive: isBudgetPositive,
				Spent:            spent,
				Currency:         p.Currency,
				Expenses:         expenseViews,
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

		groupViews = append(groupViews, GroupView{
			ID:          grp.ID,
			Name:        grp.Name,
			Description: grp.Description,
			Order:       grp.Order,
			Categories:  categoryViews,
		})
	}

	return DashboardView{
		TotalIncome:             totalIncome,
		TotalExpenses:           totalExpenses,
		TotalBudgeted:           displayBudget,
		TotalBudgetedStatus:     status,
		IsTotalBudgetedNegative: isTotalBudgetedNegative,
		Currency:                p.Currency,
		Groups:                  groupViews,
	}, nil
}

func (p *DashboardPresenter) mapExpenseViews(expenses []*usecase.ExpenseResponse) ([]ExpenseView, error) {
	views := make([]ExpenseView, 0, len(expenses))
	for _, exp := range expenses {
		if exp == nil {
			continue
		}

		currency := exp.Currency
		if currency == "" {
			currency = p.Currency
		}

		expAmount, err := money.New(exp.AmountCents, currency)
		if err != nil {
			return nil, err
		}

		status := StatusUnpaid
		paidAt := ""
		if exp.IsPaid {
			status = StatusPaid
			if exp.PaidAt != nil {
				paidAt = exp.PaidAt.Format(dateLayout)
			}
		}

		views = append(views, ExpenseView{
			ID:          exp.ID,
			Amount:      expAmount,
			Currency:    currency,
			Description: exp.Description,
			Status:      status,
			SpentAt:     exp.SpentAt.Format(dateLayout),
			PaidAt:      paidAt,
		})
	}

	return views, nil
}

func categoryType(isRecurrent bool) CategoryType {
	if isRecurrent {
		return TypeRecurrent
	}
	return TypeMonthly
}

func budgetUsagePercentage(budget, spent money.Money) float64 {
	budgetAmount := budget.Amount()
	if budgetAmount <= 0 {
		return 0
	}

	percentage := (spent.Amount() / budgetAmount) * 100
	if percentage > 100 {
		return 100
	}
	if percentage < 0 {
		return 0
	}
	return percentage
}

func budgetSplitPercentages(budget, paidSpent, unpaidSpent money.Money) (float64, float64) {
	budgetAmount := budget.Amount()
	if budgetAmount <= 0 {
		return 0, 0
	}

	paidPercentage := (paidSpent.Amount() / budgetAmount) * 100
	if paidPercentage > 100 {
		paidPercentage = 100
	}

	unpaidPercentage := (unpaidSpent.Amount() / budgetAmount) * 100
	if paidPercentage+unpaidPercentage > 100 {
		unpaidPercentage = 100 - paidPercentage
	}
	if unpaidPercentage < 0 {
		unpaidPercentage = 0
	}

	return paidPercentage, unpaidPercentage
}

func (p *DashboardPresenter) moneyFromCents(cents int64) (money.Money, error) {
	return money.New(cents, p.Currency)
}
