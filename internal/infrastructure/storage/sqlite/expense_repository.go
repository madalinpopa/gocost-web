package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type SQLiteExpenseRepository struct {
	db DBExecutor
}

func NewSQLiteExpenseRepository(db DBExecutor) *SQLiteExpenseRepository {
	return &SQLiteExpenseRepository{db: db}
}

func (r *SQLiteExpenseRepository) Save(ctx context.Context, e expense.Expense) error {
	query := `
		INSERT INTO expenses (id, category_id, amount, description, spent_at, is_paid, paid_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			category_id = excluded.category_id,
			amount = excluded.amount,
			description = excluded.description,
			spent_at = excluded.spent_at,
			is_paid = excluded.is_paid,
			paid_at = excluded.paid_at,
			updated_at = CURRENT_TIMESTAMP
	`

	paidAt := sql.NullTime{}
	if paidAtValue := e.Payment.PaidAt(); paidAtValue != nil {
		paidAt = sql.NullTime{Time: *paidAtValue, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		e.ID.String(),
		e.CategoryID.String(),
		e.Amount.Amount(),
		e.Description.Value(),
		e.SpentAt,
		e.Payment.IsPaid(),
		paidAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteExpenseRepository) FindByID(ctx context.Context, id identifier.ID) (expense.Expense, error) {
	query := `SELECT id, category_id, amount, description, spent_at, is_paid, paid_at FROM expenses WHERE id = ?`

	var idStr, categoryIDStr, descriptionStr string
	var amountFloat float64
	var spentAt time.Time
	var isPaidInt int
	var paidAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(&idStr, &categoryIDStr, &amountFloat, &descriptionStr, &spentAt, &isPaidInt, &paidAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return expense.Expense{}, expense.ErrExpenseNotFound
		}
		return expense.Expense{}, err
	}

	return r.mapToExpense(idStr, categoryIDStr, amountFloat, descriptionStr, spentAt, isPaidInt == 1, paidAt)
}

func (r *SQLiteExpenseRepository) FindByUserID(ctx context.Context, userID identifier.ID) ([]expense.Expense, error) {
	query := `
		SELECT e.id, e.category_id, e.amount, e.description, e.spent_at, e.is_paid, e.paid_at 
		FROM expenses e 
		JOIN categories c ON e.category_id = c.id 
		JOIN groups g ON c.group_id = g.id 
		WHERE g.user_id = ? 
		ORDER BY e.spent_at DESC
	`

	return r.fetchExpenses(ctx, query, userID.String())
}

func (r *SQLiteExpenseRepository) FindByUserIDAndMonth(ctx context.Context, userID identifier.ID, month string) ([]expense.Expense, error) {
	start, end, err := monthToDateRange(month)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT e.id, e.category_id, e.amount, e.description, e.spent_at, e.is_paid, e.paid_at 
		FROM expenses e 
		JOIN categories c ON e.category_id = c.id 
		JOIN groups g ON c.group_id = g.id 
		WHERE g.user_id = ? AND e.spent_at >= ? AND e.spent_at < ?
		ORDER BY e.spent_at DESC
	`

	return r.fetchExpenses(ctx, query, userID.String(), start, end)
}

func (r *SQLiteExpenseRepository) fetchExpenses(ctx context.Context, query string, args ...any) ([]expense.Expense, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []expense.Expense
	for rows.Next() {
		var idStr, categoryIDStr, descriptionStr string
		var amountFloat float64
		var spentAt time.Time
		var isPaidInt int
		var paidAt sql.NullTime

		if err := rows.Scan(&idStr, &categoryIDStr, &amountFloat, &descriptionStr, &spentAt, &isPaidInt, &paidAt); err != nil {
			return nil, err
		}

		exp, err := r.mapToExpense(idStr, categoryIDStr, amountFloat, descriptionStr, spentAt, isPaidInt == 1, paidAt)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, exp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return expenses, nil
}

func (r *SQLiteExpenseRepository) Total(ctx context.Context, userID identifier.ID, month string) (float64, error) {
	start, end, err := monthToDateRange(month)
	if err != nil {
		return 0, err
	}

	query := `
		SELECT COALESCE(SUM(e.amount), 0)
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		JOIN groups g ON c.group_id = g.id
		WHERE g.user_id = ? AND e.spent_at >= ? AND e.spent_at < ?
	`
	var total float64
	err = r.db.QueryRowContext(ctx, query, userID.String(), start, end).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *SQLiteExpenseRepository) Delete(ctx context.Context, id identifier.ID) error {
	query := `DELETE FROM expenses WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return expense.ErrExpenseNotFound
	}
	return nil
}

func (r *SQLiteExpenseRepository) mapToExpense(idStr, categoryIDStr string, amountFloat float64, descriptionStr string, spentAt time.Time, isPaid bool, paidAt sql.NullTime) (expense.Expense, error) {
	id, err := identifier.ParseID(idStr)
	if err != nil {
		return expense.Expense{}, err
	}

	categoryID, err := identifier.ParseID(categoryIDStr)
	if err != nil {
		return expense.Expense{}, err
	}

	amount, err := money.NewFromFloat(amountFloat)
	if err != nil {
		return expense.Expense{}, err
	}

	description, err := expense.NewExpenseDescriptionVO(descriptionStr)
	if err != nil {
		return expense.Expense{}, err
	}

	var paidAtValue *time.Time
	if paidAt.Valid {
		paidAtValue = &paidAt.Time
	}

	paymentStatus, err := expense.NewPaymentStatus(isPaid, paidAtValue)
	if err != nil {
		return expense.Expense{}, err
	}

	exp, err := expense.NewExpense(id, categoryID, amount, description, spentAt, paymentStatus)
	if err != nil {
		return expense.Expense{}, err
	}

	return *exp, nil
}

func monthToDateRange(month string) (start, end time.Time, err error) {
	start, err = time.Parse("2006-01", month)
	if err != nil {
		return
	}
	end = start.AddDate(0, 1, 0)
	return
}
