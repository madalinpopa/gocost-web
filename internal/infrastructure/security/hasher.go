package security

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordCost   = 12
	maxPasswordLength = 72
)

var (
	ErrPasswordTooLong       = errors.New("password exceeds maximum length of 72 characters")
	ErrPasswordHashingFailed = errors.New("failed to hash password")
	ErrPasswordEmpty         = errors.New("password cannot be empty")
	ErrHashEmpty             = errors.New("hash cannot be empty")
	ErrInvalidHash           = errors.New("invalid bcrypt hash format")
)

type PasswordHasher struct{}

func NewPasswordHasher() PasswordHasher {
	return PasswordHasher{}
}

func (p PasswordHasher) HashPassword(password string) (string, error) {
	// Check if password exceeds maximum length
	if len(password) > maxPasswordLength {
		return "", fmt.Errorf("password exceeds maximum length of %d characters", maxPasswordLength)
	}

	// Hash the password with the minimum cost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), minPasswordCost)
	if err != nil {
		return "", errors.Join(ErrPasswordHashingFailed, err)
	}

	return string(hashedPassword), nil
}

func (p PasswordHasher) CheckPasswordHash(password, hash string) bool {
	// Check if password exceeds maximum length
	if len(password) > maxPasswordLength {
		return false
	}

	// Compare the password with the hash
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (p PasswordHasher) ValidatePassword(password string) error {
	if password == "" {
		return ErrPasswordEmpty
	}

	if len(password) > maxPasswordLength {
		return ErrPasswordTooLong
	}

	return nil
}

func (p PasswordHasher) ValidateHash(hash string) error {
	if hash == "" {
		return ErrHashEmpty
	}

	// Basic validation - bcrypt hashes start with $2a$, $2b$, or $2y$
	if len(hash) < 60 || (hash[:4] != "$2a$" && hash[:4] != "$2b$" && hash[:4] != "$2y$") {
		return ErrInvalidHash
	}

	return nil
}
