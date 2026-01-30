package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/madalinpopa/gocost-web/internal/domain"
	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/domain/identity"
	"github.com/madalinpopa/gocost-web/internal/domain/income"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
)

// SqliteUnitOfWork implements uow.UnitOfWork for SQLite.
type SqliteUnitOfWork struct {
	db *sql.DB
	tx *sql.Tx
}

// NewUnitOfWork creates a new instance of SqliteUnitOfWork.
func NewUnitOfWork(db *sql.DB) *SqliteUnitOfWork {
	return &SqliteUnitOfWork{db: db}
}

// Ensure SqliteUnitOfWork implements uow.UnitOfWork
var _ domain.UnitOfWork = (*SqliteUnitOfWork)(nil)

func (u *SqliteUnitOfWork) UserRepository() identity.UserRepository {
	if u.tx != nil {
		return NewSQLiteUserRepository(u.tx)
	}
	return NewSQLiteUserRepository(u.db)
}

func (u *SqliteUnitOfWork) IncomeRepository() income.IncomeRepository {
	if u.tx != nil {
		return NewSQLiteIncomeRepository(u.tx)
	}
	return NewSQLiteIncomeRepository(u.db)
}

func (u *SqliteUnitOfWork) ExpenseRepository() expense.ExpenseRepository {
	if u.tx != nil {
		return NewSQLiteExpenseRepository(u.tx)
	}
	return NewSQLiteExpenseRepository(u.db)
}

func (u *SqliteUnitOfWork) TrackingRepository() tracking.GroupRepository {
	if u.tx != nil {
		return NewSQLiteTrackingRepository(u.tx)
	}
	return NewSQLiteTrackingRepository(u.db)
}

func (u *SqliteUnitOfWork) Begin(ctx context.Context) (domain.UnitOfWork, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &SqliteUnitOfWork{
		db: u.db,
		tx: tx,
	}, nil
}

func (u *SqliteUnitOfWork) Commit() error {
	if u.tx == nil {
		return nil
	}
	return u.tx.Commit()
}

func (u *SqliteUnitOfWork) Rollback() error {
	if u.tx == nil {
		return nil
	}
	return u.tx.Rollback()
}
