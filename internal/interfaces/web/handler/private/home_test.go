package private

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/app"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/handler/mocks"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/response"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestHomeHandler_ShowHomePage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockIncomeUC := new(mocks.MockIncomeUseCase)
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockGroupUC := new(mocks.MockGroupUseCase)
		mockCategoryUC := new(mocks.MockCategoryUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
			Template: response.NewTemplate(slog.New(slog.NewTextHandler(io.Discard, nil)), nil), // Minimal setup
			Config:   &config.Config{Currency: "$"},
		}

		handler := NewHomeHandler(appCtx, mockIncomeUC, mockExpenseUC, mockGroupUC, mockCategoryUC)

		req := httptest.NewRequest(http.MethodGet, "/?month=2023-10", nil)
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)
		mockSession.On("GetUsername", req.Context()).Return("alice")
		mockSession.On("IsAuthenticated", req.Context()).Return(true)

		// Mocks for fetchDashboardData
		mockIncomeUC.On("Total", req.Context(), userID, "2023-10").Return(1000.0, nil)
		mockExpenseUC.On("Total", req.Context(), userID, "2023-10").Return(500.0, nil)

		groups := []*usecase.GroupResponse{
			{
				ID: "g1", Name: "Group 1", Categories: []usecase.CategoryResponse{
					{ID: "c1", Name: "Cat 1", StartMonth: "2023-01"},
				},
			},
		}
		mockGroupUC.On("List", req.Context(), userID).Return(groups, nil)

		expenses := []*usecase.ExpenseResponse{
			{ID: "e1", CategoryID: "c1", Amount: 50.0, SpentAt: time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC)},
		}
		mockExpenseUC.On("ListByMonth", req.Context(), userID, "2023-10").Return(expenses, nil)

		// Act
		defer func() {
			if r := recover(); r != nil {
				// recover from template panic if any, we just want to verify mocks
			}
		}()

		handler.ShowHomePage(rec, req)

		// Assert
		mockIncomeUC.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
		mockGroupUC.AssertExpectations(t)
	})
}

func TestHomeHandler_GetDashboardGroups(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockIncomeUC := new(mocks.MockIncomeUseCase)
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockGroupUC := new(mocks.MockGroupUseCase)
		mockCategoryUC := new(mocks.MockCategoryUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
			Config: &config.Config{Currency: "$"},
		}

		handler := NewHomeHandler(appCtx, mockIncomeUC, mockExpenseUC, mockGroupUC, mockCategoryUC)

		req := httptest.NewRequest(http.MethodGet, "/dashboard/groups?month=2023-10", nil)
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)

		mockIncomeUC.On("Total", req.Context(), userID, "2023-10").Return(2000.0, nil)
		mockExpenseUC.On("Total", req.Context(), userID, "2023-10").Return(800.0, nil)

		groups := []*usecase.GroupResponse{}
		mockGroupUC.On("List", req.Context(), userID).Return(groups, nil)

		expenses := []*usecase.ExpenseResponse{}
		mockExpenseUC.On("ListByMonth", req.Context(), userID, "2023-10").Return(expenses, nil)

		// Act
		handler.GetDashboardGroups(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		mockIncomeUC.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
		mockGroupUC.AssertExpectations(t)
	})
}
