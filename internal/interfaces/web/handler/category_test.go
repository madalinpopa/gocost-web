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

	"github.com/go-playground/form/v4"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCategoryHandler_CreateCategory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockCategoryUC := new(MockCategoryUseCase)
		mockErrorHandler := new(MockErrorHandler)
		mockSession := new(MockSessionManager)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewCategoryHandler(appCtx, mockCategoryUC)

		formValues := url.Values{}
		formValues.Set("group-id", "group-123")
		formValues.Set("category-name", "Test Category")
		formValues.Set("category-desc", "Test Description")
		formValues.Set("type", "monthly")
		formValues.Set("category-start", "2023-01")
		formValues.Set("category-budget", "100.00")

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		expectedReq := &usecase.CreateCategoryRequest{
			Name:        "Test Category",
			Description: "Test Description",
			IsRecurrent: false,
			StartMonth:  "2023-01",
			Budget:      100.0,
		}

		mockSession.On("GetUserID", req.Context()).Return("user-123")
		mockSession.On("GetCurrency", req.Context()).Return("USD")

		mockCategoryUC.On("Create", req.Context(), mock.MatchedBy(func(r *usecase.CreateCategoryRequest) bool {
			return r.Name == expectedReq.Name &&
				r.Description == expectedReq.Description &&
				r.IsRecurrent == expectedReq.IsRecurrent &&
				r.StartMonth == expectedReq.StartMonth &&
				r.UserID == "user-123" &&
				r.GroupID == "group-123" &&
				r.Currency == "USD"
		})).Return(&usecase.CategoryResponse{ID: "cat-1"}, nil)

		// Act
		handler.CreateCategory(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Contains(t, rec.Header().Get("HX-Trigger"), "dashboard:refresh")
		mockCategoryUC.AssertExpectations(t)
		mockSession.AssertExpectations(t)
	})

	t.Run("invalid form data", func(t *testing.T) {
		// Arrange
		mockCategoryUC := new(MockCategoryUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewCategoryHandler(appCtx, mockCategoryUC)

		// Missing group-id and name
		formValues := url.Values{}
		// group-id missing
		formValues.Set("category-name", "") // empty name

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		// Act
		handler.CreateCategory(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "group ID is required")
		assert.Contains(t, rec.Body.String(), "this field is required") // Name error

		mockCategoryUC.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("usecase error - user facing", func(t *testing.T) {
		// Arrange
		mockCategoryUC := new(MockCategoryUseCase)
		mockErrorHandler := new(MockErrorHandler)
		mockSession := new(MockSessionManager)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewCategoryHandler(appCtx, mockCategoryUC)

		formValues := url.Values{}
		formValues.Set("group-id", "group-123")
		formValues.Set("category-name", "Test Category")
		formValues.Set("category-desc", "Test Description")
		formValues.Set("type", "monthly")
		formValues.Set("category-start", "2023-01")
		formValues.Set("category-budget", "100.00")

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		expectedErr := tracking.ErrCategoryNameExists
		mockSession.On("GetUserID", req.Context()).Return("user-123")
		mockSession.On("GetCurrency", req.Context()).Return("USD")
		mockCategoryUC.On("Create", req.Context(), mock.Anything).Return(nil, expectedErr)

		// Act
		handler.CreateCategory(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "Category name already exists") // Translated error

		mockCategoryUC.AssertExpectations(t)
		mockSession.AssertExpectations(t)
	})

	t.Run("usecase error - internal", func(t *testing.T) {
		// Arrange
		mockCategoryUC := new(MockCategoryUseCase)
		mockErrorHandler := new(MockErrorHandler)
		mockSession := new(MockSessionManager)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Decoder: form.NewDecoder(),
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewCategoryHandler(appCtx, mockCategoryUC)

		formValues := url.Values{}
		formValues.Set("group-id", "group-123")
		formValues.Set("category-name", "Test Category")
		formValues.Set("category-desc", "Test Description")
		formValues.Set("type", "monthly")
		formValues.Set("category-start", "2023-01")
		formValues.Set("category-budget", "100.00")

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		expectedErr := errors.New("database failure")
		mockSession.On("GetUserID", req.Context()).Return("user-123")
		mockSession.On("GetCurrency", req.Context()).Return("USD")
		mockCategoryUC.On("Create", req.Context(), mock.Anything).Return(nil, expectedErr)

		// Act
		handler.CreateCategory(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "An unexpected error occurred") // Default translated error

		mockCategoryUC.AssertExpectations(t)
		mockSession.AssertExpectations(t)
	})
}

func TestCategoryHandler_DeleteCategory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockCategoryUC := new(MockCategoryUseCase)
		mockErrorHandler := new(MockErrorHandler)
		mockSession := new(MockSessionManager)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewCategoryHandler(appCtx, mockCategoryUC)

		req := httptest.NewRequest(http.MethodDelete, "/groups/group-1/categories/cat-1", nil)
		req.SetPathValue("groupID", "group-1")
		req.SetPathValue("id", "cat-1")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")
		mockCategoryUC.On("Delete", req.Context(), "user-123", "group-1", "cat-1").Return(nil)

		// Act
		handler.DeleteCategory(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Contains(t, rec.Header().Get("HX-Trigger"), "dashboard:refresh")
		mockCategoryUC.AssertExpectations(t)
		mockSession.AssertExpectations(t)
	})

	t.Run("usecase error", func(t *testing.T) {
		// Arrange
		mockCategoryUC := new(MockCategoryUseCase)
		mockErrorHandler := new(MockErrorHandler)
		mockSession := new(MockSessionManager)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewCategoryHandler(appCtx, mockCategoryUC)

		req := httptest.NewRequest(http.MethodDelete, "/groups/group-1/categories/cat-1", nil)
		req.SetPathValue("groupID", "group-1")
		req.SetPathValue("id", "cat-1")
		rec := httptest.NewRecorder()

		expectedErr := errors.New("delete failed")
		mockSession.On("GetUserID", req.Context()).Return("user-123")
		mockCategoryUC.On("Delete", req.Context(), "user-123", "group-1", "cat-1").Return(expectedErr)

		mockErrorHandler.On("Error", rec, req, http.StatusInternalServerError, expectedErr).Return()

		// Act
		handler.DeleteCategory(rec, req)

		// Assert
		mockCategoryUC.AssertExpectations(t)
		mockErrorHandler.AssertExpectations(t)
		mockSession.AssertExpectations(t)
	})
}

