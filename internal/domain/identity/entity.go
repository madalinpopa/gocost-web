package identity

import (
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
)

type ID = identifier.ID

type User struct {
		ID       ID
		Email    EmailVO
		Username UsernameVO
		Password PasswordVO
		Currency CurrencyVO
	}
	
	func NewUser(id ID, username UsernameVO, email EmailVO, password PasswordVO, currency CurrencyVO) *User {
		return &User{
			ID:       id,
			Username: username,
			Email:    email,
			Password: password,
			Currency: currency,
		}
	}
	