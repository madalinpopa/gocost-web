package income

import (
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/stretchr/testify/assert"
)

func TestNewIncome(t *testing.T) {
	t.Run("creates valid income", func(t *testing.T) {
		// Arrange
		id, _ := identifier.NewID()
		userID, _ := identifier.NewID()
		amount, _ := money.New(1000)
		source, _ := NewSourceVO("Salary")
		receivedAt := time.Now()

		// Act
		income, err := NewIncome(id, userID, amount, source, receivedAt)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, income)
		assert.Equal(t, id, income.ID)
		assert.Equal(t, userID, income.UserID)
		assert.Equal(t, amount, income.Amount)
		assert.Equal(t, source, income.Source)
		assert.Equal(t, receivedAt, income.ReceivedAt)
	})

	t.Run("invalid amount", func(t *testing.T) {
		// Arrange
		id, _ := identifier.NewID()
		userID, _ := identifier.NewID()
		amount, _ := money.New(0) // Zero amount
		source, _ := NewSourceVO("Salary")
		receivedAt := time.Now()

		// Act
		income, err := NewIncome(id, userID, amount, source, receivedAt)

		// Assert
		assert.ErrorIs(t, err, ErrInvalidAmount)
		assert.Nil(t, income)
	})

	t.Run("empty source is allowed", func(t *testing.T) {
		id, _ := identifier.NewID()
		userID, _ := identifier.NewID()
		amount, _ := money.New(1000)
		receivedAt := time.Now()
		source, _ := NewSourceVO("")

		income, err := NewIncome(id, userID, amount, source, receivedAt)
		assert.NoError(t, err)
		assert.Equal(t, "", income.Source.Value())
	})
}
