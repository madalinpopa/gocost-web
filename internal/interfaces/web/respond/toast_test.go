package respond_test

import (
	"testing"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"

	"github.com/stretchr/testify/assert"
)

func TestToastEvent(t *testing.T) {
	// Act
	event := respond.ToastEvent(respond.Success, "Operation successful")

	// Assert
	assert.Contains(t, event, "showToast")
	payload, ok := event["showToast"].(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, "success", payload["level"])
	assert.Equal(t, "Operation successful", payload["message"])
}
