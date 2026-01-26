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
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"
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
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("income-amount", "100.50")
		formValues.Set("income-desc", "Salary")
		formValues.Set("current-month", "2023-10")

		req := httptest.NewRequest(http.MethodPost, "/incomes", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")

		expectedReq := &usecase.CreateIncomeRequest{
			Amount:     100.50,
			Source:     "Salary",
			ReceivedAt: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		}

		mockIncomeUC.On("Create", req.Context(), "user-123", mock.MatchedBy(func(r *usecase.CreateIncomeRequest) bool {
			return r.Amount == expectedReq.Amount && r.Source == expectedReq.Source && r.ReceivedAt.Equal(expectedReq.ReceivedAt)
		})).Return(&usecase.IncomeResponse{ID: "inc-1"}, nil)

		// Act
		handler.CreateIncome(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Contains(t, rec.Header().Get("HX-Trigger"), "dashboard:refresh")
		mockSession.AssertExpectations(t)
		mockIncomeUC.AssertExpectations(t)
	})

	t.Run("invalid form data", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		// Missing amount
		formValues := url.Values{}
		formValues.Set("income-desc", "Salary")
		formValues.Set("current-month", "2023-10")

		req := httptest.NewRequest(http.MethodPost, "/incomes", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		// Act
		handler.CreateIncome(rec, req)

		// Assert
		// Should render form with errors
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		// Basic check that response contains error message
		assert.Contains(t, rec.Body.String(), "amount must be a number")

		mockSession.AssertNotCalled(t, "GetUserID", mock.Anything)
		mockIncomeUC.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("income-amount", "100.50")
		formValues.Set("income-desc", "Salary")
		formValues.Set("current-month", "2023-10")

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
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("income-amount", "100.50")
		formValues.Set("income-desc", "Salary")
		formValues.Set("current-month", "2023-10")

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
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Logger:  logger,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
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

func TestIncomeHandler_GetCreateForm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Logger:  logger,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		req := httptest.NewRequest(http.MethodGet, "/incomes/form?current-month=2023-10", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.GetCreateForm(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Add Income")
		assert.Contains(t, rec.Body.String(), "2023-10")
	})

	t.Run("missing current-month", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Logger:  logger,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		req := httptest.NewRequest(http.MethodGet, "/incomes/form", nil)
		rec := httptest.NewRecorder()

		mockErrorHandler.On("Error", rec, req, http.StatusBadRequest, mock.Anything).Return()

		// Act
		handler.GetCreateForm(rec, req)

		// Assert
		mockErrorHandler.AssertExpectations(t)
	})
}
func TestIncomeHandler_DeleteIncome(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSession := new(MockSessionManager)
		mockIncomeUC := new(MockIncomeUseCase)
		mockExpenseUC := new(MockExpenseUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "$"},
			Session: mockSession,
			Logger:  logger,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}
		handler := NewIncomeHandler(appCtx, mockIncomeUC, mockExpenseUC)

		req := httptest.NewRequest(http.MethodDelete, "/incomes/inc-1", nil)
		req.SetPathValue("id", "inc-1")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")

		mockIncomeUC.On("Delete", req.Context(), "user-123", "inc-1").Return(nil)

		handler.DeleteIncome(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Contains(t, rec.Header().Get("HX-Trigger"), "dashboard:refresh")
		mockSession.AssertExpectations(t)
		mockIncomeUC.AssertExpectations(t)
	})
}
