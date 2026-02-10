package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/income"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type SQLiteIncomeRepository struct {
	db DBExecutor
}

func NewSQLiteIncomeRepository(db DBExecutor) *SQLiteIncomeRepository {
	return &SQLiteIncomeRepository{db: db}
}

func (r *SQLiteIncomeRepository) Save(ctx context.Context, i income.Income) error {
	query := `
		INSERT INTO incomes (id, user_id, amount, source, received_at) 
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			user_id = excluded.user_id,
			amount = excluded.amount,
			source = excluded.source,
			received_at = excluded.received_at,
			updated_at = CURRENT_TIMESTAMP
	`

	source := sql.NullString{}
	if i.Source.Value() != "" {
		source = sql.NullString{String: i.Source.Value(), Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		i.ID.String(),
		i.UserID.String(),
		i.Amount.Cents(), // Use Cents()
		source,
		i.ReceivedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save income: %w", err)
	}

	return nil
}

func (r *SQLiteIncomeRepository) FindByID(ctx context.Context, id identifier.ID) (income.Income, error) {
	// Join users to get currency
	query := `
		SELECT i.id, i.user_id, i.amount, i.source, i.received_at, u.currency 
		FROM incomes i
		JOIN users u ON i.user_id = u.id
		WHERE i.id = ?
	`

	var idStr, userIDStr, currencyStr string
	var amountCents int64
	var source sql.NullString
	var receivedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(&idStr, &userIDStr, &amountCents, &source, &receivedAt, &currencyStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return income.Income{}, income.ErrIncomeNotFound
		}
		return income.Income{}, fmt.Errorf("failed to find income by id: %w", err)
	}

	return r.mapToIncome(idStr, userIDStr, amountCents, currencyStr, source, receivedAt)
}

func (r *SQLiteIncomeRepository) FindByUserID(ctx context.Context, userID identifier.ID) ([]income.Income, error) {
	query := `
			SELECT i.id, i.user_id, i.amount, i.source, i.received_at, u.currency 
			FROM incomes i
			JOIN users u ON i.user_id = u.id
			WHERE i.user_id = ? 
			ORDER BY i.received_at DESC
		`

	return r.fetchIncomes(ctx, query, userID.String())
}

func (r *SQLiteIncomeRepository) FindByUserIDAndMonth(ctx context.Context, userID identifier.ID, month string) ([]income.Income, error) {
	start, end, err := monthToDateRange(month)
	if err != nil {
		return nil, fmt.Errorf("failed to parse month: %w", err)
	}

	query := `
			SELECT i.id, i.user_id, i.amount, i.source, i.received_at, u.currency 
			FROM incomes i
			JOIN users u ON i.user_id = u.id
			WHERE i.user_id = ? AND i.received_at >= ? AND i.received_at < ?
			ORDER BY i.received_at DESC
		`

	return r.fetchIncomes(ctx, query, userID.String(), start, end)
}

	rows, err := r.db.QueryContext(ctx, query, userID.String(), start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query incomes by month: %w", err)
	}
	defer rows.Close()

	var incomes []income.Income
	for rows.Next() {
		var idStr, userIDStr, currencyStr string
		var amountCents int64
		var source sql.NullString
		var receivedAt time.Time

		if err := rows.Scan(&idStr, &userIDStr, &amountCents, &source, &receivedAt, &currencyStr); err != nil {
			return nil, fmt.Errorf("failed to scan income row: %w", err)
		}

		inc, err := r.mapToIncome(idStr, userIDStr, amountCents, currencyStr, source, receivedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to map income: %w", err)
		}
		incomes = append(incomes, inc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating incomes: %w", err)
	}

	return incomes, nil
}

func (r *SQLiteIncomeRepository) TotalByUserIDAndMonth(ctx context.Context, userID identifier.ID, month string) (money.Money, error) {
	start, end, err := monthToDateRange(month)
	if err != nil {
		return money.Money{}, fmt.Errorf("failed to parse month: %w", err)
	}

	query := `
		SELECT COALESCE(SUM(i.amount), 0), u.currency
		FROM users u
		LEFT JOIN incomes i ON u.id = i.user_id AND i.received_at >= ? AND i.received_at < ?
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
		return money.Money{}, fmt.Errorf("failed to calculate income total: %w", err)
	}

	return money.New(totalCents, currencyStr)
}

func (r *SQLiteIncomeRepository) Delete(ctx context.Context, id identifier.ID) error {
	query := `DELETE FROM incomes WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete income: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return income.ErrIncomeNotFound
	}
	return nil
}

func (r *SQLiteIncomeRepository) mapToIncome(idStr, userIDStr string, amountCents int64, currencyStr string, source sql.NullString, receivedAt time.Time) (income.Income, error) {
	id, err := identifier.ParseID(idStr)
	if err != nil {
		return income.Income{}, err
	}

	userID, err := identifier.ParseID(userIDStr)
	if err != nil {
		return income.Income{}, err
	}

	amount, err := money.New(amountCents, currencyStr)
	if err != nil {
		return income.Income{}, err
	}

	sourceValue := ""
	if source.Valid {
		sourceValue = source.String
	}
	sourceVO, err := income.NewSourceVO(sourceValue)
	if err != nil {
		return income.Income{}, err
	}

	inc, err := income.NewIncome(id, userID, amount, sourceVO, receivedAt)
	if err != nil {
		return income.Income{}, err
	}

	return *inc, nil
}
