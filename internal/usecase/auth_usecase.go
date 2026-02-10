package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/madalinpopa/gocost-web/internal/domain"
	"github.com/madalinpopa/gocost-web/internal/domain/identity"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/security"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

const minPasswordLength = 8

type AuthUseCaseImpl struct {
	uow    domain.UnitOfWork
	logger *slog.Logger
	hasher security.PasswordHasher
}

func NewAuthUseCase(uow domain.UnitOfWork, logger *slog.Logger, h security.PasswordHasher) AuthUseCaseImpl {
	return AuthUseCaseImpl{
		uow:    uow,
		logger: logger,
		hasher: h,
	}
}

func (u AuthUseCaseImpl) Register(ctx context.Context, req *RegisterUserRequest) (*UserResponse, error) {
	if req == nil {
		return nil, errors.New("register request is nil")
	}

	email, err := identity.NewEmailVO(req.Email)
	if err != nil {
		return nil, err
	}

	username, err := identity.NewUsernameVO(req.Username)
	if err != nil {
		return nil, err
	}

	if err := u.hasher.ValidatePassword(req.Password); err != nil {
		return nil, err
	}
	if len(req.Password) < minPasswordLength {
		return nil, identity.ErrPasswordTooShort
	}

	repo := u.uow.UserRepository()
	exists, err := repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, identity.ErrUserAlreadyExists
	}

	exists, err = repo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, identity.ErrUserAlreadyExists
	}

	hashedPassword, err := u.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	password, err := identity.NewPasswordVO(hashedPassword)
	if err != nil {
		return nil, err
	}

	id, err := identifier.NewID()
	if err != nil {
		return nil, err
	}

	currency, err := identity.NewCurrencyVO(req.Currency)
	if err != nil {
		return nil, err
	}

	user := identity.NewUser(id, username, email, password, currency)

	txUOW, err := u.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}

	if err := txUOW.UserRepository().Save(ctx, *user); err != nil {
		_ = txUOW.Rollback()
		return nil, err
	}

	if err := txUOW.Commit(); err != nil {
		_ = txUOW.Rollback()
		return nil, err
	}

	return &UserResponse{
		ID:       user.ID.String(),
		Email:    user.Email.Value(),
		Username: user.Username.Value(),
		Currency: user.Currency.Value(),
	}, nil
}

func (u AuthUseCaseImpl) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	if req == nil {
		return nil, errors.New("login request is nil")
	}
	if err := u.hasher.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	repo := u.uow.UserRepository()

	var (
		user identity.User
		err  error
	)
	if email, emailErr := identity.NewEmailVO(req.EmailOrUsername); emailErr == nil {
		user, err = repo.FindByEmail(ctx, email)
	} else {
		username, usernameErr := identity.NewUsernameVO(req.EmailOrUsername)
		if usernameErr != nil {
			return nil, usernameErr
		}
		user, err = repo.FindByUsername(ctx, username)
	}

	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := u.hasher.ValidateHash(user.Password.Value()); err != nil {
		return nil, err
	}
	if !u.hasher.CheckPasswordHash(req.Password, user.Password.Value()) {
		return nil, ErrInvalidCredentials
	}

	return &LoginResponse{
		UserID:   user.ID.String(),
		Email:    user.Email.Value(),
		Username: user.Username.Value(),
		FullName: "",
		Role:     "",
		Currency: user.Currency.Value(),
	}, nil
}