func TestCategoryHandler_GetCreateForm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockCategoryUC := new(MockCategoryUseCase)
		mockErrorHandler := new(MockErrorHandler)
		mockSession := new(MockSessionManager)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewCategoryHandler(appCtx, mockCategoryUC)

		req := httptest.NewRequest(http.MethodGet, "/categories/form?group-id=group-1&category-start=2023-01", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.GetCreateForm(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Add Category")
		assert.Contains(t, rec.Body.String(), "group-1")
		assert.Contains(t, rec.Body.String(), "2023-01")
	})

	t.Run("missing group-id", func(t *testing.T) {
		// Arrange
		mockCategoryUC := new(MockCategoryUseCase)
		mockErrorHandler := new(MockErrorHandler)
		mockSession := new(MockSessionManager)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewCategoryHandler(appCtx, mockCategoryUC)

		req := httptest.NewRequest(http.MethodGet, "/categories/form?category-start=2023-01", nil)
		rec := httptest.NewRecorder()

		mockErrorHandler.On("Error", rec, req, http.StatusBadRequest, mock.Anything).Return()

		// Act
		handler.GetCreateForm(rec, req)

		// Assert
		mockErrorHandler.AssertExpectations(t)
	})

	t.Run("missing category-start", func(t *testing.T) {
		// Arrange
		mockCategoryUC := new(MockCategoryUseCase)
		mockErrorHandler := new(MockErrorHandler)
		mockSession := new(MockSessionManager)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config:  &config.Config{Currency: "USD"},
			Logger:  logger,
			Session: mockSession,
			Errors:  newTestErrors(logger, mockErrorHandler),
			Notify:  respond.NewNotify(logger),
		}

		handler := NewCategoryHandler(appCtx, mockCategoryUC)

		req := httptest.NewRequest(http.MethodGet, "/categories/form?group-id=group-1", nil)
		rec := httptest.NewRecorder()

		mockErrorHandler.On("Error", rec, req, http.StatusBadRequest, mock.Anything).Return()

		// Act
		handler.GetCreateForm(rec, req)

		// Assert
		mockErrorHandler.AssertExpectations(t)
	})
}
