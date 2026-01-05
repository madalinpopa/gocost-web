package identity

import "regexp"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

const (
	minHashLength      = 59
	maximumEmailLength = 254
	minimumEmailLength = 5
)

type EmailVO struct {
	address string
}

func NewEmailVO(address string) (EmailVO, error) {
	if address == "" {
		return EmailVO{}, ErrEmptyEmail
	}
	if len(address) > maximumEmailLength {
		return EmailVO{}, ErrEmailTooLong
	}
	if len(address) < minimumEmailLength {
		return EmailVO{}, ErrEmailTooShort
	}
	if !emailRegex.MatchString(address) {
		return EmailVO{}, ErrInvalidEmailFormat
	}

	return EmailVO{address: address}, nil
}

func (e EmailVO) Value() string {
	return e.address
}

func (e EmailVO) String() string {
	return e.address
}

func (e EmailVO) Equals(other EmailVO) bool {
	return e.address == other.address
}

type UsernameVO struct {
	name string
}

func NewUsernameVO(name string) (UsernameVO, error) {
	if name == "" {
		return UsernameVO{}, ErrEmptyUsername
	}
	if len(name) > 30 {
		return UsernameVO{}, ErrUsernameTooLong
	}
	if len(name) < 3 {
		return UsernameVO{}, ErrUsernameTooShort
	}
	return UsernameVO{name: name}, nil
}

func (u UsernameVO) Value() string {
	return u.name
}

func (u UsernameVO) String() string {
	return u.name
}

func (u UsernameVO) Equals(other UsernameVO) bool {
	return u.name == other.name
}

type PasswordVO struct {
	hash string
}

func NewPasswordVO(hash string) (PasswordVO, error) {
	if hash == "" {
		return PasswordVO{}, ErrEmptyPassword
	}
	if len(hash) < minHashLength {
		return PasswordVO{}, ErrInvalidHash
	}
	return PasswordVO{hash: hash}, nil
}

func (p PasswordVO) Value() string {
	return p.hash
}

func (p PasswordVO) String() string {
	return p.hash
}
