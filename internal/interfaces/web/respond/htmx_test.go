package respond_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"

	"github.com/stretchr/testify/assert"
)

func TestHtmx_Redirect(t *testing.T) {
	// Arrange
	h := respond.NewHtmx(nil) // ErrorHandler not needed for Redirect
	w := httptest.NewRecorder()
	url := "/new-location"

	// Act
	h.Redirect(w, url)

	// Assert
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, url, w.Header().Get("HX-Redirect"))
}

func TestHtmx_Location(t *testing.T) {
	// Arrange
	h := respond.NewHtmx(nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	url := "/target-path"
	target := "#content"
	swap := "outerHTML"

	// Act
	h.Location(w, req, url, target, swap)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	val := w.Header().Get("HX-Location")
	assert.NotEmpty(t, val)

	var locMap map[string]string
	err := json.Unmarshal([]byte(val), &locMap)
	assert.NoError(t, err)
	assert.Equal(t, url, locMap["path"])
	assert.Equal(t, target, locMap["target"])
	assert.Equal(t, swap, locMap["swap"])
}

func TestHtmx_Location_Simple(t *testing.T) {
	// Arrange
	h := respond.NewHtmx(nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	url := "/simple"

	// Act
	h.Location(w, req, url, "", "")

	// Assert
	val := w.Header().Get("HX-Location")
	var locMap map[string]string
	_ = json.Unmarshal([]byte(val), &locMap)
	assert.Equal(t, url, locMap["path"])
	assert.NotContains(t, locMap, "target")
	assert.NotContains(t, locMap, "swap")
}
