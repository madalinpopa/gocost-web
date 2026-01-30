package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/madalinpopa/gocost-web/internal/domain/identity"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
)

type SQLiteUserRepository struct {
	db DBExecutor
}

func NewSQLiteUserRepository(db DBExecutor) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (r *SQLiteUserRepository) Save(ctx context.Context, user identity.User) error {
	query := `
		INSERT INTO users (id, username, email, password, currency)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			username = excluded.username,
			email = excluded.email,
			password = excluded.password,
			currency = excluded.currency
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID.String(),
		user.Username.Value(),
		user.Email.Value(),
		user.Password.Value(),
		user.Currency.Value(),
	)
	if err != nil {
		if isUniqueConstraintViolation(err) {
			return identity.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (r *SQLiteUserRepository) FindByID(ctx context.Context, id identity.ID) (identity.User, error) {
	query := `SELECT id, username, email, password, currency FROM users WHERE id = ?`

	var idStr, usernameStr, emailStr, passwordStr, currencyStr string

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(&idStr, &usernameStr, &emailStr, &passwordStr, &currencyStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return identity.User{}, identity.ErrUserNotFound
		}
		return identity.User{}, fmt.Errorf("failed to find user by id: %w", err)
	}

	return r.mapToUser(idStr, usernameStr, emailStr, passwordStr, currencyStr)
}

func (r *SQLiteUserRepository) FindByEmail(ctx context.Context, email identity.EmailVO) (identity.User, error) {
	query := `SELECT id, username, email, password, currency FROM users WHERE email = ?`

	var idStr, usernameStr, emailStr, passwordStr, currencyStr string

	err := r.db.QueryRowContext(ctx, query, email.Value()).Scan(&idStr, &usernameStr, &emailStr, &passwordStr, &currencyStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return identity.User{}, identity.ErrUserNotFound
		}
		return identity.User{}, fmt.Errorf("failed to find user by email: %w", err)
	}

	return r.mapToUser(idStr, usernameStr, emailStr, passwordStr, currencyStr)
}

func (r *SQLiteUserRepository) FindByUsername(ctx context.Context, username identity.UsernameVO) (identity.User, error) {
	query := `SELECT id, username, email, password, currency FROM users WHERE username = ?`

	var idStr, usernameStr, emailStr, passwordStr, currencyStr string

	err := r.db.QueryRowContext(ctx, query, username.Value()).Scan(&idStr, &usernameStr, &emailStr, &passwordStr, &currencyStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return identity.User{}, identity.ErrUserNotFound
		}
		return identity.User{}, fmt.Errorf("failed to find user by username: %w", err)
	}

	return r.mapToUser(idStr, usernameStr, emailStr, passwordStr, currencyStr)
}

func (r *SQLiteUserRepository) ExistsByEmail(ctx context.Context, email identity.EmailVO) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email.Value()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists by email: %w", err)
	}

	return exists, nil
}

func (r *SQLiteUserRepository) ExistsByUsername(ctx context.Context, username identity.UsernameVO) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, username.Value()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists by username: %w", err)
	}

	return exists, nil
}

func (r *SQLiteUserRepository) mapToUser(idStr, usernameStr, emailStr, passwordStr, currencyStr string) (identity.User, error) {
	id, err := identifier.ParseID(idStr)
	if err != nil {
		return identity.User{}, err
	}

	username, err := identity.NewUsernameVO(usernameStr)
	if err != nil {
		return identity.User{}, err
	}

	email, err := identity.NewEmailVO(emailStr)
	if err != nil {
		return identity.User{}, err
	}

	password, err := identity.NewPasswordVO(passwordStr)
	if err != nil {
		return identity.User{}, err
	}

	currency, err := identity.NewCurrencyVO(currencyStr)
	if err != nil {
		return identity.User{}, err
	}

	return *identity.NewUser(id, username, email, password, currency), nil
}
