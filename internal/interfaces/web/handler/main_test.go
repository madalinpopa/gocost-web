package handler

import (
	"log/slog"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
)

func newTestResponse(logger *slog.Logger, errHandler web.ErrorHandler) web.Response {
	resp := web.NewResponse(logger)
	if errHandler != nil {
		resp.Handle = errHandler
	}
	return resp
}
