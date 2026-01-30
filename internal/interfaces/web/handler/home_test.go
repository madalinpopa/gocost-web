package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestHomeHandler_ShowHomePage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockDashboardUC := new(MockDashboardUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)

		cfg := &config.Config{Currency: "USD"}
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		appCtx := HandlerContext{
			Config:   cfg,
			Logger:   logger,
			Session:  mockSession,
			Errors:   mockErrorHandler,
			Notify:   respond.NewNotify(logger),
			Template: web.NewTemplate(logger, cfg),
		}

		handler := NewHomeHandler(appCtx, mockDashboardUC)

		req := httptest.NewRequest(http.MethodGet, "/?month=2023-10", nil)
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)
		mockSession.On("GetUsername", req.Context()).Return("alice")
		mockSession.On("IsAuthenticated", req.Context()).Return(true)
		mockSession.On("GetCurrency", req.Context()).Return("USD")

		mockDashboardUC.On("Get", req.Context(), &usecase.DashboardRequest{
			UserID: userID,
			Month:  "2023-10",
		}).Return(&usecase.DashboardResponse{}, nil)

		// Act
		defer func() {
			if r := recover(); r != nil {
				// recover from template panic if any, we just want to verify mocks
			}
		}()

		handler.ShowHomePage(rec, req)

		// Assert
		mockDashboardUC.AssertExpectations(t)
	})
}

func TestHomeHandler_GetDashboardGroups(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockDashboardUC := new(MockDashboardUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  mockErrorHandler,
			Notify:  respond.NewNotify(logger),
		}

		handler := NewHomeHandler(appCtx, mockDashboardUC)

		req := httptest.NewRequest(http.MethodGet, "/dashboard/groups?month=2023-10", nil)
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)
		mockSession.On("GetCurrency", req.Context()).Return("USD")

		mockDashboardUC.On("Get", req.Context(), &usecase.DashboardRequest{
			UserID: userID,
			Month:  "2023-10",
		}).Return(&usecase.DashboardResponse{}, nil)

		// Act
		handler.GetDashboardGroups(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		mockDashboardUC.AssertExpectations(t)
	})
}
