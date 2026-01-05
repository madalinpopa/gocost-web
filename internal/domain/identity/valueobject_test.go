package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmailVO(t *testing.T) {
	t.Run("empty email", func(t *testing.T) {
		// Arrange

		// Act
		_, err := NewEmailVO("")

		// Assert
		assert.ErrorIs(t, err, ErrEmptyEmail)
	})

	t.Run("exceding maximum length email", func(t *testing.T) {
		// Arrange
		longEmail := "a" + string(make([]byte, 253-len("@example.com"))) + "@example.com"

		// Act
		emailVO, err := NewEmailVO(longEmail)

		// Assert
		assert.Error(t, err, ErrEmailTooLong)
		assert.Empty(t, emailVO)
	})

	t.Run("valid email", func(t *testing.T) {
		// Arrange
		validEmail := "user@example.com"

		// Act
		emailVO, err := NewEmailVO(validEmail)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, validEmail, emailVO.Value())
	})

	t.Run("invalid email format", func(t *testing.T) {
		// Arrange
		invalidEmail := "user@@example..com"

		// Act
		_, err := NewEmailVO(invalidEmail)

		// Assert
		assert.ErrorIs(t, err, ErrInvalidEmailFormat)
	})
}

func TestEmailVO_Equals(t *testing.T) {
	t.Run("equal emails", func(t *testing.T) {
		// Arrange
		email := "test@example.com"
		email1, _ := NewEmailVO(email)
		email2, _ := NewEmailVO(email)

		// Act
		areEqual := email1.Equals(email2)

		// Assert
		assert.True(t, areEqual)
	})

	t.Run("unequal emails", func(t *testing.T) {
		// Arrange
		email1, _ := NewEmailVO("test1@example.com")
		email2, _ := NewEmailVO("test2@example.com")

		// Act
		areEqual := email1.Equals(email2)

		// Assert
		assert.False(t, areEqual)
	})
}

func TestEmailVO_String(t *testing.T) {
	t.Run("string representation", func(t *testing.T) {
		// Arrange
		emailStr := "test@example.com"
		emailVO, _ := NewEmailVO(emailStr)

		// Act
		result := emailVO.String()

		// Assert
		assert.Equal(t, emailStr, result)
	})
}

func TestEmailVO_Value(t *testing.T) {
	t.Run("value retrieval", func(t *testing.T) {
		// Arrange
		emailStr := "test@example.com"
		emailVO, _ := NewEmailVO(emailStr)

		// Act
		result := emailVO.Value()

		// Assert
		assert.Equal(t, emailStr, result)
	})
}

func TestNewUsernameVO(t *testing.T) {
	t.Run("empty username", func(t *testing.T) {
		// Arrange

		// Act
		_, err := NewUsernameVO("")

		// Assert
		assert.ErrorIs(t, err, ErrEmptyUsername)
	})

	t.Run("exceding maximum length username", func(t *testing.T) {
		// Arrange
		longUsername := string(make([]byte, 31))

		// Act
		usernameVO, err := NewUsernameVO(longUsername)

		// Assert
		assert.ErrorIs(t, err, ErrUsernameTooLong)
		assert.Empty(t, usernameVO)
	})

	t.Run("short username", func(t *testing.T) {
		// Arrange
		shortUsername := "ab"

		// Act
		_, err := NewUsernameVO(shortUsername)

		// Assert
		assert.ErrorIs(t, err, ErrUsernameTooShort)
	})

	t.Run("valid username", func(t *testing.T) {
		// Arrange
		validUsername := "valid-user"

		// Act
		usernameVO, err := NewUsernameVO(validUsername)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, validUsername, usernameVO.Value())
	})
}

func TestUsernameVO_Equals(t *testing.T) {
	t.Run("equal usernames", func(t *testing.T) {
		// Arrange
		username := "test-user"
		username1, _ := NewUsernameVO(username)
		username2, _ := NewUsernameVO(username)

		// Act
		areEqual := username1.Equals(username2)

		// Assert
		assert.True(t, areEqual)
	})

	t.Run("unequal usernames", func(t *testing.T) {
		// Arrange
		username1, _ := NewUsernameVO("user-one")
		username2, _ := NewUsernameVO("user-two")

		// Act
		areEqual := username1.Equals(username2)

		// Assert
		assert.False(t, areEqual)
	})
}

func TestUsernameVO_String(t *testing.T) {
	t.Run("string representation", func(t *testing.T) {
		// Arrange
		usernameStr := "test-user"
		usernameVO, _ := NewUsernameVO(usernameStr)

		// Act
		result := usernameVO.String()

		// Assert
		assert.Equal(t, usernameStr, result)
	})
}

func TestUsernameVO_Value(t *testing.T) {
	t.Run("value retrieval", func(t *testing.T) {
		// Arrange
		usernameStr := "test-user"
		usernameVO, _ := NewUsernameVO(usernameStr)

		// Act
		result := usernameVO.Value()

		// Assert
		assert.Equal(t, usernameStr, result)
	})
}

func TestNewPasswordVO(t *testing.T) {
	t.Run("empty password", func(t *testing.T) {
		// Arrange

		// Act
		_, err := NewPasswordVO("")

		// Assert
		assert.ErrorIs(t, err, ErrEmptyPassword)
	})

	t.Run("invalid hash length", func(t *testing.T) {
		// Arrange
		// 58 chars
		shortHash := "1234567890123456789012345678901234567890123456789012345678"

		// Act
		_, err := NewPasswordVO(shortHash)

		// Assert
		assert.ErrorIs(t, err, ErrInvalidHash)
	})

	t.Run("valid password hash", func(t *testing.T) {
		// Arrange
		// 60 chars
		validHash := "$2a$12$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12"

		// Act
		passwordVO, err := NewPasswordVO(validHash)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, validHash, passwordVO.Value())
	})
}

func TestPasswordVO_String(t *testing.T) {
	t.Run("string representation", func(t *testing.T) {
		// Arrange
		passwordStr := "$2a$12$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12"
		passwordVO, _ := NewPasswordVO(passwordStr)

		// Act
		result := passwordVO.String()

		// Assert
		assert.Equal(t, passwordStr, result)
	})
}

func TestPasswordVO_Value(t *testing.T) {
	t.Run("value retrieval", func(t *testing.T) {
		// Arrange
		passwordStr := "$2a$12$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12"
		passwordVO, _ := NewPasswordVO(passwordStr)

		// Act
		result := passwordVO.Value()

		// Assert
		assert.Equal(t, passwordStr, result)
	})
}
