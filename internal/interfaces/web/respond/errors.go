package respond

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

type errorHandler struct {
	logger RequestLogger
}

func NewErrorHandler(logger *slog.Logger) ErrorHandler {
	return errorHandler{
		logger: NewRequestLogger(logger),
	}
}

func (e errorHandler) LogServerError(r *http.Request, err error) {
	e.logger.ErrorWithRequest(r, err, "trace", string(debug.Stack()))
}

func (e errorHandler) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	e.LogServerError(r, err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (e errorHandler) Error(w http.ResponseWriter, r *http.Request, status int, err error) {
	e.logger.ErrorWithRequest(r, err, "status", status)
	http.Error(w, http.StatusText(status), status)
}
