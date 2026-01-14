package router

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/justinas/alice"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/middleware"
	"github.com/stretchr/testify/assert"
)

// mockHandler is a simple handler that writes a test response
func mockHandler(response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}
}

// mockMiddleware creates a simple middleware for testing
func mockMiddleware(header string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-Middleware", header)
			next.ServeHTTP(w, r)
		})
	}
}

// createTestRouter creates a router with minimal middleware setup for testing
func createTestRouter() *Router {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.New()

	m := &middleware.Middleware{}

	// Create router
	router := &Router{
		mux:        http.NewServeMux(),
		middleware: m,
	}

	// Override middleware chains with simpler test middleware to avoid dependencies
	router.baseRoutes = alice.New(mockMiddleware("base"))
	router.dynamicRoutes = alice.New(mockMiddleware("dynamic"))
	router.protectedRoutes = alice.New(mockMiddleware("protected"))

	// Suppress unused variable warnings
	_ = logger
	_ = cfg

	return router
}

func TestNew(t *testing.T) {
	t.Run("creates router with all middleware chains", func(t *testing.T) {
		// Arrange
		m := &middleware.Middleware{}

		// Act
		router := New(m)

		// Assert
		assert.NotNil(t, router)
		assert.NotNil(t, router.mux)
		assert.NotNil(t, router.middleware)
		assert.NotNil(t, router.baseRoutes)
		assert.NotNil(t, router.dynamicRoutes)
		assert.NotNil(t, router.protectedRoutes)
	})
}

func TestRouter_RegisterPublicHandler(t *testing.T) {
	t.Run("registers GET handler for public route", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		handler := mockHandler("public response")

		// Act
		router.RegisterPublicHandler(http.MethodGet, "/public", handler)

		// Assert
		req := httptest.NewRequest(http.MethodGet, "/public", nil)
		rec := httptest.NewRecorder()
		router.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public response", rec.Body.String())
		assert.Equal(t, "dynamic", rec.Header().Get("X-Test-Middleware"))
	})

	t.Run("registers POST handler for public route", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		handler := mockHandler("created")

		// Act
		router.RegisterPublicHandler(http.MethodPost, "/register", handler)

		// Assert
		req := httptest.NewRequest(http.MethodPost, "/register", nil)
		rec := httptest.NewRecorder()
		router.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "created", rec.Body.String())
	})

	t.Run("different methods on same path", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		getHandler := mockHandler("GET response")
		postHandler := mockHandler("POST response")

		// Act
		router.RegisterPublicHandler(http.MethodGet, "/form", getHandler)
		router.RegisterPublicHandler(http.MethodPost, "/form", postHandler)

		// Assert GET
		getReq := httptest.NewRequest(http.MethodGet, "/form", nil)
		getRec := httptest.NewRecorder()
		router.mux.ServeHTTP(getRec, getReq)
		assert.Equal(t, "GET response", getRec.Body.String())

		// Assert POST
		postReq := httptest.NewRequest(http.MethodPost, "/form", nil)
		postRec := httptest.NewRecorder()
		router.mux.ServeHTTP(postRec, postReq)
		assert.Equal(t, "POST response", postRec.Body.String())
	})
}

func TestRouter_RegisterPrivateHandler(t *testing.T) {
	t.Run("registers protected route", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		handler := mockHandler("private data")

		// Act
		router.RegisterPrivateHandler(http.MethodGet, "/dashboard", handler)

		// Assert
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rec := httptest.NewRecorder()
		router.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "private data", rec.Body.String())
		assert.Equal(t, "protected", rec.Header().Get("X-Test-Middleware"))
	})

	t.Run("registers DELETE handler for private route", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		handler := mockHandler("deleted")

		// Act
		router.RegisterPrivateHandler(http.MethodDelete, "/expenses/123", handler)

		// Assert
		req := httptest.NewRequest(http.MethodDelete, "/expenses/123", nil)
		rec := httptest.NewRecorder()
		router.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "deleted", rec.Body.String())
	})

	t.Run("registers PUT handler for private route", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		handler := mockHandler("updated")

		// Act
		router.RegisterPrivateHandler(http.MethodPut, "/profile", handler)

		// Assert
		req := httptest.NewRequest(http.MethodPut, "/profile", nil)
		rec := httptest.NewRecorder()
		router.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "updated", rec.Body.String())
	})
}

func TestRouter_RegisterUnprotectedHandler(t *testing.T) {
	t.Run("registers unprotected route", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		handler := mockHandler("OK")

		// Act
		router.RegisterUnprotectedHandler(http.MethodGet, "/health", handler)

		// Assert
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		router.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
		assert.Equal(t, "base", rec.Header().Get("X-Test-Middleware"))
	})

	t.Run("registers static health check", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		handler := mockHandler("healthy")

		// Act
		router.RegisterUnprotectedHandler(http.MethodGet, "/api/health", handler)

		// Assert
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()
		router.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "healthy", rec.Body.String())
	})
}

