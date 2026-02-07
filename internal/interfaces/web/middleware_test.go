package web

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/stretchr/testify/assert"
)

type stubAuthSessionManager struct {
	isAuthenticated bool
	userID          string
	username        string
	destroyErr      error
	destroyCalled   bool
}

func (s *stubAuthSessionManager) RenewToken(context.Context) error {
	return nil
}

func (s *stubAuthSessionManager) Destroy(context.Context) error {
	s.destroyCalled = true
	return s.destroyErr
}

func (s *stubAuthSessionManager) IsAuthenticated(context.Context) bool {
	return s.isAuthenticated
}

func (s *stubAuthSessionManager) GetSessionStore() *scs.SessionManager {
	return nil
}

func (s *stubAuthSessionManager) GetUserID(context.Context) string {
	return s.userID
}

func (s *stubAuthSessionManager) GetUsername(context.Context) string {
	return s.username
}

func (s *stubAuthSessionManager) GetCurrency(context.Context) string {
	return ""
}

func (s *stubAuthSessionManager) SetUserID(context.Context, string) {}

func (s *stubAuthSessionManager) SetUsername(context.Context, string) {}

func (s *stubAuthSessionManager) SetCurrency(context.Context, string) {}

type stubErrorHandler struct {
	logServerErrorCalls int
}

func (s *stubErrorHandler) ServerError(http.ResponseWriter, *http.Request, error) {}

func (s *stubErrorHandler) Error(http.ResponseWriter, *http.Request, int, error) {}

func (s *stubErrorHandler) LogServerError(*http.Request, error) {
	s.logServerErrorCalls++
}

func TestMiddleware_getClientIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		trusted    []string
		want       string
	}{
		{
			name:       "ignores xff when remote is not trusted",
			remoteAddr: "203.0.113.10:4321",
			xff:        "198.51.100.7",
			trusted:    nil,
			want:       "203.0.113.10",
		},
		{
			name:       "uses xff when remote is trusted by exact ip",
			remoteAddr: "203.0.113.10:4321",
			xff:        "198.51.100.7",
			trusted:    []string{"203.0.113.10"},
			want:       "198.51.100.7",
		},
		{
			name:       "uses xff when remote is trusted by cidr",
			remoteAddr: "203.0.113.10:4321",
			xff:        "198.51.100.8",
			trusted:    []string{"203.0.113.0/24"},
			want:       "198.51.100.8",
		},
		{
			name:       "falls back to remote when first xff ip is invalid",
			remoteAddr: "203.0.113.10:4321",
			xff:        "invalid, 198.51.100.9",
			trusted:    []string{"203.0.113.10"},
			want:       "203.0.113.10",
		},
		{
			name:       "handles malformed remote address gracefully",
			remoteAddr: "malformed-remote",
			xff:        "198.51.100.10",
			trusted:    []string{"203.0.113.10"},
			want:       "malformed-remote",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &Middleware{
				config: &config.Config{
					TrustedProxies: tt.trusted,
				},
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}

			assert.Equal(t, tt.want, m.getClientIP(req))
		})
	}
}

func TestMiddleware_CheckAllowedHosts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		allowed     []string
		requestHost string
		wantAllowed bool
	}{
		{
			name:        "allows matching domain",
			allowed:     []string{"localhost"},
			requestHost: "localhost:8080",
			wantAllowed: true,
		},
		{
			name:        "allows matching ipv4",
			allowed:     []string{"127.0.0.1"},
			requestHost: "127.0.0.1:8080",
			wantAllowed: true,
		},
		{
			name:        "allows matching ipv4 cidr",
			allowed:     []string{"192.168.1.0/24"},
			requestHost: "192.168.1.55:8080",
			wantAllowed: true,
		},
		{
			name:        "allows matching ipv6",
			allowed:     []string{"2001:db8::1"},
			requestHost: "[2001:db8::1]:8080",
			wantAllowed: true,
		},
		{
			name:        "allows matching ipv6 cidr",
			allowed:     []string{"2001:db8::/32"},
			requestHost: "[2001:db8::2]:8080",
			wantAllowed: true,
		},
		{
			name:        "rejects non matching host",
			allowed:     []string{"localhost"},
			requestHost: "example.com:8080",
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &Middleware{
				config: &config.Config{
					AllowedHosts: tt.allowed,
				},
			}

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = tt.requestHost
			rec := httptest.NewRecorder()

			m.CheckAllowedHosts(next).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantAllowed, nextCalled)
			if tt.wantAllowed {
				assert.Equal(t, http.StatusNoContent, rec.Code)
				return
			}
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
	}
}

func TestMiddleware_LoginRequired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		session               *stubAuthSessionManager
		wantStatus            int
		wantLocation          string
		wantNextCalled        bool
		wantDestroyCalled     bool
		wantLogServerErrCalls int
		wantCacheControl      string
	}{
		{
			name: "redirects when unauthenticated",
			session: &stubAuthSessionManager{
				isAuthenticated: false,
				userID:          "",
			},
			wantStatus:            http.StatusSeeOther,
			wantLocation:          "/login",
			wantNextCalled:        false,
			wantDestroyCalled:     false,
			wantLogServerErrCalls: 0,
			wantCacheControl:      "",
		},
		{
			name: "destroys session and redirects when authenticated but user id missing",
			session: &stubAuthSessionManager{
				isAuthenticated: true,
				userID:          "",
			},
			wantStatus:            http.StatusSeeOther,
			wantLocation:          "/login",
			wantNextCalled:        false,
			wantDestroyCalled:     true,
			wantLogServerErrCalls: 0,
			wantCacheControl:      "",
		},
		{
			name: "logs server error when destroy fails",
			session: &stubAuthSessionManager{
				isAuthenticated: true,
				userID:          "",
				destroyErr:      errors.New("destroy failed"),
			},
			wantStatus:            http.StatusSeeOther,
			wantLocation:          "/login",
			wantNextCalled:        false,
			wantDestroyCalled:     true,
			wantLogServerErrCalls: 1,
			wantCacheControl:      "",
		},
		{
			name: "allows request when user id is present",
			session: &stubAuthSessionManager{
				isAuthenticated: true,
				userID:          "user-123",
			},
			wantStatus:            http.StatusNoContent,
			wantLocation:          "",
			wantNextCalled:        true,
			wantDestroyCalled:     false,
			wantLogServerErrCalls: 0,
			wantCacheControl:      "no-store",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errHandler := &stubErrorHandler{}
			m := &Middleware{
				logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
				session: tt.session,
				errors:  errHandler,
			}

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
			rec := httptest.NewRecorder()

			m.LoginRequired(next).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantLocation, rec.Header().Get("Location"))
			assert.Equal(t, tt.wantNextCalled, nextCalled)
			assert.Equal(t, tt.wantDestroyCalled, tt.session.destroyCalled)
			assert.Equal(t, tt.wantLogServerErrCalls, errHandler.logServerErrorCalls)
			assert.Equal(t, tt.wantCacheControl, rec.Header().Get("Cache-Control"))
		})
	}
}
