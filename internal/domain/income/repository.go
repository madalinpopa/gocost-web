package income

import "context"

type IncomeRepository interface {
	Save(ctx context.Context, income Income) error
	FindByID(ctx context.Context, id ID) (Income, error)
	FindByUserID(ctx context.Context, userID ID) ([]Income, error)
	Delete(ctx context.Context, id ID) error
}
