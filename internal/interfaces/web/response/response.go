package response

import (
	"log/slog"
)

type Response struct {
	logger *slog.Logger
	Handle ErrorHandler
	Htmx   HtmxHandler
	Notify NotifyHandler
}

func NewResponse(logger *slog.Logger) Response {

	errHandler := newErrorHandlers(logger)
	notifier := newNotify(logger)
	htmxRes := newHtmx(logger, errHandler)
	return Response{
		logger: logger,
		Notify: notifier,
		Handle: errHandler,
		Htmx:   htmxRes,
	}
}
