package respond_test

import (
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"

	"github.com/stretchr/testify/assert"
)

func TestNotify_Trigger(t *testing.T) {
	// Arrange
	n := respond.NewNotify(slog.Default())
	w := httptest.NewRecorder()
	events := map[string]any{
		"event1": "data1",
		"event2": 123,
	}

	// Act
	n.Trigger(w, events)

	// Assert
	val := w.Header().Get("HX-Trigger")
	assert.NotEmpty(t, val)

	var decoded map[string]any
	err := json.Unmarshal([]byte(val), &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "data1", decoded["event1"])
	assert.Equal(t, float64(123), decoded["event2"]) // JSON numbers are floats
}

func TestNotify_Toast(t *testing.T) {
	// Arrange
	n := respond.NewNotify(slog.Default())
	w := httptest.NewRecorder()

	// Act
	n.Toast(w, respond.ErrorMsg, "Something went wrong")

	// Assert
	val := w.Header().Get("HX-Trigger")
	assert.NotEmpty(t, val)

	var decoded map[string]any
	err := json.Unmarshal([]byte(val), &decoded)
	assert.NoError(t, err)

	toastData, ok := decoded["showToast"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "error", toastData["level"])
	assert.Equal(t, "Something went wrong", toastData["message"])
}
