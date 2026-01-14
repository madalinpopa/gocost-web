package usecase

import (
	"log/slog"

	"github.com/madalinpopa/gocost-web/internal/infrastructure/storage/sqlite"
	"github.com/madalinpopa/gocost-web/internal/platform/security"
)

type UseCase struct {
	AuthUseCase     AuthUseCase
	IncomeUseCase   IncomeUseCase
	GroupUseCase    GroupUseCase
	CategoryUseCase CategoryUseCase
	ExpenseUseCase  ExpenseUseCase
}

func New(uow *sqlite.SqliteUnitOfWork, logger *slog.Logger) *UseCase {
	// Infra services
	passwordHasher := security.NewPasswordHasher()

	// Use cases
	authUseCase := NewAuthUseCase(uow, logger, passwordHasher)
	incomeUseCase := NewIncomeUseCase(uow, logger)
	groupUseCase := NewGroupUseCase(uow, logger)
	categoryUseCase := NewCategoryUseCase(uow, logger)
	expenseUseCase := NewExpenseUseCase(uow, logger)

	return &UseCase{
		AuthUseCase:     authUseCase,
		IncomeUseCase:   incomeUseCase,
		GroupUseCase:    groupUseCase,
		CategoryUseCase: categoryUseCase,
		ExpenseUseCase:  expenseUseCase,
	}
}
