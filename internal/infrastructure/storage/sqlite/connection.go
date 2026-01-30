package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/madalinpopa/gocost-web/migrations"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

// sqliteDriver specifies the driver name for connecting to SQLite database
const sqliteDriver = "sqlite3"

// maxOpenConns limits concurrent connections to 1 for SQLite to prevent "database is locked" errors.
// SQLite supports only one writer at a time due to its file-based architecture.
const maxOpenConns = 1

// NewDatabaseConnection opens an SQLite database using provided DSN and verifies the connection.
func NewDatabaseConnection(ctx context.Context, dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("empty dsn")
	}

	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}

	options := fmt.Sprintf("%s%s_busy_timeout=5000&_journal_mode=wal&_foreign_keys=1", dsn, separator)
	db, err := sql.Open(sqliteDriver, options)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)

	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable database optimization
	_, err = db.ExecContext(ctx, "PRAGMA synchronous=NORMAL;")
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set synchronous pragma: %w", err)
	}
	_, err = db.ExecContext(ctx, "PRAGMA temp_store = memory;")
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set temp_store pragma: %w", err)
	}
	// Note: 30GB mmap limit is an upper bound, not an immediate allocation.
	_, err = db.ExecContext(ctx, "PRAGMA mmap_size=30000000000;")
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set mmap_size pragma: %w", err)
	}

	return db, nil
}

func MakeMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrations.MigrationFiles)

	if err := goose.SetDialect(sqliteDriver); err != nil {
		return err
	}

	if err := goose.Up(db, "."); err != nil {
		return err
	}

	return nil
}
