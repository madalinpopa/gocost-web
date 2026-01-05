package views

import (
	"context"

	"github.com/madalinpopa/gocost-web/internal/infrastructure/session"
	"github.com/madalinpopa/gocost-web/internal/usecase"
)

type UserView struct {
	ID       string
	Email    string
	Username string
}

func NewUserViewFromResponse(res *usecase.UserResponse) *UserView {
	return &UserView{
		ID:       res.ID,
		Email:    res.Email,
		Username: res.Username,
	}
}

func NewUserFromSession(ctx context.Context, s session.AuthSessionManager) *UserView {
	return &UserView{
		ID:       s.GetUserID(ctx),
		Username: s.GetUsername(ctx),
	}
}
