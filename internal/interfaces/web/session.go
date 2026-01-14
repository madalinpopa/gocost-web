package web

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/madalinpopa/gocost-web/internal/config"
)

type contextKey string

const (
	IsAuthenticatedKey    = contextKey("isAuthenticated")
	AuthenticatedUserKey  = contextKey("authenticatedUser")
	authenticatedUserID   = "authenticatedUserID"
	authenticatedUsername = "authenticatedUsername"
)

type AuthSessionManager interface {
	RenewToken(ctx context.Context) error
	Destroy(ctx context.Context) error
	IsAuthenticated(ctx context.Context) bool
	GetSessionStore() *scs.SessionManager
	GetUserID(ctx context.Context) string
	GetUsername(ctx context.Context) string
	SetUserID(ctx context.Context, userID string)
	SetUsername(ctx context.Context, username string)
}

// AuthenticatedUser represents the user data stored in the session and context.
// It provides a clear structure for accessing authenticated user information.
type AuthenticatedUser struct {
	ID       string
	Username string
}

// Manager provides a wrapper around scs.SessionManager to manage session operations.
type Manager struct {
	Manager *scs.SessionManager
}

// NewSession initializes and returns a new instance of Manager with a SQLite store and a 12-hour session lifetime.
func NewSession(db *sql.DB, c *config.Config) *Manager {
	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(db)
	sessionManager.Lifetime = 1 * time.Hour
	sessionManager.IdleTimeout = 20 * time.Minute
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Persist = false // TODO: Need to implement remember me
	if c.GetEnvironment() == "production" {
		sessionManager.Cookie.SameSite = http.SameSiteStrictMode
		sessionManager.Cookie.Secure = true
	}
	return &Manager{
		Manager: sessionManager,
	}
}

func (m *Manager) GetSessionStore() *scs.SessionManager {
	return m.Manager
}

func (m *Manager) RenewToken(ctx context.Context) error {
	return m.Manager.RenewToken(ctx)
}

func (m *Manager) Destroy(ctx context.Context) error {
	return m.Manager.Destroy(ctx)
}

func (m *Manager) IsAuthenticated(ctx context.Context) bool {
	return m.Manager.Exists(ctx, authenticatedUserID)
}

func (m *Manager) GetUserID(ctx context.Context) string {
	return m.Manager.GetString(ctx, authenticatedUserID)
}

func (m *Manager) GetUsername(ctx context.Context) string {
	return m.Manager.GetString(ctx, authenticatedUsername)
}

func (m *Manager) SetUserID(ctx context.Context, userID string) {
	m.Manager.Put(ctx, authenticatedUserID, userID)
}

func (m *Manager) SetUsername(ctx context.Context, username string) {
	m.Manager.Put(ctx, authenticatedUsername, username)
}
