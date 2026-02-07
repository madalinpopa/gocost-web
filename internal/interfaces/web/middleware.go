package web

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/justinas/nosurf"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"
)

// responseWriter is a wrapper around responseWriter that captures the status code of the response.
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader sets the HTTP status code for the response and writes it using the underlying ResponseWriter.
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Unwrap returns the underlying ResponseWriter for http.ResponseController compatibility.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

type Middleware struct {
	logger  *slog.Logger
	config  *config.Config
	session AuthSessionManager
	errors  respond.ErrorHandler

	trustedProxyParseOnce sync.Once
	trustedProxyCIDRs     []*net.IPNet
	trustedProxyIPs       []net.IP

	allowedHostsParseOnce  sync.Once
	normalizedAllowedHosts []string
}

func NewMiddleware(l *slog.Logger, c *config.Config, s AuthSessionManager, errors respond.ErrorHandler) *Middleware {
	if errors == nil {
		errors = respond.NewErrorHandler(l)
	}

	m := &Middleware{
		logger:  l,
		config:  c,
		session: s,
		errors:  errors,
	}

	m.parseTrustedProxies()
	m.parseAllowedHosts()

	return m
}

// Headers sets HTTP security headers to enhance security and forwards the request to the next handler.
func (m *Middleware) Headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Security-Policy to restrict the sources of content such as scripts, styles, and images
		w.Header().Set("Content-Security-Policy", strings.TrimSpace(`
			default-src 'self';
			script-src 'self' 'unsafe-eval' 'unsafe-inline' cdn.jsdelivr.net umami.coderustle.dev;
			style-src 'self' 'unsafe-inline' cdn.jsdelivr.net;
			img-src 'self' data: cdn.jsdelivr.net api.iconify.design;
			font-src 'self' cdn.jsdelivr.net;
			connect-src 'self' api.iconify.design umami.coderustle.dev;
			object-src 'none';
			base-uri 'self';
			form-action 'self';
		`))

		// Set Referrer-Policy to control the amount of referrer information sent with requests
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Set X-Content-Type-Options to prevent browsers from interpreting files as a different MIME type
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Set X-Frame-Options to prevent clickjacking attacks by disallowing the page from being framed
		w.Header().Set("X-Frame-Options", "deny")

		// Set X-XSS-Protection to turn off the browser's XSS protection, preventing unintended behavior
		w.Header().Set("X-XSS-Protection", "0")

		// Set Server header to a generic value
		w.Header().Set("Server", "Go")

		// HSTS.
		// TODO: Enable this only for production
		// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Set Permissions-Policy to control which browser features and APIs can be used by the application
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// Logging is middleware that logs HTTP request details such as IP, protocol, method, URL, and response status.
func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		ip := m.getClientIP(r)

		var (
			proto  = r.Proto
			method = r.Method
			url    = r.URL.RequestURI()
			status = rw.status
		)

		m.logger.Info("request", "ip", ip, "proto", proto, "method", method, "url", url, "status", status)
	})
}

// Recover handles panics during HTTP request processing, logs the error, and sends a 500 response with connection closed.
func (m *Middleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.Error("panic recovered", "error", err, "stack", string(debug.Stack()))
				w.Header().Set("Connection", "close")
				m.errors.ServerError(w, r, fmt.Errorf("panic recovered: %v", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CsrfToken is middleware that applies CSRF protection to HTTP requests using nosurf, setting a secure base cookie.
func (m *Middleware) CsrfToken(next http.Handler) http.Handler {
	cookie := http.Cookie{
		Name:     "gocost_csrf",
		HttpOnly: true,
		MaxAge:   86400, // 24 hours
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	if m.config.GetEnvironment() == "production" {
		cookie.Domain = m.config.Domain
		cookie.Secure = true
		cookie.SameSite = http.SameSiteStrictMode
	}

	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(cookie)

	csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logger.Error("CSRF validation failed", "reason", nosurf.Reason(r))
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}))

	return csrfHandler
}

// CheckAllowedHosts checks if the request is coming from an allowed host.
func (m *Middleware) CheckAllowedHosts(next http.Handler) http.Handler {
	if m.config == nil || len(m.config.AllowedHosts) == 0 {
		return next
	}

	m.parseAllowedHosts()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}
		host = normalizeHost(host)

		if !isAllowedHost(host, m.normalizedAllowedHosts) {
			http.Error(w, "Forbidden: host not allowed", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) getClientIP(r *http.Request) string {
	remoteIP := extractRemoteIP(r.RemoteAddr)

	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" && m.isTrustedProxy(remoteIP) {
		// Take the first IP and validate it
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				return parsedIP.String()
			}
		}
	}

	return remoteIP
}

