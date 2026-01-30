package income

import (
	"context"

	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type IncomeRepository interface {
	Save(ctx context.Context, income Income) error
	FindByID(ctx context.Context, id ID) (Income, error)
	FindByUserID(ctx context.Context, userID ID) ([]Income, error)
	FindByUserIDAndMonth(ctx context.Context, userID ID, month string) ([]Income, error)
	TotalByUserIDAndMonth(ctx context.Context, userID ID, month string) (money.Money, error)
	Delete(ctx context.Context, id ID) error
}
