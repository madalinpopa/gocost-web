package views

import (
	"context"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
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

func NewUserFromSession(ctx context.Context, s web.AuthSessionManager) *UserView {
	return &UserView{
		ID:       s.GetUserID(ctx),
		Username: s.GetUsername(ctx),
	}
}
