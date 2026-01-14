package web_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func newTestManagerWithContext(t *testing.T) (*web.Manager, context.Context) {
	t.Helper()

	sessionManager := scs.New()
	ctx, err := sessionManager.Load(context.Background(), "")
	if err != nil {
		t.Fatalf("load session context: %v", err)
	}

	return &web.Manager{Manager: sessionManager}, ctx
}

func TestNew_NonProductionDefaults(t *testing.T) {
	db := newTestDB(t)
	cfg := config.New().WithEnvironment("development")

	manager := web.NewSession(db, cfg)
	assert.NotNil(t, manager)
	if manager == nil {
		return
	}
	assert.NotNil(t, manager.Manager)
	if manager.Manager == nil {
		return
	}

	store, ok := manager.Manager.Store.(*sqlite3store.SQLite3Store)
	assert.True(t, ok)
	if !ok {
		return
	}
	t.Cleanup(store.StopCleanup)

	assert.Equal(t, time.Hour, manager.Manager.Lifetime)
	assert.Equal(t, 20*time.Minute, manager.Manager.IdleTimeout)
	assert.True(t, manager.Manager.Cookie.HttpOnly)
	assert.False(t, manager.Manager.Cookie.Persist)
	assert.Equal(t, http.SameSiteLaxMode, manager.Manager.Cookie.SameSite)
	assert.False(t, manager.Manager.Cookie.Secure)
}

func TestNew_ProductionCookie(t *testing.T) {
	db := newTestDB(t)
	cfg := config.New().WithEnvironment("production")

	manager := web.NewSession(db, cfg)
	assert.NotNil(t, manager)
	if manager == nil {
		return
	}
	assert.NotNil(t, manager.Manager)
	if manager.Manager == nil {
		return
	}
	store, ok := manager.Manager.Store.(*sqlite3store.SQLite3Store)
	assert.True(t, ok)
	if !ok {
		return
	}
	t.Cleanup(store.StopCleanup)

	assert.Equal(t, http.SameSiteStrictMode, manager.Manager.Cookie.SameSite)
	assert.True(t, manager.Manager.Cookie.Secure)
}

func TestManager_UserAccessors(t *testing.T) {
	manager, ctx := newTestManagerWithContext(t)

	assert.False(t, manager.IsAuthenticated(ctx))
	assert.Empty(t, manager.GetUserID(ctx))
	assert.Empty(t, manager.GetUsername(ctx))

	manager.SetUserID(ctx, "user-123")
	manager.SetUsername(ctx, "alice")

	assert.True(t, manager.IsAuthenticated(ctx))
	assert.Equal(t, "user-123", manager.GetUserID(ctx))
	assert.Equal(t, "alice", manager.GetUsername(ctx))
}

func TestManager_RenewTokenAndDestroy(t *testing.T) {
	manager, ctx := newTestManagerWithContext(t)

	assert.NoError(t, manager.RenewToken(ctx))
	assert.NotEmpty(t, manager.GetSessionStore().Token(ctx))

	manager.SetUserID(ctx, "user-123")

	assert.NoError(t, manager.Destroy(ctx))
	assert.False(t, manager.IsAuthenticated(ctx))
	assert.Empty(t, manager.GetSessionStore().Token(ctx))
}
