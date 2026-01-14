package money_test

import (
	"testing"

	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/stretchr/testify/assert"
)

func TestNewMoney(t *testing.T) {
	t.Run("should create Money with valid cents", func(t *testing.T) {
		// Expected
		expectedCents := int64(1500)

		// Act
		m, err := money.New(1500)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCents, m.Cents())
	})

	t.Run("should return error for negative cents", func(t *testing.T) {
		// Act
		_, err := money.New(-500)

		// Assert
		assert.ErrorIs(t, err, money.ErrNegativeAmount)
	})

	t.Run("should be fine with zero cents", func(t *testing.T) {
		// Act
		m, err := money.New(0)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int64(0), m.Cents())
	})

	t.Run("should create Money from float amount", func(t *testing.T) {
		// Expected
		expectedCents := int64(2500)

		// Act
		m, err := money.NewFromFloat(25.00)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCents, m.Cents())
	})

	t.Run("should return error for negative float amount", func(t *testing.T) {
		// Act
		_, err := money.NewFromFloat(-10.50)

		// Assert
		assert.ErrorIs(t, err, money.ErrNegativeAmount)
	})
}
