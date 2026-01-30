package expense

import (
	"context"

	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type ExpenseRepository interface {
	Save(ctx context.Context, expense Expense) error
	FindByID(ctx context.Context, id ID) (Expense, error)
	FindByUserID(ctx context.Context, userID ID) ([]Expense, error)
	FindByUserIDAndMonth(ctx context.Context, userID ID, month string) ([]Expense, error)
	TotalsByCategoryAndMonth(ctx context.Context, userID ID, month string) ([]CategoryTotals, error)
	ReassignCategoryFromMonth(ctx context.Context, fromCategoryID ID, toCategoryID ID, month string) error
	Delete(ctx context.Context, id ID) error
	Total(ctx context.Context, userID ID, month string) (money.Money, error)
}

type CategoryTotals struct {
	CategoryID ID
	Total      money.Money
	PaidTotal  money.Money
}
