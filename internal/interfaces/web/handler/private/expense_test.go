package private

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
	"github.com/madalinpopa/gocost-web/internal/app"
	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/handler/mocks"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/response"
	"github.com/madalinpopa/gocost-web/internal/shared/money"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExpenseHandler_CreateExpense(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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

		expectedSpentAt, _ := time.Parse("2006-01", "2023-10")

		mockExpenseUC.On("Create", req.Context(), userID, mock.MatchedBy(func(r *usecase.CreateExpenseRequest) bool {
			return r.CategoryID == "cat-123" &&
				r.Amount == 10.50 &&
				r.Description == "Lunch" &&
				r.SpentAt.Equal(expectedSpentAt) &&
				r.IsPaid == false &&
				r.PaidAt == nil
		})).Return(&usecase.ExpenseResponse{ID: "exp-1"}, nil)

		// Act
		handler.CreateExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "true", rec.Header().Get("HX-Refresh"))
		mockSession.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("success paid", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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

		expectedSpentAt, _ := time.Parse("2006-01", "2023-10")

		mockExpenseUC.On("Create", req.Context(), userID, mock.MatchedBy(func(r *usecase.CreateExpenseRequest) bool {
			return r.IsPaid == true &&
				r.PaidAt != nil &&
				r.SpentAt.Equal(expectedSpentAt)
		})).Return(&usecase.ExpenseResponse{ID: "exp-2"}, nil)

		// Act
		handler.CreateExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		mockSession.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("invalid form data", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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

		expectedErr := tracking.ErrCategoryNotFound
		mockExpenseUC.On("Create", req.Context(), userID, mock.Anything).Return(nil, expectedErr)

		// Act
		handler.CreateExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "Category not found")
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("usecase error - internal", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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

		expectedErr := errors.New("db error")
		mockExpenseUC.On("Create", req.Context(), userID, mock.Anything).Return(nil, expectedErr)

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
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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

		expectedSpentAt, _ := time.Parse("2006-01-02", "2023-10-28")
		mockExpenseUC.On("Get", req.Context(), userID, "exp-123").Return(&usecase.ExpenseResponse{
			ID:      "exp-123",
			SpentAt: expectedSpentAt,
		}, nil)

		mockExpenseUC.On("Update", req.Context(), userID, mock.MatchedBy(func(r *usecase.UpdateExpenseRequest) bool {
			return r.ID == "exp-123" &&
				r.CategoryID == "cat-123" &&
				r.Amount == 20.00 &&
				r.Description == "Dinner" &&
				r.SpentAt.Equal(expectedSpentAt)
		})).Return(&usecase.ExpenseResponse{ID: "exp-123"}, nil)

		// Act
		handler.EditExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "true", rec.Header().Get("HX-Refresh"))
		mockSession.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("invalid form data", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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

		mockExpenseUC.On("Get", req.Context(), userID, "exp-123").Return(&usecase.ExpenseResponse{
			ID:      "exp-123",
			SpentAt: time.Now(),
		}, nil)

		expectedErr := money.ErrNegativeAmount
		mockExpenseUC.On("Update", req.Context(), userID, mock.Anything).Return(nil, expectedErr)

		// Act
		handler.EditExpense(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "Amount cannot be negative")
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("usecase error - expense not found", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Decoder: form.NewDecoder(),
			Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "true", rec.Header().Get("HX-Refresh"))
		mockSession.AssertExpectations(t)
		mockExpenseUC.AssertExpectations(t)
	})

	t.Run("usecase error", func(t *testing.T) {
		// Arrange
		mockExpenseUC := new(mocks.MockExpenseUseCase)
		mockSession := new(mocks.MockSessionManager)
		mockErrorHandler := new(mocks.MockErrorHandler)

		appCtx := app.ApplicationContext{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			Session: mockSession,
			Response: response.Response{
				Handle: mockErrorHandler,
			},
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

