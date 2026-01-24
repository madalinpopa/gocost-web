package respond_test

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"

	"github.com/stretchr/testify/assert"
)

func TestRequestLogger_ErrorWithRequest(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	reqLogger := respond.NewRequestLogger(logger)

	req := httptest.NewRequest("GET", "/test-url", nil)
	err := errors.New("test error")

	// Act
	reqLogger.ErrorWithRequest(req, err, "custom_key", "custom_value")

	// Assert
	logOutput := buf.String()
	assert.Contains(t, logOutput, "test error")
	assert.Contains(t, logOutput, "/test-url")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "custom_key")
	assert.Contains(t, logOutput, "custom_value")
}

func TestRequestLogger_ErrorWithRequest_NoError(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	reqLogger := respond.NewRequestLogger(logger)

	req := httptest.NewRequest("GET", "/test-url", nil)

	// Act
	reqLogger.ErrorWithRequest(req, nil)

	// Assert
	assert.Empty(t, buf.String())
}
