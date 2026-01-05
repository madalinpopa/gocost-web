package identity

import (
	"github.com/madalinpopa/gocost-web/internal/shared/identifier"
)

type ID = identifier.ID

type User struct {
	ID       ID
	Email    EmailVO
	Username UsernameVO
	Password PasswordVO
}

func NewUser(id ID, username UsernameVO, email EmailVO, password PasswordVO) *User {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
		Password: password,
	}
}
