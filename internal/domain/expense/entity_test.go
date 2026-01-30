package expense

import (
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/stretchr/testify/assert"
)

func TestNewExpense(t *testing.T) {
	t.Run("creates valid expense", func(t *testing.T) {
		// Arrange
		id, _ := identifier.NewID()
		categoryID, _ := identifier.NewID()
		amount, _ := money.New(5000, "USD")
		description, _ := NewExpenseDescriptionVO("Lunch")
		spentAt := time.Now()
		payment := NewUnpaidStatus()

		// Act
		expense, err := NewExpense(id, categoryID, amount, description, spentAt, payment)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, expense)
		assert.Equal(t, id, expense.ID)
		assert.Equal(t, categoryID, expense.CategoryID)
		assert.Equal(t, amount, expense.Amount)
		assert.Equal(t, description, expense.Description)
		assert.Equal(t, spentAt, expense.SpentAt)
		assert.Equal(t, payment, expense.Payment)
	})

	t.Run("invalid amount - zero", func(t *testing.T) {
		// Arrange
		id, _ := identifier.NewID()
		categoryID, _ := identifier.NewID()
		amount, _ := money.New(0, "USD")
		description, _ := NewExpenseDescriptionVO("Lunch")
		spentAt := time.Now()
		payment := NewUnpaidStatus()

		// Act
		expense, err := NewExpense(id, categoryID, amount, description, spentAt, payment)

		// Assert
		assert.ErrorIs(t, err, ErrInvalidAmount)
		assert.Nil(t, expense)
	})

	t.Run("invalid amount - negative", func(t *testing.T) {
		// Arrange
		id, _ := identifier.NewID()
		categoryID, _ := identifier.NewID()
		description, _ := NewExpenseDescriptionVO("Lunch")
		spentAt := time.Now()
		payment := NewUnpaidStatus()
		amount, _ := money.New(-100, "USD")
		// Act
		expense, err := NewExpense(id, categoryID, amount, description, spentAt, payment)
		// Assert
		assert.ErrorIs(t, err, ErrInvalidAmount)
		assert.Nil(t, expense)
	})
}