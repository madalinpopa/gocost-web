package money_test

import (
	"testing"

	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/stretchr/testify/assert"
)

func TestNewMoney(t *testing.T) {
	t.Run("should create Money with valid cents and currency", func(t *testing.T) {
		// Expected
		expectedCents := int64(1500)
		currency := "USD"

		// Act
		m, err := money.New(1500, currency)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCents, m.Cents())
		assert.Equal(t, currency, m.Currency())
	})

	t.Run("should ALLOW negative cents (change from previous behavior)", func(t *testing.T) {
		// Act
		m, err := money.New(-500, "USD")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int64(-500), m.Cents())
		
		isNegative, err := m.IsNegative()
		assert.NoError(t, err)
		assert.True(t, isNegative)
	})

	t.Run("should fail with invalid currency", func(t *testing.T) {
		// Act
		_, err := money.New(100, "INVALID")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, money.ErrInvalidCurrency, err)
	})

	t.Run("should fail with empty currency", func(t *testing.T) {
		// Act
		_, err := money.New(100, "")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, money.ErrInvalidCurrency, err)
	})

	t.Run("should be fine with zero cents", func(t *testing.T) {
		// Act
		m, err := money.New(0, "USD")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int64(0), m.Cents())
	})

	t.Run("should create Money from float amount", func(t *testing.T) {
		// Expected
		expectedCents := int64(2500)

		// Act
		m, err := money.NewFromFloat(25.00, "USD")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCents, m.Cents())
	})

	t.Run("should ALLOW negative float amount", func(t *testing.T) {
		// Act
		m, err := money.NewFromFloat(-10.50, "USD")

		// Assert
		assert.NoError(t, err)

		isNegative, err := m.IsNegative()
		assert.NoError(t, err)
		assert.True(t, isNegative)
	})
}

func TestMoney_Operations(t *testing.T) {
	t.Run("Add same currency", func(t *testing.T) {
		m1, _ := money.New(100, "USD")
		m2, _ := money.New(200, "USD")
		result, err := m1.Add(m2)
		assert.NoError(t, err)
		assert.Equal(t, int64(300), result.Cents())
	})

	t.Run("Add different currency should fail", func(t *testing.T) {
		m1, _ := money.New(100, "USD")
		m2, _ := money.New(200, "EUR")
		_, err := m1.Add(m2)
		assert.Error(t, err)
	})

	t.Run("Subtract same currency", func(t *testing.T) {
		m1, _ := money.New(300, "USD")
		m2, _ := money.New(100, "USD")
		result, err := m1.Subtract(m2)
		assert.NoError(t, err)
		assert.Equal(t, int64(200), result.Cents())
	})

	t.Run("Subtract to negative should NOT error", func(t *testing.T) {
		m1, _ := money.New(100, "USD")
		m2, _ := money.New(200, "USD")
		result, err := m1.Subtract(m2)
		assert.NoError(t, err)
		isNegative, err := result.IsNegative()
		assert.NoError(t, err)
		assert.True(t, isNegative)
	})

	t.Run("Display formatting", func(t *testing.T) {
		m, _ := money.New(12345, "USD")
		assert.Equal(t, "$123.45", m.Display())

		m2, _ := money.New(12345, "EUR")
		assert.Contains(t, m2.Display(), "123.45")
	})

	t.Run("Uninitialized Money Display", func(t *testing.T) {
		var m money.Money
		assert.Equal(t, "N/A", m.Display())
		assert.Equal(t, "Money<nil>", m.String())
	})
}
