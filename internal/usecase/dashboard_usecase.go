package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/madalinpopa/gocost-web/internal/domain"
	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
)

type DashboardUseCaseImpl struct {
	uow    domain.UnitOfWork
	logger *slog.Logger
}

func NewDashboardUseCase(uow domain.UnitOfWork, logger *slog.Logger) DashboardUseCaseImpl {
	return DashboardUseCaseImpl{
		uow:    uow,
		logger: logger,
	}
}

func (u DashboardUseCaseImpl) Get(ctx context.Context, userID string, month string) (*DashboardResponse, error) {
	if userID == "" {
		return nil, errors.New("user id cannot be empty")
	}
	if month == "" {
		return nil, errors.New("month cannot be empty")
	}

	uID, err := identifier.ParseID(userID)
	if err != nil {
		return nil, err
	}

	trackingRepo := u.uow.TrackingRepository()
	groups, err := trackingRepo.FindByUserIDAndMonth(ctx, uID, month)
	if err != nil {
		return nil, err
	}

	incomeTotal, err := u.uow.IncomeRepository().TotalByUserIDAndMonth(ctx, uID, month)
	if err != nil {
		return nil, err
	}

	expenseTotal, err := u.uow.ExpenseRepository().Total(ctx, uID, month)
	if err != nil {
		return nil, err
	}

	categoryTotals, err := u.uow.ExpenseRepository().TotalsByCategoryAndMonth(ctx, uID, month)
	if err != nil {
		return nil, err
	}

	expenses, err := u.uow.ExpenseRepository().FindByUserIDAndMonth(ctx, uID, month)
	if err != nil {
		return nil, err
	}

	activeCategoryIDs := make(map[string]struct{})
	for _, group := range groups {
		for _, category := range group.Categories {
			activeCategoryIDs[category.ID.String()] = struct{}{}
		}
	}

	expensesByCategory := make(map[string][]*ExpenseResponse)
	for _, exp := range expenses {
		categoryID := exp.CategoryID.String()
		if _, ok := activeCategoryIDs[categoryID]; !ok {
			continue
		}
		expensesByCategory[categoryID] = append(expensesByCategory[categoryID], mapExpenseToResponse(&exp))
	}

	type totals struct {
		spentCents int64
		paidCents  int64
	}
	totalsByCategory := make(map[string]totals, len(categoryTotals))
	var paidExpensesCents int64
	for _, categoryTotal := range categoryTotals {
		categoryID := categoryTotal.CategoryID.String()
		totalsByCategory[categoryID] = totals{
			spentCents: categoryTotal.Total.Cents(),
			paidCents:  categoryTotal.PaidTotal.Cents(),
		}
		paidExpensesCents += categoryTotal.PaidTotal.Cents()
	}

	var totalBudgetedCents int64
	groupResponses := make([]DashboardGroupResponse, 0, len(groups))
	for _, group := range groups {
		categories := make([]DashboardCategoryResponse, 0, len(group.Categories))
		for _, category := range group.Categories {
			categoryID := category.ID.String()
			categoryTotals := totalsByCategory[categoryID]

			budgetCents := category.Budget.Cents()
			totalBudgetedCents += budgetCents

			categories = append(categories, DashboardCategoryResponse{
				ID:             categoryID,
				Name:           category.Name.Value(),
				Description:    category.Description.Value(),
				IsRecurrent:    category.IsRecurrent,
				StartMonth:     category.StartMonth.Value(),
				EndMonth:       category.EndMonth.Value(),
				BudgetCents:    budgetCents,
				SpentCents:     categoryTotals.spentCents,
				PaidSpentCents: categoryTotals.paidCents,
				Expenses:       expensesByCategory[categoryID],
			})
		}

		groupResponses = append(groupResponses, DashboardGroupResponse{
			ID:          group.ID.String(),
			Name:        group.Name.Value(),
			Description: group.Description.Value(),
			Order:       group.Order.Value(),
			Categories:  categories,
		})
	}

	return &DashboardResponse{
		TotalIncomeCents:   incomeTotal.Cents(),
		TotalExpensesCents: expenseTotal.Cents(),
		TotalBudgetedCents: totalBudgetedCents,
		PaidExpensesCents:  paidExpensesCents,
		Groups:             groupResponses,
	}, nil
}

func mapExpenseToResponse(exp *expense.Expense) *ExpenseResponse {
	if exp == nil {
		return nil
	}

	return &ExpenseResponse{
		ID:          exp.ID.String(),
		CategoryID:  exp.CategoryID.String(),
		Amount:      exp.Amount.Amount(),
		Description: exp.Description.Value(),
		SpentAt:     exp.SpentAt,
		IsPaid:      exp.Payment.IsPaid(),
		PaidAt:      exp.Payment.PaidAt(),
	}
}

var _ DashboardUseCase = (*DashboardUseCaseImpl)(nil)
