package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/domain/identity"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestAuthUseCase(repo *MockUserRepository) AuthUseCaseImpl {
	if repo == nil {
		repo = &MockUserRepository{}
	}

	txUOW := &MockUnitOfWork{UserRepo: repo}
	txUOW.On("Commit").Return(nil)
	txUOW.On("Rollback").Return(nil)

	baseUOW := &MockUnitOfWork{UserRepo: repo}
	baseUOW.On("Begin", mock.Anything).Return(txUOW, nil)

	return NewAuthUseCase(
		baseUOW,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		security.NewPasswordHasher(),
	)
}

func newTestUser(t *testing.T, email, username, hash string) identity.User {
	t.Helper()

	emailVO, err := identity.NewEmailVO(email)
	require.NoError(t, err)

	usernameVO, err := identity.NewUsernameVO(username)
	require.NoError(t, err)

	passwordVO, err := identity.NewPasswordVO(hash)
	require.NoError(t, err)

	id, err := identifier.NewID()
	require.NoError(t, err)

	currencyVO, err := identity.NewCurrencyVO("USD")
	require.NoError(t, err)

	return *identity.NewUser(id, usernameVO, emailVO, passwordVO, currencyVO)
}

func TestAuthUseCase_Register(t *testing.T) {
	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestAuthUseCase(nil)

		resp, err := usecase.Register(context.Background(), nil)

		assert.Nil(t, resp)
		assert.EqualError(t, err, "register request is nil")
	})

	t.Run("returns error for invalid email", func(t *testing.T) {
		usecase := newTestAuthUseCase(nil)

		req := &RegisterUserRequest{
			EmailRequest:    EmailRequest{Email: "invalid-email"},
			UsernameRequest: UsernameRequest{Username: "validuser"},
			Password:        "password1",
			Currency:        "USD",
		}

		resp, err := usecase.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identity.ErrInvalidEmailFormat)
	})

	t.Run("returns error for invalid username", func(t *testing.T) {
		usecase := newTestAuthUseCase(nil)

		req := &RegisterUserRequest{
			EmailRequest:    EmailRequest{Email: "user@example.com"},
			UsernameRequest: UsernameRequest{Username: "ab"},
			Password:        "password1",
			Currency:        "USD",
		}

		resp, err := usecase.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identity.ErrUsernameTooShort)
	})

	t.Run("returns error for empty password", func(t *testing.T) {
		usecase := newTestAuthUseCase(nil)

		req := &RegisterUserRequest{
			EmailRequest:    EmailRequest{Email: "user@example.com"},
			UsernameRequest: UsernameRequest{Username: "validuser"},
			Password:        "",
			Currency:        "USD",
		}

		resp, err := usecase.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, security.ErrPasswordEmpty)
	})

	t.Run("returns error for short password", func(t *testing.T) {
		usecase := newTestAuthUseCase(nil)

		req := &RegisterUserRequest{
			EmailRequest:    EmailRequest{Email: "user@example.com"},
			UsernameRequest: UsernameRequest{Username: "validuser"},
			Password:        "short77",
			Currency:        "USD",
		}

		resp, err := usecase.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identity.ErrPasswordTooShort)
	})

	t.Run("returns error when email already exists", func(t *testing.T) {
		repo := &MockUserRepository{}
		repo.On("ExistsByEmail", mock.Anything, mock.Anything).Return(true, nil)

		usecase := newTestAuthUseCase(repo)

		req := &RegisterUserRequest{
			EmailRequest:    EmailRequest{Email: "user@example.com"},
			UsernameRequest: UsernameRequest{Username: "validuser"},
			Password:        "password1",
			Currency:        "USD",
		}

		resp, err := usecase.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identity.ErrUserAlreadyExists)
	})

	t.Run("returns error when username already exists", func(t *testing.T) {
		repo := &MockUserRepository{}
		repo.On("ExistsByEmail", mock.Anything, mock.Anything).Return(false, nil)
		repo.On("ExistsByUsername", mock.Anything, mock.Anything).Return(true, nil)

		usecase := newTestAuthUseCase(repo)

		req := &RegisterUserRequest{
			EmailRequest:    EmailRequest{Email: "user@example.com"},
			UsernameRequest: UsernameRequest{Username: "validuser"},
			Password:        "password1",
			Currency:        "USD",
		}

		resp, err := usecase.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identity.ErrUserAlreadyExists)
	})

	t.Run("returns error when email lookup fails", func(t *testing.T) {
		expectedErr := errors.New("email lookup failed")
		repo := &MockUserRepository{}
		repo.On("ExistsByEmail", mock.Anything, mock.Anything).Return(false, expectedErr)

		usecase := newTestAuthUseCase(repo)

		req := &RegisterUserRequest{
			EmailRequest:    EmailRequest{Email: "user@example.com"},
			UsernameRequest: UsernameRequest{Username: "validuser"},
			Password:        "password1",
			Currency:        "USD",
		}

		resp, err := usecase.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns error when username lookup fails", func(t *testing.T) {
		expectedErr := errors.New("username lookup failed")
		repo := &MockUserRepository{}
		repo.On("ExistsByEmail", mock.Anything, mock.Anything).Return(false, nil)
		repo.On("ExistsByUsername", mock.Anything, mock.Anything).Return(false, expectedErr)

		usecase := newTestAuthUseCase(repo)

		req := &RegisterUserRequest{
			EmailRequest:    EmailRequest{Email: "user@example.com"},
			UsernameRequest: UsernameRequest{Username: "validuser"},
			Password:        "password1",
			Currency:        "USD",
		}

		resp, err := usecase.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns error when save fails", func(t *testing.T) {
		expectedErr := errors.New("save failed")
		repo := &MockUserRepository{}
		repo.On("ExistsByEmail", mock.Anything, mock.Anything).Return(false, nil)
		repo.On("ExistsByUsername", mock.Anything, mock.Anything).Return(false, nil)
		repo.On("Save", mock.Anything, mock.Anything).Return(expectedErr)

		usecase := newTestAuthUseCase(repo)

		req := &RegisterUserRequest{
			EmailRequest:    EmailRequest{Email: "user@example.com"},
			UsernameRequest: UsernameRequest{Username: "validuser"},
			Password:        "password1",
			Currency:        "USD",
		}

		resp, err := usecase.Register(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("saves user and returns response", func(t *testing.T) {
		var savedUser identity.User
		var saved bool
		repo := &MockUserRepository{}
		repo.On("ExistsByEmail", mock.Anything, mock.Anything).Return(false, nil)
		repo.On("ExistsByUsername", mock.Anything, mock.Anything).Return(false, nil)
		repo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedUser = args.Get(1).(identity.User)
			saved = true
		})

		usecase := newTestAuthUseCase(repo)

		req := &RegisterUserRequest{
			EmailRequest:    EmailRequest{Email: "user@example.com"},
			UsernameRequest: UsernameRequest{Username: "validuser"},
			Password:        "password1",
			Currency:        "USD",
		}

		resp, err := usecase.Register(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, saved)
		assert.Equal(t, req.Email, savedUser.Email.Value())
		assert.Equal(t, req.Username, savedUser.Username.Value())
		assert.NotEqual(t, req.Password, savedUser.Password.Value())
		assert.Equal(t, req.Email, resp.Email)
		assert.Equal(t, req.Username, resp.Username)
		assert.Equal(t, req.Currency, resp.Currency)
		assert.NotEmpty(t, resp.ID)
		parsedID, parseErr := identifier.ParseID(resp.ID)
		require.NoError(t, parseErr)
		assert.True(t, savedUser.ID.Equals(parsedID))
	})
}