func TestRouter_RegisterStaticFiles(t *testing.T) {
	t.Run("registers static file handler", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/css")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("body { color: red; }"))
		})

		// Act
		router.RegisterStaticFiles("/static/", staticHandler)

		// Assert
		req := httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil)
		rec := httptest.NewRecorder()
		router.mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/css", rec.Header().Get("Content-Type"))
		assert.Equal(t, "body { color: red; }", rec.Body.String())
	})

	t.Run("registers multiple static paths", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		cssHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("css"))
		})
		jsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("js"))
		})

		// Act
		router.RegisterStaticFiles("/css/", cssHandler)
		router.RegisterStaticFiles("/js/", jsHandler)

		// Assert CSS
		cssReq := httptest.NewRequest(http.MethodGet, "/css/style.css", nil)
		cssRec := httptest.NewRecorder()
		router.mux.ServeHTTP(cssRec, cssReq)
		assert.Equal(t, "css", cssRec.Body.String())

		// Assert JS
		jsReq := httptest.NewRequest(http.MethodGet, "/js/app.js", nil)
		jsRec := httptest.NewRecorder()
		router.mux.ServeHTTP(jsRec, jsReq)
		assert.Equal(t, "js", jsRec.Body.String())
	})
}

func TestRouter_Handlers(t *testing.T) {
	t.Run("returns configured handler", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		router.RegisterPublicHandler(http.MethodGet, "/test", mockHandler("test"))

		// Act
		handler := router.Handlers()

		// Assert
		assert.NotNil(t, handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test", rec.Body.String())
	})

	t.Run("applies base middleware to all routes", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		router.RegisterPublicHandler(http.MethodGet, "/public", mockHandler("public"))
		router.RegisterPrivateHandler(http.MethodGet, "/private", mockHandler("private"))
		router.RegisterUnprotectedHandler(http.MethodGet, "/health", mockHandler("health"))

		// Act
		handler := router.Handlers()

		// Assert - all routes should have base middleware applied
		testCases := []struct {
			path     string
			expected string
		}{
			{"/public", "public"},
			{"/private", "private"},
			{"/health", "health"},
		}

		for _, tc := range testCases {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code, "path: "+tc.path)
			assert.Equal(t, tc.expected, rec.Body.String(), "path: "+tc.path)
		}
	})
}

func TestRouter_Integration(t *testing.T) {
	t.Run("complex routing scenario", func(t *testing.T) {
		// Arrange
		router := createTestRouter()

		// Register various types of routes
		router.RegisterUnprotectedHandler(http.MethodGet, "/health", mockHandler("OK"))
		router.RegisterPublicHandler(http.MethodGet, "/login", mockHandler("login page"))
		router.RegisterPublicHandler(http.MethodPost, "/login", mockHandler("login submit"))
		router.RegisterPrivateHandler(http.MethodGet, "/dashboard", mockHandler("dashboard"))
		router.RegisterPrivateHandler(http.MethodPost, "/expenses", mockHandler("expense created"))
		router.RegisterPrivateHandler(http.MethodDelete, "/expenses/{id}", mockHandler("expense deleted"))

		handler := router.Handlers()

		// Test all routes
		testCases := []struct {
			method       string
			path         string
			expectedBody string
			expectedCode int
		}{
			{http.MethodGet, "/health", "OK", http.StatusOK},
			{http.MethodGet, "/login", "login page", http.StatusOK},
			{http.MethodPost, "/login", "login submit", http.StatusOK},
			{http.MethodGet, "/dashboard", "dashboard", http.StatusOK},
			{http.MethodPost, "/expenses", "expense created", http.StatusOK},
			{http.MethodDelete, "/expenses/123", "expense deleted", http.StatusOK},
		}

		for _, tc := range testCases {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedCode, rec.Code,
				"method: %s, path: %s", tc.method, tc.path)
			assert.Equal(t, tc.expectedBody, rec.Body.String(),
				"method: %s, path: %s", tc.method, tc.path)
		}
	})

	t.Run("route not found returns 404", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		router.RegisterPublicHandler(http.MethodGet, "/exists", mockHandler("exists"))
		handler := router.Handlers()

		// Act
		req := httptest.NewRequest(http.MethodGet, "/not-exists", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		// Arrange
		router := createTestRouter()
		router.RegisterPublicHandler(http.MethodGet, "/only-get", mockHandler("get only"))
		handler := router.Handlers()

		// Act - try POST on GET-only route
		req := httptest.NewRequest(http.MethodPost, "/only-get", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})
}
