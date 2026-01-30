package money

import (
	"errors"
	"strings"

	"github.com/Rhymond/go-money"
)

// ErrInvalidCurrency indicates that the provided currency code is invalid
var ErrInvalidCurrency = errors.New("invalid currency code")

type Money struct {
	m *money.Money
}

// New creates a new Money instance from cents and currency code
func New(cents int64, currency string) (Money, error) {
	if money.GetCurrency(currency) == nil {
		return Money{}, ErrInvalidCurrency
	}

	return Money{m: money.New(cents, currency)}, nil
}

// NewFromFloat creates a new Money instance from a float amount and currency code
func NewFromFloat(amount float64, currency string) (Money, error) {
	if money.GetCurrency(currency) == nil {
		return Money{}, ErrInvalidCurrency
	}

	return Money{m: money.NewFromFloat(amount, currency)}, nil
}

func (m Money) Cents() int64 {
	if m.m == nil {
		return 0
	}
	return m.m.Amount()
}

func (m Money) Amount() float64 {
	if m.m == nil {
		return 0.0
	}
	return m.m.AsMajorUnits()
}

func (m Money) Currency() string {
	if m.m == nil {
		return ""
	}
	return m.m.Currency().Code
}

func (m Money) Display() string {
	if m.m == nil {
		return "N/A"
	}

	currency := m.m.Currency()
	if currency == nil {
		return m.m.Display()
	}

	amount := m.m.Amount()
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	formatter := money.NewFormatter(currency.Fraction, currency.Decimal, currency.Thousand, "", "1")
	formattedAmount := formatter.Format(amount)

	grapheme := currency.Grapheme
	if grapheme == "" {
		grapheme = currency.Code
	}
	if grapheme == "" {
		return sign + formattedAmount
	}

	template := currency.Template
	amountIndex := strings.Index(template, "1")
	graphemeIndex := strings.Index(template, "$")
	if amountIndex == -1 || graphemeIndex == -1 {
		return sign + grapheme + " " + formattedAmount
	}

	if graphemeIndex < amountIndex {
		return sign + grapheme + " " + formattedAmount
	}

	return sign + formattedAmount + " " + grapheme
}

func (m Money) String() string {
	if m.m == nil {
		return "Money<nil>"
	}
	return m.m.Display()
}

func (m Money) Equals(other Money) (bool, error) {
	if m.m == nil || other.m == nil {
		return false, errors.New("uninitialized money")
	}
	return m.m.Equals(other.m)
}

func (m Money) IsZero() (bool, error) {
	if m.m == nil {
		return false, errors.New("uninitialized money")
	}
	return m.m.IsZero(), nil
}

func (m Money) IsPositive() (bool, error) {
	if m.m == nil {
		return false, errors.New("uninitialized money")
	}
	return m.m.IsPositive(), nil
}

func (m Money) IsNegative() (bool, error) {
	if m.m == nil {
		return false, errors.New("uninitialized money")
	}
	return m.m.IsNegative(), nil
}

func (m Money) Add(other Money) (Money, error) {
	if m.m == nil || other.m == nil {
		return Money{}, errors.New("uninitialized money")
	}
	res, err := m.m.Add(other.m)
	if err != nil {
		return Money{}, err
	}
	return Money{m: res}, nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.m == nil || other.m == nil {
		return Money{}, errors.New("uninitialized money")
	}
	res, err := m.m.Subtract(other.m)
	if err != nil {
		return Money{}, err
	}
	return Money{m: res}, nil
}

func (m Money) Multiply(factor int64) Money {
	if m.m == nil {
		return Money{}
	}
	return Money{m: m.m.Multiply(factor)}
}

func (m Money) GreaterThan(other Money) (bool, error) {
	if m.m == nil || other.m == nil {
		return false, errors.New("uninitialized money")
	}
	return m.m.GreaterThan(other.m)
}

func (m Money) LessThan(other Money) (bool, error) {
	if m.m == nil || other.m == nil {
		return false, errors.New("uninitialized money")
	}
	return m.m.LessThan(other.m)
}

func (m Money) GreaterThanOrEqual(other Money) (bool, error) {
	if m.m == nil || other.m == nil {
		return false, errors.New("uninitialized money")
	}
	return m.m.GreaterThanOrEqual(other.m)
}

func (m Money) LessThanOrEqual(other Money) (bool, error) {
	if m.m == nil || other.m == nil {
		return false, errors.New("uninitialized money")
	}
	return m.m.LessThanOrEqual(other.m)
}
