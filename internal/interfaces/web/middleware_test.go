package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/stretchr/testify/assert"
)

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
