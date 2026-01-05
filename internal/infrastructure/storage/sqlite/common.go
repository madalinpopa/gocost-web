package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/mattn/go-sqlite3"
)

// DBExecutor defines the common interface for sql.DB and sql.Tx
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func isUniqueConstraintViolation(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return true
		}
		if sqliteErr.Code == sqlite3.ErrConstraint && strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return true
		}
	}
	return false
}
