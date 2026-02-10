package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
		e.Amount.Cents(), // Changed from Amount() float to Cents() int64
		e.Description.Value(),
		e.SpentAt,
		e.Payment.IsPaid(),
		paidAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save expense: %w", err)
	}

	return nil
}

func (r *SQLiteExpenseRepository) FindByID(ctx context.Context, id identifier.ID) (expense.Expense, error) {
	// Updated query to join up to users to get currency
	query := `
		SELECT e.id, e.category_id, e.amount, e.description, e.spent_at, e.is_paid, e.paid_at, u.currency
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		JOIN groups g ON c.group_id = g.id
		JOIN users u ON g.user_id = u.id
		WHERE e.id = ?
	`

	var idStr, categoryIDStr, descriptionStr, currencyStr string
	var amountCents int64
	var spentAt time.Time
	var isPaidInt int
	var paidAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&categoryIDStr,
		&amountCents,
		&descriptionStr,
		&spentAt,
		&isPaidInt,
		&paidAt,
		&currencyStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return expense.Expense{}, expense.ErrExpenseNotFound
		}
		return expense.Expense{}, fmt.Errorf("failed to find expense by id: %w", err)
	}

	return r.mapToExpense(
		idStr,
		categoryIDStr,
		amountCents,
		currencyStr,
		descriptionStr,
		spentAt,
		isPaidInt == 1,
		paidAt,
	)
}

func (r *SQLiteExpenseRepository) FindByUserID(ctx context.Context, userID identifier.ID) ([]expense.Expense, error) {
	query := `
		SELECT e.id, e.category_id, e.amount, e.description, e.spent_at, e.is_paid, e.paid_at, u.currency
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		JOIN groups g ON c.group_id = g.id
		JOIN users u ON g.user_id = u.id
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
		SELECT e.id, e.category_id, e.amount, e.description, e.spent_at, e.is_paid, e.paid_at, u.currency
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		JOIN groups g ON c.group_id = g.id
		JOIN users u ON g.user_id = u.id
		WHERE g.user_id = ? AND e.spent_at >= ? AND e.spent_at < ?
		ORDER BY e.spent_at DESC
	`

	return r.fetchExpenses(ctx, query, userID.String(), start, end)
}

