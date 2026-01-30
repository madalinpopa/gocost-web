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
	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExpenseHandler_CreateExpense(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("category-id", "cat-123")
		formValues.Set("expense-amount", "10.50")
		formValues.Set("expense-desc", "Lunch")
		formValues.Set("month", "2023-10")
		formValues.Set("payment-status", "unpaid")

		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)
		mockSession.On("GetCurrency", req.Context()).Return("USD")

		expectedSpentAt, _ := time.Parse("2006-01", "2023-10")

		mockExpenseUC.On("Create", req.Context(), mock.MatchedBy(func(r *usecase.CreateExpenseRequest) bool {
			return r.CategoryID == "cat-123" &&
				r.Amount == 10.50 &&
				r.Description == "Lunch" &&
				r.SpentAt.Equal(expectedSpentAt) &&
				r.IsPaid == false &&
				r.PaidAt == nil &&
				r.UserID == userID &&
				r.Currency == "USD"
		})).Return(&usecase.ExpenseResponse{ID: "exp-1"}, nil)

		// Act
		handler.CreateExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Contains(t, rec.Header().Get("HX-Trigger"), "dashboard:refresh")
		mockSession.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("success paid", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("category-id", "cat-123")
		formValues.Set("expense-amount", "100.00")
		formValues.Set("expense-desc", "Rent")
		formValues.Set("month", "2023-10")
		formValues.Set("payment-status", "paid")

		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)
		mockSession.On("GetCurrency", req.Context()).Return("USD")

		expectedSpentAt, _ := time.Parse("2006-01", "2023-10")

		mockExpenseUC.On("Create", req.Context(), mock.MatchedBy(func(r *usecase.CreateExpenseRequest) bool {
			return r.IsPaid == true &&
				r.PaidAt != nil &&
				r.SpentAt.Equal(expectedSpentAt) &&
				r.UserID == userID &&
				r.Currency == "USD"
		})).Return(&usecase.ExpenseResponse{ID: "exp-2"}, nil)

		// Act
		handler.CreateExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Contains(t, rec.Header().Get("HX-Trigger"), "dashboard:refresh")
		mockSession.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("invalid form data", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		formValues := url.Values{}
		// Missing category-id and amount
		formValues.Set("expense-desc", "") // invalid desc if validated? (min len?) Form validation rules?

		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		// Act
		handler.CreateExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "category ID is required")

		mockExpenseUC.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("usecase error - user facing", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("category-id", "cat-123")
		formValues.Set("expense-amount", "10.00")
		formValues.Set("expense-desc", "Lunch")
		formValues.Set("month", "2023-10")
		formValues.Set("payment-status", "unpaid")

		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)
		mockSession.On("GetCurrency", req.Context()).Return("USD")

		expectedErr := tracking.ErrCategoryNotFound
		mockExpenseUC.On("Create", req.Context(), mock.Anything).Return(nil, expectedErr)

		// Act
		handler.CreateExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "Category not found")
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("usecase error - internal", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("category-id", "cat-123")
		formValues.Set("expense-amount", "10.00")
		formValues.Set("expense-desc", "Lunch")
		formValues.Set("month", "2023-10")
		formValues.Set("payment-status", "unpaid")

		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)
		mockSession.On("GetCurrency", req.Context()).Return("USD")

		expectedErr := errors.New("db error")
		mockExpenseUC.On("Create", req.Context(), mock.Anything).Return(nil, expectedErr)

		// Act
		handler.CreateExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "An unexpected error occurred")
		mockExpenseUC.AssertExpectations(t)
	})
}

