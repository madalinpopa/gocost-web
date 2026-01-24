package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogoutHandler(session *MockSessionManager) LogoutHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if session == nil {
		session = new(MockSessionManager)
	}

	errHandler := respond.NewErrorHandler(logger)

	appCtx := HandlerContext{
		Config:  &config.Config{Currency: "$"},
		Logger:  logger,
		Session: session,
		Errors:  errHandler,
		Htmx:    respond.NewHtmx(errHandler),
		Notify:  respond.NewNotify(logger),
	}

	return NewLogoutHandler(appCtx, new(MockAuthUseCase))
}

func TestLogoutHandler_SubmitLogout(t *testing.T) {
	t.Run("redirects to login on success", func(t *testing.T) {
		session := new(MockSessionManager)
		handler := newTestLogoutHandler(session)
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		rec := httptest.NewRecorder()

		session.On("GetUserID", req.Context()).Return("user-1")
		session.On("RenewToken", req.Context()).Return(nil)
		session.On("Destroy", req.Context()).Return(nil)

		handler.SubmitLogout(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, "/login", res.Header.Get("Location"))

		session.AssertExpectations(t)
	})

	t.Run("returns server error when renew fails", func(t *testing.T) {
		session := new(MockSessionManager)
		handler := newTestLogoutHandler(session)
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		rec := httptest.NewRecorder()

		session.On("GetUserID", req.Context()).Return("user-1")
		expectedErr := errors.New("renew failed")
		session.On("RenewToken", req.Context()).Return(expectedErr)

		handler.SubmitLogout(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		session.AssertCalled(t, "RenewToken", req.Context())
		session.AssertNotCalled(t, "Destroy", mock.Anything)
		assert.Empty(t, res.Header.Get("Location"))
	})

	t.Run("returns server error when destroy fails", func(t *testing.T) {
		session := new(MockSessionManager)
		handler := newTestLogoutHandler(session)
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		rec := httptest.NewRecorder()

		session.On("GetUserID", req.Context()).Return("user-1")
		session.On("RenewToken", req.Context()).Return(nil)
		expectedErr := errors.New("destroy failed")
		session.On("Destroy", req.Context()).Return(expectedErr)

		handler.SubmitLogout(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		session.AssertCalled(t, "RenewToken", req.Context())
		session.AssertCalled(t, "Destroy", req.Context())
		assert.Empty(t, res.Header.Get("Location"))
	})
}
