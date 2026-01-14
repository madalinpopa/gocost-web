package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/app"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/response"
	"github.com/stretchr/testify/assert"
)

func newTestIndexHandler() IndexHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	templater := response.NewTemplate(logger, config.New())
	appCtx := app.HandlerContext{Template: templater}

	return NewIndexHandler(appCtx)
}

func TestIndexHandler_ShowIndexPage(t *testing.T) {
	t.Run("renders the index page", func(t *testing.T) {
		handler := newTestIndexHandler()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ShowIndexPage(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		body := rec.Body.String()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, body, "<title>Go Cost - Expense Tracker</title>")
		assert.Contains(t, body, "OWN")
		assert.Contains(t, body, "YOUR")
		assert.Contains(t, body, "COST")
	})
}
