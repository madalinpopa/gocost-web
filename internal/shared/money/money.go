package money

import (
	"errors"
	"fmt"
)

const (
	// Money precision
	multiplier = 100
)

var (

	// ErrInvalidAmount indicates that the provided money amount is invalid
	ErrInvalidAmount = errors.New("invalid money amount")

	// ErrNegativeAmount indicates that the money amount cannot be negative
	ErrNegativeAmount = errors.New("money amount cannot be negative")
)

type Money struct {
	cents int64
}

func New(cents int64) (Money, error) {
	if cents < 0 {
		return Money{}, ErrNegativeAmount
	}

	return Money{cents: cents}, nil
}

func NewFromFloat(amount float64) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}

	cents := int64(amount * multiplier)
	return Money{cents: cents}, nil
}

func (m Money) Cents() int64 {
	return m.cents
}

func (m Money) Amount() float64 {
	return float64(m.cents) / multiplier
}

func (m Money) String() string {
	return fmt.Sprintf("%.2f", m.Amount())
}

func (m Money) Equals(other Money) bool {
	return m.cents == other.cents
}

func (m Money) IsZero() bool {
	return m.cents == 0
}

func (m Money) IsPositive() bool {
	return m.cents > 0
}

func (m Money) Add(other Money) Money {
	return Money{cents: m.cents + other.cents}
}

func (m Money) Subtract(other Money) (Money, error) {
	result := m.cents - other.cents
	if result < 0 {
		return Money{}, ErrNegativeAmount
	}
	return Money{cents: result}, nil
}

func (m Money) Multiply(factor int64) Money {
	return Money{cents: m.cents * factor}
}

func (m Money) GreaterThan(other Money) bool {
	return m.cents > other.cents
}

func (m Money) LessThan(other Money) bool {
	return m.cents < other.cents
}

func (m Money) GreaterThanOrEqual(other Money) bool {
	return m.cents >= other.cents
}

func (m Money) LessThanOrEqual(other Money) bool {
	return m.cents <= other.cents
}
