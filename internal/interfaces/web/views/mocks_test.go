package views

import (
	"context"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/mock"
)

type mockAuthSessionManager struct {
	mock.Mock
}

func (m *mockAuthSessionManager) RenewToken(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockAuthSessionManager) Destroy(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockAuthSessionManager) IsAuthenticated(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *mockAuthSessionManager) GetSessionStore() *scs.SessionManager {
	args := m.Called()
	return args.Get(0).(*scs.SessionManager)
}

func (m *mockAuthSessionManager) GetUserID(ctx context.Context) string {
	args := m.Called(ctx)
	return args.String(0)
}

func (m *mockAuthSessionManager) GetUsername(ctx context.Context) string {
	args := m.Called(ctx)
	return args.String(0)
}

func (m *mockAuthSessionManager) SetUserID(ctx context.Context, userID string) {
	m.Called(ctx, userID)
}

func (m *mockAuthSessionManager) SetUsername(ctx context.Context, username string) {
	m.Called(ctx, username)
}
