package identity

import (
	"testing"

	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	t.Run("creates valid user", func(t *testing.T) {
		// Arrange
		id, _ := identifier.NewID()
		username, _ := NewUsernameVO("testuser")
				email, _ := NewEmailVO("test@example.com")
				password, _ := NewPasswordVO("$2a		2$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12")
				currency, _ := NewCurrencyVO("USD")
		
				// Act
				user := NewUser(id, username, email, password, currency)
		
				// Assert
				assert.NotNil(t, user)
				assert.Equal(t, id, user.ID)
				assert.Equal(t, username, user.Username)
				assert.Equal(t, email, user.Email)
				assert.Equal(t, password, user.Password)
				assert.Equal(t, currency, user.Currency)
			})
		}
		