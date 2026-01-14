package domain

import (
	"context"

	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/domain/identity"
	"github.com/madalinpopa/gocost-web/internal/domain/income"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
)

// UnitOfWork defines the contract for a transactional unit of work.
type UnitOfWork interface {
	UserRepository() identity.UserRepository
	IncomeRepository() income.IncomeRepository
	ExpenseRepository() expense.ExpenseRepository
	TrackingRepository() tracking.GroupRepository
	Begin(ctx context.Context) (UnitOfWork, error)
	Commit() error
	Rollback() error
}