func TestExpenseHandler_EditExpense(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("expense-id", "exp-123")
		formValues.Set("category-id", "cat-123")
		formValues.Set("edit-amount", "20.00")
		formValues.Set("edit-desc", "Dinner")
		formValues.Set("payment-status", "unpaid")

		req := httptest.NewRequest(http.MethodPost, "/expenses/edit", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)
		mockSession.On("GetCurrency", req.Context()).Return("USD")

		expectedSpentAt, _ := time.Parse("2006-01-02", "2023-10-28")
		mockExpenseUC.On("Get", req.Context(), userID, "exp-123").Return(&usecase.ExpenseResponse{
			ID:      "exp-123",
			SpentAt: expectedSpentAt,
		}, nil)

		mockExpenseUC.On("Update", req.Context(), mock.MatchedBy(func(r *usecase.UpdateExpenseRequest) bool {
			return r.ID == "exp-123" &&
				r.CategoryID == "cat-123" &&
				r.Amount == 20.00 &&
				r.Description == "Dinner" &&
				r.SpentAt.Equal(expectedSpentAt) &&
				r.UserID == userID &&
				r.Currency == "USD"
		})).Return(&usecase.ExpenseResponse{ID: "exp-123"}, nil)

		// Act
		handler.EditExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Contains(t, rec.Header().Get("HX-Trigger"), "dashboard:refresh")
		mockSession.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("invalid form data", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		formValues := url.Values{}
		// Missing ID and amount

		req := httptest.NewRequest(http.MethodPost, "/expenses/edit", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		// Act
		handler.EditExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		mockExpenseUC.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("usecase error - translated", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("expense-id", "exp-123")
		formValues.Set("category-id", "cat-123")
		formValues.Set("edit-amount", "20.00")
		formValues.Set("edit-desc", "Dinner")
		formValues.Set("payment-status", "unpaid")

		req := httptest.NewRequest(http.MethodPost, "/expenses/edit", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)
		mockSession.On("GetCurrency", req.Context()).Return("USD")

		mockExpenseUC.On("Get", req.Context(), userID, "exp-123").Return(&usecase.ExpenseResponse{
			ID:      "exp-123",
			SpentAt: time.Now(),
		}, nil)

		expectedErr := expense.ErrInvalidAmount
		mockExpenseUC.On("Update", req.Context(), mock.Anything).Return(nil, expectedErr)

		// Act
		handler.EditExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "Amount cannot be negative")
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("usecase error - expense not found", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		formValues := url.Values{}
		formValues.Set("expense-id", "exp-123")
		formValues.Set("category-id", "cat-123")
		formValues.Set("edit-amount", "20.00")
		formValues.Set("edit-desc", "Dinner")
		formValues.Set("payment-status", "unpaid")

		req := httptest.NewRequest(http.MethodPost, "/expenses/edit", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		userID := "user-123"
		mockSession.On("GetUserID", req.Context()).Return(userID)

		expectedErr := expense.ErrExpenseNotFound
		mockExpenseUC.On("Get", req.Context(), userID, "exp-123").Return(nil, expectedErr)

		// Act
		handler.EditExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "Expense not found")
		mockExpenseUC.AssertExpectations(t)
	})
}

func TestExpenseHandler_DeleteExpense(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		req := httptest.NewRequest(http.MethodDelete, "/expenses/exp-1", nil)
		req.SetPathValue("id", "exp-1")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")
		mockExpenseUC.On("Delete", req.Context(), "user-123", "exp-1").Return(nil)

		// Act
		handler.DeleteExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Contains(t, rec.Header().Get("HX-Trigger"), "dashboard:refresh")
		mockSession.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("usecase error", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		req := httptest.NewRequest(http.MethodDelete, "/expenses/exp-1", nil)
		req.SetPathValue("id", "exp-1")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")
		expectedErr := errors.New("delete failed")
		mockExpenseUC.On("Delete", req.Context(), "user-123", "exp-1").Return(expectedErr)

		mockErrorHandler.On("Error", rec, req, http.StatusInternalServerError, expectedErr).Return()

		// Act
		handler.DeleteExpense(rec, req)

		// Assert
		mockSession.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
		mockErrorHandler.AssertExpectations(t)
	})
}

func TestExpenseHandler_GetCreateForm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		req := httptest.NewRequest(http.MethodGet, "/expenses/form?category-id=cat-1&month=2023-10", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.GetCreateForm(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Add Expense")
		assert.Contains(t, rec.Body.String(), "cat-1")
		assert.Contains(t, rec.Body.String(), "2023-10")
	})

	t.Run("missing category-id", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		req := httptest.NewRequest(http.MethodGet, "/expenses/form?month=2023-10", nil)
		rec := httptest.NewRecorder()

		mockErrorHandler.On("Error", rec, req, http.StatusBadRequest, mock.Anything).Return()

		// Act
		handler.GetCreateForm(rec, req)

		// Assert
		mockErrorHandler.AssertExpectations(t)
	})

	t.Run("missing month", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(MockExpenseUseCase)
		mockSession := new(MockSessionManager)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewExpenseHandler(appCtx, mockExpenseUC)

		req := httptest.NewRequest(http.MethodGet, "/expenses/form?category-id=cat-1", nil)
		rec := httptest.NewRecorder()

		mockErrorHandler.On("Error", rec, req, http.StatusBadRequest, mock.Anything).Return()

		// Act
		handler.GetCreateForm(rec, req)

		// Assert
		mockErrorHandler.AssertExpectations(t)
	})
}