func (m *Middleware) isTrustedProxy(remoteIP string) bool {
	ip := net.ParseIP(remoteIP)
	if ip == nil {
		return false
	}

	m.parseTrustedProxies()

	for _, trustedCIDR := range m.trustedProxyCIDRs {
		if trustedCIDR.Contains(ip) {
			return true
		}
	}

	for _, trustedIP := range m.trustedProxyIPs {
		if trustedIP.Equal(ip) {
			return true
		}
	}

	return false
}

func (m *Middleware) parseTrustedProxies() {
	m.trustedProxyParseOnce.Do(func() {
		if m.config == nil || len(m.config.TrustedProxies) == 0 {
			return
		}

		for _, trustedProxy := range m.config.TrustedProxies {
			candidate := strings.TrimSpace(trustedProxy)
			if candidate == "" {
				continue
			}

			if _, cidr, err := net.ParseCIDR(candidate); err == nil {
				m.trustedProxyCIDRs = append(m.trustedProxyCIDRs, cidr)
				continue
			}

			if trustedIP := net.ParseIP(candidate); trustedIP != nil {
				m.trustedProxyIPs = append(m.trustedProxyIPs, trustedIP)
			}
		}
	})
}

func isAllowedHost(host string, normalizedAllowedHosts []string) bool {
	if len(normalizedAllowedHosts) == 0 {
		return false
	}

	parsedIP := net.ParseIP(host)
	for _, candidate := range normalizedAllowedHosts {
		if candidate == "" {
			continue
		}

		if candidate == host {
			return true
		}

		if parsedIP == nil {
			continue
		}

		if _, cidr, err := net.ParseCIDR(candidate); err == nil && cidr.Contains(parsedIP) {
			return true
		}
	}

	return false
}

func (m *Middleware) parseAllowedHosts() {
	m.allowedHostsParseOnce.Do(func() {
		if m.config == nil || len(m.config.AllowedHosts) == 0 {
			return
		}

		m.normalizedAllowedHosts = make([]string, 0, len(m.config.AllowedHosts))
		for _, allowedHost := range m.config.AllowedHosts {
			m.normalizedAllowedHosts = append(m.normalizedAllowedHosts, normalizeHost(allowedHost))
		}
	})
}

func extractRemoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		if parsedIP := net.ParseIP(host); parsedIP != nil {
			return parsedIP.String()
		}

		return host
	}

	if parsedIP := net.ParseIP(remoteAddr); parsedIP != nil {
		return parsedIP.String()
	}

	return remoteAddr
}

func normalizeHost(host string) string {
	trimmed := strings.TrimSpace(strings.ToLower(host))
	trimmed = strings.TrimPrefix(trimmed, "[")
	trimmed = strings.TrimSuffix(trimmed, "]")
	return strings.TrimSuffix(trimmed, ".")
}

// LoadSession loads and saves session data to and from the session cookie.
func (m *Middleware) LoadSession(next http.Handler) http.Handler {
	return m.session.GetSessionStore().LoadAndSave(next)
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the user ID from the session.
		// If it doesn't exist, call the next handler and return.
		userID := m.session.GetUserID(r.Context())
		if userID == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Create a user object from the session data.
		user := AuthenticatedUser{
			ID:       userID,
			Username: m.session.GetUsername(r.Context()),
		}

		// Add the authentication status and user info to the request context.
		ctx := context.WithValue(r.Context(), IsAuthenticatedKey, true)
		ctx = context.WithValue(ctx, AuthenticatedUserKey, user)
		r = r.WithContext(ctx)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) LoginRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := m.session.GetUserID(r.Context())
		if userID == "" {
			if m.session.IsAuthenticated(r.Context()) {
				// Session exists, but user data is missing - possible tampering.
				if err := m.session.Destroy(r.Context()); err != nil {
					m.logger.Error(err.Error(), "method", r.Method, "url", r.URL.RequestURI())
					m.errors.LogServerError(r, fmt.Errorf("failed to destroy session: %w", err))
				}
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require authentication aren't stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")

		// And call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