func TestAuthUseCase_Login(t *testing.T) {
	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestAuthUseCase(&MockUserRepository{})

		resp, err := usecase.Login(context.Background(), nil)

		assert.Nil(t, resp)
		assert.EqualError(t, err, "login request is nil")
	})

	t.Run("returns error for empty password", func(t *testing.T) {
		usecase := newTestAuthUseCase(&MockUserRepository{})

		req := &LoginRequest{
			EmailOrUsername: "user@example.com",
			Password:        "",
		}

		resp, err := usecase.Login(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, security.ErrPasswordEmpty)
	})

	t.Run("returns error for invalid email and username", func(t *testing.T) {
		usecase := newTestAuthUseCase(&MockUserRepository{})

		req := &LoginRequest{
			EmailOrUsername: "ab",
			Password:        "password1",
		}

		resp, err := usecase.Login(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identity.ErrUsernameTooShort)
	})

	t.Run("returns invalid credentials when email not found", func(t *testing.T) {
		repo := &MockUserRepository{}
		repo.On("FindByEmail", mock.Anything, mock.Anything).Return(identity.User{}, identity.ErrUserNotFound)

		usecase := newTestAuthUseCase(repo)

		req := &LoginRequest{
			EmailOrUsername: "user@example.com",
			Password:        "password1",
		}

		resp, err := usecase.Login(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("returns invalid credentials when username not found", func(t *testing.T) {
		repo := &MockUserRepository{}
		repo.On("FindByUsername", mock.Anything, mock.Anything).Return(identity.User{}, identity.ErrUserNotFound)

		usecase := newTestAuthUseCase(repo)

		req := &LoginRequest{
			EmailOrUsername: "validuser",
			Password:        "password1",
		}

		resp, err := usecase.Login(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("returns error when repository lookup fails", func(t *testing.T) {
		expectedErr := errors.New("lookup failed")
		repo := &MockUserRepository{}
		repo.On("FindByEmail", mock.Anything, mock.Anything).Return(identity.User{}, expectedErr)

		usecase := newTestAuthUseCase(repo)

		req := &LoginRequest{
			EmailOrUsername: "user@example.com",
			Password:        "password1",
		}

		resp, err := usecase.Login(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns error when stored hash is invalid", func(t *testing.T) {
		invalidHash := strings.Repeat("x", 60)
		user := newTestUser(t, "user@example.com", "validuser", invalidHash)
		repo := &MockUserRepository{}
		repo.On("FindByEmail", mock.Anything, mock.Anything).Return(user, nil)

		usecase := newTestAuthUseCase(repo)

		req := &LoginRequest{
			EmailOrUsername: "user@example.com",
			Password:        "password1",
		}

		resp, err := usecase.Login(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, security.ErrInvalidHash)
	})

	t.Run("returns invalid credentials when password does not match", func(t *testing.T) {
		hasher := security.NewPasswordHasher()
		hash, err := hasher.HashPassword("correct-password")
		require.NoError(t, err)

		user := newTestUser(t, "user@example.com", "validuser", hash)
		repo := &MockUserRepository{}
		repo.On("FindByEmail", mock.Anything, mock.Anything).Return(user, nil)

		usecase := newTestAuthUseCase(repo)

		req := &LoginRequest{
			EmailOrUsername: "user@example.com",
			Password:        "wrong-password",
		}

		resp, err := usecase.Login(context.Background(), req)

		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("returns response for email login", func(t *testing.T) {
		hasher := security.NewPasswordHasher()
		hash, err := hasher.HashPassword("password1")
		require.NoError(t, err)

		user := newTestUser(t, "user@example.com", "validuser", hash)
		repo := &MockUserRepository{}
		repo.On("FindByEmail", mock.Anything, mock.Anything).Return(user, nil)

		usecase := newTestAuthUseCase(repo)

		req := &LoginRequest{
			EmailOrUsername: "user@example.com",
			Password:        "password1",
		}

		resp, err := usecase.Login(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, user.ID.String(), resp.UserID)
		assert.Equal(t, user.Email.Value(), resp.Email)
		assert.Equal(t, user.Username.Value(), resp.Username)
		assert.Equal(t, "USD", resp.Currency)
		assert.Empty(t, resp.FullName)
		assert.Empty(t, resp.Role)
	})

	t.Run("returns response for username login", func(t *testing.T) {
		hasher := security.NewPasswordHasher()
		hash, err := hasher.HashPassword("password1")
		require.NoError(t, err)

		user := newTestUser(t, "user@example.com", "validuser", hash)

		repo := &MockUserRepository{}
		repo.On("FindByUsername", mock.Anything, mock.Anything).Return(user, nil)

		usecase := newTestAuthUseCase(repo)

		req := &LoginRequest{
			EmailOrUsername: "validuser",
			Password:        "password1",
		}

		resp, err := usecase.Login(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, user.ID.String(), resp.UserID)
		assert.Equal(t, user.Email.Value(), resp.Email)
		assert.Equal(t, user.Username.Value(), resp.Username)
		assert.Equal(t, "USD", resp.Currency)
	})
}
