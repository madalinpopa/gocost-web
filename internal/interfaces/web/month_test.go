package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMonthParam(t *testing.T) {
	t.Run("valid month", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?month=2025-05", nil)
		current, prev, next := GetMonthParam(req)

		assert.Equal(t, "2025-05-01", current.Format("2006-01-02"))
		assert.Equal(t, "2025-04-01", prev.Format("2006-01-02"))
		assert.Equal(t, "2025-06-01", next.Format("2006-01-02"))
	})

	t.Run("invalid month defaults to now", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?month=invalid", nil)
		current, prev, next := GetMonthParam(req)

		now := time.Now()
		assert.WithinDuration(t, now, current, 1*time.Second)

		expectedPrev := current.AddDate(0, -1, 0)
		expectedNext := current.AddDate(0, 1, 0)

		assert.Equal(t, expectedPrev, prev)
		assert.Equal(t, expectedNext, next)
	})

	t.Run("empty month defaults to now", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		current, _, _ := GetMonthParam(req)
		assert.WithinDuration(t, time.Now(), current, 1*time.Second)
	})
}