func (r *SQLiteExpenseRepository) TotalsByCategoryAndMonth(ctx context.Context, userID identifier.ID, month string) ([]expense.CategoryTotals, error) {
	start, end, err := monthToDateRange(month)
	if err != nil {
		return nil, fmt.Errorf("failed to parse month: %w", err)
	}

	query := `
		SELECT e.category_id,
			COALESCE(SUM(e.amount), 0) AS total_amount,
			COALESCE(SUM(CASE WHEN e.is_paid = 1 THEN e.amount ELSE 0 END), 0) AS paid_amount,
			u.currency
		FROM expenses e
		JOIN categories c ON e.category_id = c.id
		JOIN groups g ON c.group_id = g.id
		JOIN users u ON g.user_id = u.id
		WHERE g.user_id = ? AND e.spent_at >= ? AND e.spent_at < ?
		GROUP BY e.category_id, u.currency
	`

	rows, err := r.db.QueryContext(ctx, query, userID.String(), start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query category totals: %w", err)
	}
	defer rows.Close()

	var totals []expense.CategoryTotals
	for rows.Next() {
		var categoryIDStr, currencyStr string
		var totalCents int64
		var paidCents int64

		if err := rows.Scan(&categoryIDStr, &totalCents, &paidCents, &currencyStr); err != nil {
			return nil, fmt.Errorf("failed to scan category total row: %w", err)
		}

		categoryID, err := identifier.ParseID(categoryIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse category ID: %w", err)
		}

		totalAmount, err := money.New(totalCents, currencyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to create total amount: %w", err)
		}

		paidAmount, err := money.New(paidCents, currencyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to create paid amount: %w", err)
		}

		totals = append(totals, expense.CategoryTotals{
			CategoryID: categoryID,
			Total:      totalAmount,
			PaidTotal:  paidAmount,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category totals: %w", err)
	}

	return totals, nil
}

func (r *SQLiteExpenseRepository) ReassignCategoryFromMonth(ctx context.Context, userID identifier.ID, fromCategoryID identifier.ID, toCategoryID identifier.ID, month string) error {
	start, _, err := monthToDateRange(month)
	if err != nil {
		return fmt.Errorf("failed to parse month: %w", err)
	}

	query := `
			UPDATE expenses
			SET category_id = ?
			WHERE category_id = ?
			  AND spent_at >= ?
			  AND EXISTS (
				SELECT 1
				FROM categories c
				JOIN groups g ON c.group_id = g.id
				WHERE c.id = expenses.category_id AND g.user_id = ?
			  )
			  AND EXISTS (
				SELECT 1
				FROM categories c
				JOIN groups g ON c.group_id = g.id
				WHERE c.id = ? AND g.user_id = ?
			  )
		`

	_, err = r.db.ExecContext(ctx, query, toCategoryID.String(), fromCategoryID.String(), start, userID.String(), toCategoryID.String(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to reassign expenses: %w", err)
	}

	return nil
}

func (r *SQLiteExpenseRepository) fetchExpenses(ctx context.Context, query string, args ...any) ([]expense.Expense, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query expenses: %w", err)
	}
	defer rows.Close()

	var expenses []expense.Expense
	for rows.Next() {
		var idStr, categoryIDStr, descriptionStr, currencyStr string
		var amountCents int64
		var spentAt time.Time
		var isPaidInt int
		var paidAt sql.NullTime

		if err := rows.Scan(&idStr, &categoryIDStr, &amountCents, &descriptionStr, &spentAt, &isPaidInt, &paidAt, &currencyStr); err != nil {
			return nil, fmt.Errorf("failed to scan expense row: %w", err)
		}

		exp, err := r.mapToExpense(idStr, categoryIDStr, amountCents, currencyStr, descriptionStr, spentAt, isPaidInt == 1, paidAt)
		if err != nil {
			return nil, fmt.Errorf("failed to map expense: %w", err)
		}
		expenses = append(expenses, exp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expenses: %w", err)
	}

	return expenses, nil
}

func (r *SQLiteExpenseRepository) Total(ctx context.Context, userID identifier.ID, month string) (money.Money, error) {
	start, end, err := monthToDateRange(month)
	if err != nil {
		return money.Money{}, fmt.Errorf("failed to parse month: %w", err)
	}

	query := `
		SELECT COALESCE(SUM(e.amount), 0), u.currency
		FROM users u
		LEFT JOIN groups g ON u.id = g.user_id
		LEFT JOIN categories c ON g.id = c.group_id
		LEFT JOIN expenses e ON c.id = e.category_id AND e.spent_at >= ? AND e.spent_at < ?
		WHERE u.id = ?
		GROUP BY u.currency
	`
	var totalCents int64
	var currencyStr string
	err = r.db.QueryRowContext(ctx, query, start, end, userID.String()).Scan(&totalCents, &currencyStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return money.Money{}, fmt.Errorf("failed to get user currency: %w", err)
		}
		return money.Money{}, fmt.Errorf("failed to calculate expense total: %w", err)
	}
	return money.New(totalCents, currencyStr)
}

func (r *SQLiteExpenseRepository) Delete(ctx context.Context, id identifier.ID) error {
	query := `DELETE FROM expenses WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return expense.ErrExpenseNotFound
	}
	return nil
}

func (r *SQLiteExpenseRepository) mapToExpense(idStr, categoryIDStr string, amountCents int64, currencyStr string, descriptionStr string, spentAt time.Time, isPaid bool, paidAt sql.NullTime) (expense.Expense, error) {
	id, err := identifier.ParseID(idStr)
	if err != nil {
		return expense.Expense{}, err
	}

	categoryID, err := identifier.ParseID(categoryIDStr)
	if err != nil {
		return expense.Expense{}, err
	}

	amount, err := money.New(amountCents, currencyStr)
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
