package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-playground/form/v4"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLoginHandler(authMock *MockAuthUseCase, sessionMock *MockSessionManager) LoginHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.New()
	templater := web.NewTemplate(logger, cfg)
	errHandler := respond.NewErrorHandler(logger)

	if sessionMock == nil {
		sessionMock = new(MockSessionManager)
	}

	appCtx := HandlerContext{
		Config:   cfg,
		Logger:   logger,
		Template: templater,
		Errors:   errHandler,
		Htmx:     respond.NewHtmx(errHandler),
		Notify:   respond.NewNotify(logger),
		Decoder:  form.NewDecoder(),
		Session:  sessionMock,
	}

	if authMock == nil {
		authMock = new(MockAuthUseCase)
	}

	return NewLoginHandler(appCtx, authMock)
}

func TestLoginHandler_ShowLoginPage(t *testing.T) {
	t.Run("renders the login page", func(t *testing.T) {
		handler := newTestLoginHandler(nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		rec := httptest.NewRecorder()

		handler.ShowLoginPage(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		body := rec.Body.String()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, body, "<title>Go Cost - Expense Tracker</title>")
		assert.Contains(t, body, "LOGIN")
		assert.Contains(t, body, "ACCESS YOUR DASHBOARD")
	})
}

func TestLoginHandler_ShowLoginForm(t *testing.T) {
	t.Run("renders the login form", func(t *testing.T) {
		handler := newTestLoginHandler(nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/login/form", nil)
		rec := httptest.NewRecorder()

		handler.ShowLoginForm(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		body := rec.Body.String()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, body, "name=\"email\"")
		assert.Contains(t, body, "name=\"password\"")
		assert.Contains(t, body, "AUTHENTICATE")
	})
}

func TestLoginHandler_SubmitLoginForm(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		authMock := new(MockAuthUseCase)
		sessionMock := new(MockSessionManager)
		handler := newTestLoginHandler(authMock, sessionMock)

		formVals := url.Values{}
		formVals.Add("email", "test@example.com")
		formVals.Add("password", "password123")

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(formVals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		expectedReq := &usecase.LoginRequest{
			EmailOrUsername: "test@example.com",
			Password:        "password123",
		}

		authMock.On("Login", req.Context(), expectedReq).Return(&usecase.LoginResponse{
			UserID:   "user-123",
			Username: "testuser",
			Currency: "USD",
		}, nil)

		sessionMock.On("RenewToken", req.Context()).Return(nil)
		sessionMock.On("SetUserID", req.Context(), "user-123").Return()
		sessionMock.On("SetUsername", req.Context(), "testuser").Return()
		sessionMock.On("SetCurrency", req.Context(), "USD").Return()

		handler.SubmitLoginForm(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/home", res.Header.Get("HX-Redirect"))

		authMock.AssertExpectations(t)
		sessionMock.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		authMock := new(MockAuthUseCase)
		sessionMock := new(MockSessionManager)
		handler := newTestLoginHandler(authMock, sessionMock)

		formVals := url.Values{}
		formVals.Add("email", "test@example.com")
		formVals.Add("password", "wrongpassword")

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(formVals.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		authMock.On("Login", req.Context(), mock.Anything).Return(nil, usecase.ErrInvalidCredentials)

		handler.SubmitLoginForm(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		body := rec.Body.String()
		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		assert.Contains(t, body, "Invalid email or password.")

		authMock.AssertExpectations(t)
		sessionMock.AssertNotCalled(t, "RenewToken", mock.Anything)
	})
}
