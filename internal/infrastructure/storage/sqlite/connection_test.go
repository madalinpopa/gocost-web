package sqlite_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/infrastructure/storage/sqlite"
	"github.com/stretchr/testify/assert"
)

// TestNewDatabaseConnection validates the behavior of NewDatabaseConnection with various DSNs and error scenarios.
func TestNewDatabaseConnection(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		// Arrange
		tmpDir, err := os.MkdirTemp("", "test-database-*")
		assert.NoError(t, err)
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()

		dsn := filepath.Join(tmpDir, "test.db")

		// Act
		db, err := sqlite.NewDatabaseConnection(context.Background(), dsn)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, db)

		if db != nil {
			err = db.Ping()
			assert.NoError(t, err)
			_ = db.Close()
		}
	})

	t.Run("empty dsn", func(t *testing.T) {
		// Arrange
		dsn := ""

		// Act
		db, err := sqlite.NewDatabaseConnection(context.Background(), dsn)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.EqualError(t, err, "empty dsn")
	})
}
