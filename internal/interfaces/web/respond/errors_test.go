package respond_test

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"

	"github.com/stretchr/testify/assert"
)

func TestErrorHandler_ServerError(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	eh := respond.NewErrorHandler(logger)

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	err := errors.New("boom")

	// Act
	eh.ServerError(w, req, err)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "Internal Server Error\n", w.Body.String())
	assert.Contains(t, buf.String(), "boom")
	assert.Contains(t, buf.String(), "trace") // Stack trace should be present
}

func TestErrorHandler_Error(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	eh := respond.NewErrorHandler(logger)

	req := httptest.NewRequest("POST", "/bad-request", nil)
	w := httptest.NewRecorder()
	err := errors.New("invalid input")

	// Act
	eh.Error(w, req, http.StatusBadRequest, err)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "Bad Request\n", w.Body.String())
	assert.Contains(t, buf.String(), "invalid input")
	assert.Contains(t, buf.String(), "status")
	assert.Contains(t, buf.String(), "400")
}
