package expense

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

func TestNewExpenseDescriptionVO(t *testing.T) {
	t.Run("description too long", func(t *testing.T) {
		longDesc := strings.Repeat("a", 256)
		_, err := NewExpenseDescriptionVO(longDesc)
		assert.ErrorIs(t, err, ErrExpenseDescriptionTooLong)
	})

	t.Run("valid description", func(t *testing.T) {
		validDesc := "Weekly groceries"
		desc, err := NewExpenseDescriptionVO(validDesc)
		assert.NoError(t, err)
		assert.Equal(t, validDesc, desc.Value())
	})

	t.Run("empty description is valid", func(t *testing.T) {
		desc, err := NewExpenseDescriptionVO("")
		assert.NoError(t, err)
		assert.Equal(t, "", desc.Value())
	})
}

func TestExpenseDescriptionVO_Equals(t *testing.T) {
	t.Run("equal descriptions", func(t *testing.T) {
		d1, _ := NewExpenseDescriptionVO("Test")
		d2, _ := NewExpenseDescriptionVO("Test")
		assert.True(t, d1.Equals(d2))
	})

	t.Run("unequal descriptions", func(t *testing.T) {
		d1, _ := NewExpenseDescriptionVO("Test")
		d2, _ := NewExpenseDescriptionVO("Other")
		assert.False(t, d1.Equals(d2))
	})
}

func TestPaymentStatus(t *testing.T) {
	t.Run("unpaid without paid_at is valid", func(t *testing.T) {
		status, err := NewPaymentStatus(false, nil)
		assert.NoError(t, err)
		assert.False(t, status.IsPaid())
		assert.Nil(t, status.PaidAt())
	})

	t.Run("paid with paid_at is valid", func(t *testing.T) {
		paidAt := time.Now()
		status, err := NewPaymentStatus(true, &paidAt)
		assert.NoError(t, err)
		assert.True(t, status.IsPaid())
		assert.WithinDuration(t, paidAt, *status.PaidAt(), time.Second)
	})

	t.Run("paid without paid_at is invalid", func(t *testing.T) {
		_, err := NewPaymentStatus(true, nil)
		assert.ErrorIs(t, err, ErrPaidAtRequired)
	})

	t.Run("unpaid with paid_at is invalid", func(t *testing.T) {
		paidAt := time.Now()
		_, err := NewPaymentStatus(false, &paidAt)
		assert.ErrorIs(t, err, ErrPaidAtNotAllowed)
	})
}
