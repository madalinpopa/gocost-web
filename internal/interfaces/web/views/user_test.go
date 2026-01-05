package views

import (
	"context"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestNewUserViewFromResponse(t *testing.T) {
	res := &usecase.UserResponse{
		ID:       "user-123",
		Email:    "user@example.com",
		Username: "alice",
	}

	view := NewUserViewFromResponse(res)

	assert.NotNil(t, view)
	assert.Equal(t, "user-123", view.ID)
	assert.Equal(t, "user@example.com", view.Email)
	assert.Equal(t, "alice", view.Username)
}

func TestNewUserFromSession(t *testing.T) {
	ctx := context.Background()
	sessionManager := &mockAuthSessionManager{}
	sessionManager.On("GetUserID", ctx).Return("user-456")
	sessionManager.On("GetUsername", ctx).Return("bob")

	view := NewUserFromSession(ctx, sessionManager)

	assert.NotNil(t, view)
	assert.Equal(t, "user-456", view.ID)
	assert.Equal(t, "bob", view.Username)
	assert.Empty(t, view.Email)
}
