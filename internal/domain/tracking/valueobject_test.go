package tracking

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewNameVO(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		_, err := NewNameVO("")
		assert.ErrorIs(t, err, ErrEmptyName)
	})

	t.Run("name too long", func(t *testing.T) {
		longName := strings.Repeat("a", 101)
		_, err := NewNameVO(longName)
		assert.ErrorIs(t, err, ErrNameTooLong)
	})

	t.Run("valid name", func(t *testing.T) {
		validName := "Groceries"
		name, err := NewNameVO(validName)
		assert.NoError(t, err)
		assert.Equal(t, validName, name.Value())
		assert.Equal(t, validName, name.String())
	})
}

func TestNameVO_Equals(t *testing.T) {
	t.Run("equal names", func(t *testing.T) {
		n1, _ := NewNameVO("Groceries")
		n2, _ := NewNameVO("Groceries")
		assert.True(t, n1.Equals(n2))
	})

	t.Run("unequal names", func(t *testing.T) {
		n1, _ := NewNameVO("Groceries")
		n2, _ := NewNameVO("Utilities")
		assert.False(t, n1.Equals(n2))
	})
}

func TestNewDescriptionVO(t *testing.T) {
	t.Run("description too long", func(t *testing.T) {
		longDesc := strings.Repeat("a", 1001)
		_, err := NewDescriptionVO(longDesc)
		assert.ErrorIs(t, err, ErrDescriptionTooLong)
	})

	t.Run("valid description", func(t *testing.T) {
		validDesc := "Some description"
		desc, err := NewDescriptionVO(validDesc)
		assert.NoError(t, err)
		assert.Equal(t, validDesc, desc.Value())
	})

	t.Run("empty description is valid", func(t *testing.T) {
		desc, err := NewDescriptionVO("")
		assert.NoError(t, err)
		assert.Equal(t, "", desc.Value())
	})
}

func TestDescriptionVO_Equals(t *testing.T) {
	t.Run("equal descriptions", func(t *testing.T) {
		d1, _ := NewDescriptionVO("Test")
		d2, _ := NewDescriptionVO("Test")
		assert.True(t, d1.Equals(d2))
	})

	t.Run("unequal descriptions", func(t *testing.T) {
		d1, _ := NewDescriptionVO("Test")
		d2, _ := NewDescriptionVO("Other")
		assert.False(t, d1.Equals(d2))
	})
}

func TestMonth(t *testing.T) {
	t.Run("creates valid month", func(t *testing.T) {
		month, err := NewMonth(2024, time.January)
		assert.NoError(t, err)
		assert.Equal(t, "2024-01", month.Value())
	})

	t.Run("rejects invalid month", func(t *testing.T) {
		_, err := NewMonth(2024, time.Month(13))
		assert.ErrorIs(t, err, ErrInvalidMonth)
	})

	t.Run("parses valid month", func(t *testing.T) {
		month, err := ParseMonth("2024-11")
		assert.NoError(t, err)
		assert.Equal(t, "2024-11", month.String())
	})

	t.Run("rejects invalid format", func(t *testing.T) {
		_, err := ParseMonth("2024/11")
		assert.ErrorIs(t, err, ErrInvalidMonth)
	})

	t.Run("equal months", func(t *testing.T) {
		m1, _ := ParseMonth("2024-05")
		m2, _ := NewMonth(2024, time.May)
		assert.True(t, m1.Equals(m2))
	})
}

func TestMonth_Before(t *testing.T) {
	t.Run("returns true when month is before another", func(t *testing.T) {
		// Arrange
		earlier, _ := NewMonth(2024, time.January)
		later, _ := NewMonth(2024, time.February)

		// Act
		result := earlier.Before(later)

		// Assert
		assert.True(t, result)
	})

	t.Run("returns false when month is after another", func(t *testing.T) {
		// Arrange
		earlier, _ := NewMonth(2024, time.January)
		later, _ := NewMonth(2024, time.February)

		// Act
		result := later.Before(earlier)

		// Assert
		assert.False(t, result)
	})

	t.Run("returns false when months are equal", func(t *testing.T) {
		// Arrange
		month1, _ := NewMonth(2024, time.March)
		month2, _ := NewMonth(2024, time.March)

		// Act
		result := month1.Before(month2)

		// Assert
		assert.False(t, result)
	})

	t.Run("correctly compares months across years", func(t *testing.T) {
		// Arrange
		earlier, _ := NewMonth(2023, time.December)
		later, _ := NewMonth(2024, time.January)

		// Act
		result := earlier.Before(later)

		// Assert
		assert.True(t, result)
	})
}

func TestNewOrderVO(t *testing.T) {
	t.Run("creates valid order with zero", func(t *testing.T) {
		// Arrange & Act
		order, err := NewOrderVO(0)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0, order.Value())
	})

	t.Run("creates valid order with positive value", func(t *testing.T) {
		// Arrange & Act
		order, err := NewOrderVO(5)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 5, order.Value())
	})

	t.Run("rejects negative order value", func(t *testing.T) {
		// Arrange & Act
		_, err := NewOrderVO(-1)

		// Assert
		assert.ErrorIs(t, err, ErrInvalidOrder)
	})

	t.Run("rejects large negative order value", func(t *testing.T) {
		// Arrange & Act
		_, err := NewOrderVO(-100)

		// Assert
		assert.ErrorIs(t, err, ErrInvalidOrder)
	})
}

func TestOrderVO_Equals(t *testing.T) {
	t.Run("equal orders", func(t *testing.T) {
		// Arrange
		order1, _ := NewOrderVO(5)
		order2, _ := NewOrderVO(5)

		// Act
		result := order1.Equals(order2)

		// Assert
		assert.True(t, result)
	})

	t.Run("unequal orders", func(t *testing.T) {
		// Arrange
		order1, _ := NewOrderVO(1)
		order2, _ := NewOrderVO(2)

		// Act
		result := order1.Equals(order2)

		// Assert
		assert.False(t, result)
	})
}
