package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/domain/identity"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestRegisterHandler(session *MockSessionManager, auth *MockAuthUseCase) RegisterHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.New()
	templater := web.NewTemplate(logger, cfg)
	decoder := form.NewDecoder()

	if session == nil {
		session = new(MockSessionManager)
	}
	if auth == nil {
		auth = new(MockAuthUseCase)
	}

	appCtx := HandlerContext{
		Config:   cfg,
		Logger:   logger,
		Decoder:  decoder,
		Session:  session,
		Template: templater,
		Response: web.NewResponse(logger),
	}

	return NewRegisterHandler(appCtx, auth)
}

func TestRegisterHandler_ShowRegisterPage(t *testing.T) {
	t.Run("renders the register page", func(t *testing.T) {
		handler := newTestRegisterHandler(nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/register", nil)
		rec := httptest.NewRecorder()

		handler.ShowRegisterPage(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		body := rec.Body.String()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, body, "<title>Go Cost - Expense Tracker</title>")
		assert.Contains(t, body, "REGISTER")
		assert.Contains(t, body, "CREATE YOUR ACCOUNT")
		assert.Contains(t, body, strconv.Itoa(time.Now().Year()))
	})
}

func TestRegisterHandler_ShowRegisterForm(t *testing.T) {
	t.Run("renders the register form", func(t *testing.T) {
		handler := newTestRegisterHandler(nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/register/form", nil)
		rec := httptest.NewRecorder()

		handler.ShowRegisterForm(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		body := rec.Body.String()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, body, "name=\"email\"")
		assert.Contains(t, body, "name=\"username\"")
		assert.Contains(t, body, "name=\"password\"")
		assert.Contains(t, body, "CREATE ACCOUNT")
	})
}

func TestRegisterHandler_SubmitRegisterForm(t *testing.T) {
	t.Run("successful registration redirects to home", func(t *testing.T) {
		session := new(MockSessionManager)
		auth := new(MockAuthUseCase)
		handler := newTestRegisterHandler(session, auth)

		formData := url.Values{}
		formData.Set("email", "test@example.com")
		formData.Set("username", "testuser")
		formData.Set("password", "password123")

		req := httptest.NewRequest(http.MethodPost, "/register/form", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		expectedReq := &usecase.RegisterUserRequest{
			EmailRequest:    usecase.EmailRequest{Email: "test@example.com"},
			UsernameRequest: usecase.UsernameRequest{Username: "testuser"},
			Password:        "password123",
		}

		auth.On("Register", req.Context(), expectedReq).Return(&usecase.UserResponse{
			ID: "user-1", Username: "testuser",
		}, nil)
		session.On("RenewToken", req.Context()).Return(nil)
		session.On("SetUserID", req.Context(), "user-1").Return()
		session.On("SetUsername", req.Context(), "testuser").Return()

		handler.SubmitRegisterForm(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/home", res.Header.Get("HX-Redirect"))

		session.AssertExpectations(t)
		auth.AssertExpectations(t)
	})

	t.Run("validation error re-renders form", func(t *testing.T) {
		session := new(MockSessionManager)
		auth := new(MockAuthUseCase)
		handler := newTestRegisterHandler(session, auth)

		formData := url.Values{}
		formData.Set("email", "invalid-email")
		formData.Set("username", "")
		formData.Set("password", "short")

		req := httptest.NewRequest(http.MethodPost, "/register/form", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		handler.SubmitRegisterForm(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		body := rec.Body.String()
		assert.Contains(t, body, "please enter a valid e-mail address")
		assert.Contains(t, body, "this field is required")
		assert.Contains(t, body, "password must be at least 8 characters long")

		auth.AssertNotCalled(t, "Register", mock.Anything, mock.Anything)
	})

	t.Run("registration error (user exists) re-renders form with error", func(t *testing.T) {
		session := new(MockSessionManager)
		auth := new(MockAuthUseCase)
		handler := newTestRegisterHandler(session, auth)

		formData := url.Values{}
		formData.Set("email", "exists@example.com")
		formData.Set("username", "exists")
		formData.Set("password", "password123")

		req := httptest.NewRequest(http.MethodPost, "/register/form", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		auth.On("Register", req.Context(), mock.Anything).Return(nil, identity.ErrUserAlreadyExists)

		handler.SubmitRegisterForm(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		body := rec.Body.String()
		assert.Contains(t, body, "An account with this email or username already exists.")

		auth.AssertExpectations(t)
		session.AssertNotCalled(t, "RenewToken", mock.Anything)
	})
}
