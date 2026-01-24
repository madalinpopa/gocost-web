package handler

import (
	"log/slog"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"
)

func newTestErrors(logger *slog.Logger, errHandler respond.ErrorHandler) respond.ErrorHandler {
	if errHandler != nil {
		return errHandler
	}
	return respond.NewErrorHandler(logger)
}
