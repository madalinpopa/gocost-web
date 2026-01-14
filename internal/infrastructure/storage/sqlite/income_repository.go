package sqlite

import (
	"context"
	"database/sql"
	"errors"
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
		i.Amount.Amount(),
		source,
		i.ReceivedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteIncomeRepository) FindByID(ctx context.Context, id identifier.ID) (income.Income, error) {
	query := `SELECT id, user_id, amount, source, received_at FROM incomes WHERE id = ?`

	var idStr, userIDStr string
	var amountFloat float64
	var source sql.NullString
	var receivedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(&idStr, &userIDStr, &amountFloat, &source, &receivedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return income.Income{}, income.ErrIncomeNotFound
		}
		return income.Income{}, err
	}

	return r.mapToIncome(idStr, userIDStr, amountFloat, source, receivedAt)
}

func (r *SQLiteIncomeRepository) FindByUserID(ctx context.Context, userID identifier.ID) ([]income.Income, error) {
	query := `SELECT id, user_id, amount, source, received_at FROM incomes WHERE user_id = ? ORDER BY received_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incomes []income.Income
	for rows.Next() {
		var idStr, userIDStr string
		var amountFloat float64
		var source sql.NullString
		var receivedAt time.Time

		if err := rows.Scan(&idStr, &userIDStr, &amountFloat, &source, &receivedAt); err != nil {
			return nil, err
		}

		inc, err := r.mapToIncome(idStr, userIDStr, amountFloat, source, receivedAt)
		if err != nil {
			return nil, err
		}
		incomes = append(incomes, inc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return incomes, nil
}

func (r *SQLiteIncomeRepository) Delete(ctx context.Context, id identifier.ID) error {
	query := `DELETE FROM incomes WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return income.ErrIncomeNotFound
	}
	return nil
}

func (r *SQLiteIncomeRepository) mapToIncome(idStr, userIDStr string, amountFloat float64, source sql.NullString, receivedAt time.Time) (income.Income, error) {
	id, err := identifier.ParseID(idStr)
	if err != nil {
		return income.Income{}, err
	}

	userID, err := identifier.ParseID(userIDStr)
	if err != nil {
		return income.Income{}, err
	}

	amount, err := money.NewFromFloat(amountFloat)
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
