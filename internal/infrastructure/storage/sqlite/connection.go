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
		return nil, err
	}

	// Important: SQLite supports only one writer at a time.
	// Setting MaxOpenConns to 1 avoids "database is locked" errors during concurrent writes.
	db.SetMaxOpenConns(1)

	if err = db.PingContext(ctx); err != nil {
		if db != nil {
			_ = db.Close()
		}
		return nil, err
	}

	// Enable database optimization
	_, err = db.ExecContext(ctx, "PRAGMA synchronous=NORMAL;")
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	_, err = db.ExecContext(ctx, "PRAGMA temp_store = memory;")
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	// Note: 30GB mmap limit is an upper bound, not an immediate allocation.
	_, err = db.ExecContext(ctx, "PRAGMA mmap_size=30000000000;")
	if err != nil {
		_ = db.Close()
		return nil, err
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
