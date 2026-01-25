package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRequiredQueryParam(t *testing.T) {
	t.Run("returns value when present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?foo=bar", nil)
		val, err := GetRequiredQueryParam(req, "foo")
		assert.NoError(t, err)
		assert.Equal(t, "bar", val)
	})

	t.Run("returns error when missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		val, err := GetRequiredQueryParam(req, "foo")
		assert.Error(t, err)
		assert.Empty(t, val)
		assert.EqualError(t, err, "required query parameter 'foo' is missing")
	})

	t.Run("returns error when empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?foo=", nil)
		val, err := GetRequiredQueryParam(req, "foo")
		assert.Error(t, err)
		assert.Empty(t, val)
	})
}

func TestGetOptionalQueryParam(t *testing.T) {
	t.Run("returns value when present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?foo=bar", nil)
		val := GetOptionalQueryParam(req, "foo", "default")
		assert.Equal(t, "bar", val)
	})

	t.Run("returns default when missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		val := GetOptionalQueryParam(req, "foo", "default")
		assert.Equal(t, "default", val)
	})

	t.Run("returns default when empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?foo=", nil)
		val := GetOptionalQueryParam(req, "foo", "default")
		assert.Equal(t, "default", val)
	})
}
