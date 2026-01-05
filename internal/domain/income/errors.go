package income

import "errors"

var (
	ErrSourceTooLong  = errors.New("source exceeds maximum length of 255 characters")
	ErrInvalidAmount  = errors.New("amount must be positive")
	ErrIncomeNotFound = errors.New("income not found")
)
