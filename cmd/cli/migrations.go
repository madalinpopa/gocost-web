package main

import (
	"context"
	"database/sql"

	"github.com/madalinpopa/gocost-web/internal/infrastructure/storage/sqlite"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Open database connection
		logger.Info("connect to database", "dsn", conf.Dsn)
		db, err := sqlite.NewDatabaseConnection(context.Background(), conf.Dsn)
		if err != nil {
			logger.Error("failed to get database connection", "err", err)
			return err
		}

		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {
				logger.Error("Failed to close database", "err", err)
			}
		}(db)

		// Run database migrations
		if err := runMigrations(db); err != nil {
			logger.Error("Failed to run migrations", "err", err)
			return err
		}

		return nil
	},
}

func runMigrations(db *sql.DB) error {
	// Test the connection
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", "err", err)
		return err
	}

	logger.Info("Running migrations....")
	if err := sqlite.MakeMigrations(db); err != nil {
		logger.Error("Failed to run migrations", "err", err)
		return err
	}

	logger.Info("Migrations completed successfully!")
	return nil
}
