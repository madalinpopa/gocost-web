package mocks

import (
	"context"

	"github.com/alexedwards/scs/v2"
	"github.com/stretchr/testify/mock"
)

type MockSessionManager struct {
	mock.Mock
}

func (m *MockSessionManager) RenewToken(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSessionManager) Destroy(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSessionManager) IsAuthenticated(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockSessionManager) GetSessionStore() *scs.SessionManager {
	args := m.Called()
	return args.Get(0).(*scs.SessionManager)
}

func (m *MockSessionManager) GetUserID(ctx context.Context) string {
	args := m.Called(ctx)
	return args.String(0)
}

func (m *MockSessionManager) GetUsername(ctx context.Context) string {
	args := m.Called(ctx)
	return args.String(0)
}

func (m *MockSessionManager) SetUserID(ctx context.Context, userID string) {
	m.Called(ctx, userID)
}

func (m *MockSessionManager) SetUsername(ctx context.Context, username string) {
	m.Called(ctx, username)
}
