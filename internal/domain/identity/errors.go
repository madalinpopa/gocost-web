package identity

import "errors"

var (
	ErrInvalidEmailFormat = errors.New("invalid email format")
	ErrEmptyEmail         = errors.New("email cannot be empty")
	ErrEmailTooLong       = errors.New("email exceeds maximum length of 452 characters")
	ErrEmailTooShort      = errors.New("email must be at least 5 characters long")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters long")
	ErrEmptyUsername      = errors.New("username cannot be empty")
	ErrUsernameTooLong    = errors.New("username exceeds maximum length of 30 characters")
	ErrUsernameTooShort   = errors.New("username must be at least 3 characters long")
	ErrEmptyPassword      = errors.New("password cannot be empty")
	ErrInvalidHash        = errors.New("invalid password hash")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCurrency    = errors.New("invalid currency code")
	ErrEmptyCurrency      = errors.New("currency cannot be empty")
)
