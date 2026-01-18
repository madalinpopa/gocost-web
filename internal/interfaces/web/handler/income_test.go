package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIncomeHandler_CreateIncome(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Response: web.Response{
				Handle: mockErrorHandler,
			},
		}

		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("income-amount", "100.50")
		formValues.Set("income-desc", "Salary")
		formValues.Set("income-date", "2023-10-27")

		req := httptest.NewRequest(http.MethodPost, "/incomes", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")

		expectedReq := &usecase.CreateIncomeRequest{
			Amount:     100.50,
			Source:     "Salary",
			ReceivedAt: time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC),
		}

		mockIncomeUC.On("Create", req.Context(), "user-123", mock.MatchedBy(func(r *usecase.CreateIncomeRequest) bool {
			return r.Amount == expectedReq.Amount && r.Source == expectedReq.Source && r.ReceivedAt.Equal(expectedReq.ReceivedAt)
		})).Return(&usecase.IncomeResponse{ID: "inc-1"}, nil)

		// Act
		handler.CreateIncome(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "true", rec.Header().Get("HX-Refresh"))
		mockSession.AssertExpectations(t)
		mockIncomeUC.AssertExpectations(t)
	})

	t.Run("invalid form data", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Response: web.Response{
				Handle: mockErrorHandler,
			},
		}

		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		// Missing amount and date
		formValues := url.Values{}
		formValues.Set("income-desc", "Salary")

		req := httptest.NewRequest(http.MethodPost, "/incomes", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		// Act
		handler.CreateIncome(rec, req)

		// Assert
		// Should render form with errors (status OK)
		assert.Equal(t, http.StatusOK, rec.Code)
		// Basic check that response contains error message
		assert.Contains(t, rec.Body.String(), "amount must be greater than 0")

		mockSession.AssertNotCalled(t, "GetUserID", mock.Anything)
		mockIncomeUC.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Response: web.Response{
				Handle: mockErrorHandler,
			},
		}

		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("income-amount", "100.50")
		formValues.Set("income-desc", "Salary")
		formValues.Set("income-date", "2023-10-27")

		req := httptest.NewRequest(http.MethodPost, "/incomes", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("")

		// Act
		handler.CreateIncome(rec, req)

		// Assert
		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/login", rec.Header().Get("Location"))

		mockSession.AssertExpectations(t)
		mockIncomeUC.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("usecase error", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Response: web.Response{
				Handle: mockErrorHandler,
			},
		}

		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("income-amount", "100.50")
		formValues.Set("income-desc", "Salary")
		formValues.Set("income-date", "2023-10-27")

		req := httptest.NewRequest(http.MethodPost, "/incomes", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")

		expectedErr := errors.New("database error")
		mockIncomeUC.On("Create", req.Context(), "user-123", mock.Anything).Return(nil, expectedErr)

		mockErrorHandler.On("LogServerError", req, expectedErr).Return()

		// Act
		handler.CreateIncome(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		mockSession.AssertExpectations(t)
		mockIncomeUC.AssertExpectations(t)
		mockErrorHandler.AssertExpectations(t)
	})
}

func TestIncomeHandler_ListIncomes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)
		appCtx := HandlerContext{
			Config:   &config.Config{Currency: "$"},
			Session:  mockSession,
			Response: web.Response{Handle: mockErrorHandler},
		}
		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		req := httptest.NewRequest(http.MethodGet, "/incomes?month=2023-10", nil)
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")
		mockIncomeUC.On("ListByMonth", req.Context(), "user-123", "2023-10").Return([]*usecase.IncomeResponse{}, nil)

		handler.ListIncomes(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		mockSession.AssertExpectations(t)
		mockIncomeUC.AssertExpectations(t)
	})
}

func TestIncomeHandler_DeleteIncome(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)
		appCtx := HandlerContext{
			Config:   &config.Config{Currency: "$"},
			Session:  mockSession,
			Response: web.Response{Handle: mockErrorHandler},
		}
		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		req := httptest.NewRequest(http.MethodDelete, "/incomes/inc-1", nil)
		req.SetPathValue("id", "inc-1")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")

		receivedAt, _ := time.Parse("2006-01-02", "2023-10-15")
		mockIncomeUC.On("Get", req.Context(), "user-123", "inc-1").Return(&usecase.IncomeResponse{
			ID:         "inc-1",
			ReceivedAt: receivedAt,
		}, nil)

		mockIncomeUC.On("Delete", req.Context(), "user-123", "inc-1").Return(nil)
		mockIncomeUC.On("Total", req.Context(), "user-123", "2023-10").Return(200.0, nil)
		mockExpenseUC.On("Total", req.Context(), "user-123", "2023-10").Return(50.0, nil)

		handler.DeleteIncome(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		mockSession.AssertExpectations(t)
		mockIncomeUC.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
	})
}
