package expense

import "errors"

var (
	ErrInvalidAmount             = errors.New("amount must be positive")
	ErrInvalidMonth              = errors.New("month must be in YYYY-MM format")
	ErrExpenseNotFound           = errors.New("expense not found")
	ErrExpenseDescriptionTooLong = errors.New("expense description exceeds maximum length of 255 characters")
	ErrPaidAtRequired            = errors.New("paid_at is required when expense is marked as paid")
	ErrPaidAtNotAllowed          = errors.New("paid_at must be empty when expense is not paid")
)
