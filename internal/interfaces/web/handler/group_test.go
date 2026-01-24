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
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGroupHandler_CreateGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockGroupUC := new(MockGroupUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config: &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:   logger,
			Response: newTestResponse(logger, mockErrorHandler),
		}

		handler := NewGroupHandler(appCtx, mockGroupUC)

		formValues := url.Values{}
		formValues.Set("group-name", "Test Group")
		formValues.Set("group-desc", "Test Description")

		req := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")

		expectedReq := &usecase.CreateGroupRequest{
			Name:        "Test Group",
			Description: "Test Description",
		}

		mockGroupUC.On("Create", req.Context(), "user-123", mock.MatchedBy(func(r *usecase.CreateGroupRequest) bool {
			return r.Name == expectedReq.Name && r.Description == expectedReq.Description
		})).Return(&usecase.GroupResponse{ID: "group-1"}, nil)

		// Act
		handler.CreateGroup(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Contains(t, rec.Header().Get("HX-Trigger"), "dashboard:refresh")
		mockSession.AssertExpectations(t)
		mockGroupUC.AssertExpectations(t)
	})

	t.Run("invalid form data", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockGroupUC := new(MockGroupUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config: &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:   logger,
			Response: newTestResponse(logger, mockErrorHandler),
		}

		handler := NewGroupHandler(appCtx, mockGroupUC)

		// Empty name
		formValues := url.Values{}
		formValues.Set("group-name", "")
		formValues.Set("group-desc", "Description")

		req := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		// Act
		handler.CreateGroup(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "this field is required") // Validation message

		mockSession.AssertNotCalled(t, "GetUserID", mock.Anything)
		mockGroupUC.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("usecase error - user facing", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockGroupUC := new(MockGroupUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config: &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:   logger,
			Response: newTestResponse(logger, mockErrorHandler),
		}

		handler := NewGroupHandler(appCtx, mockGroupUC)

		formValues := url.Values{}
		formValues.Set("group-name", "Test Group")
		formValues.Set("group-desc", "Description")

		req := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")

		expectedErr := tracking.ErrNameTooLong
		mockGroupUC.On("Create", req.Context(), "user-123", mock.Anything).Return(nil, expectedErr)

		// Act
		handler.CreateGroup(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "Group name is too long") // Translated error

		mockSession.AssertExpectations(t)
		mockGroupUC.AssertExpectations(t)
	})

	t.Run("usecase error - internal", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockGroupUC := new(MockGroupUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config: &config.Config{Currency: "$"},
			Session: mockSession,
			Decoder: form.NewDecoder(),
			Logger:   logger,
			Response: newTestResponse(logger, mockErrorHandler),
		}

		handler := NewGroupHandler(appCtx, mockGroupUC)

		formValues := url.Values{}
		formValues.Set("group-name", "Test Group")
		formValues.Set("group-desc", "Description")

		req := httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(formValues.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")

		expectedErr := errors.New("database failure")
		mockGroupUC.On("Create", req.Context(), "user-123", mock.Anything).Return(nil, expectedErr)

		// Act
		handler.CreateGroup(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
		assert.Contains(t, rec.Body.String(), "An unexpected error occurred") // Default translated error

		mockSession.AssertExpectations(t)
		mockGroupUC.AssertExpectations(t)
	})
}

func TestGroupHandler_DeleteGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockGroupUC := new(MockGroupUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config: &config.Config{Currency: "$"},
			Session: mockSession,
			Logger:   logger,
			Response: newTestResponse(logger, mockErrorHandler),
		}

		handler := NewGroupHandler(appCtx, mockGroupUC)

		req := httptest.NewRequest(http.MethodDelete, "/groups/group-1", nil)
		req.SetPathValue("id", "group-1")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")
		mockGroupUC.On("Delete", req.Context(), "user-123", "group-1").Return(nil)

		// Act
		handler.DeleteGroup(rec, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Contains(t, rec.Header().Get("HX-Trigger"), "dashboard:refresh")
		mockSession.AssertExpectations(t)
		mockGroupUC.AssertExpectations(t)
	})

	t.Run("usecase error", func(t *testing.T) {
		// Arrange
		mockSession := new(MockSessionManager)
		mockGroupUC := new(MockGroupUseCase)
		mockErrorHandler := new(MockErrorHandler)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		appCtx := HandlerContext{
			Config: &config.Config{Currency: "$"},
			Session: mockSession,
			Logger:   logger,
			Response: newTestResponse(logger, mockErrorHandler),
		}

		handler := NewGroupHandler(appCtx, mockGroupUC)

		req := httptest.NewRequest(http.MethodDelete, "/groups/group-1", nil)
		req.SetPathValue("id", "group-1")
		rec := httptest.NewRecorder()

		mockSession.On("GetUserID", req.Context()).Return("user-123")
		expectedErr := errors.New("delete failed")
		mockGroupUC.On("Delete", req.Context(), "user-123", "group-1").Return(expectedErr)

		mockErrorHandler.On("Error", rec, req, http.StatusInternalServerError, expectedErr).Return()

		// Act
		handler.DeleteGroup(rec, req)

		// Assert
		mockSession.AssertExpectations(t)
		mockGroupUC.AssertExpectations(t)
		mockErrorHandler.AssertExpectations(t)
	})
}
