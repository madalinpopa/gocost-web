package response

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

type resError struct {
	logger *slog.Logger
	Templater
}

func (re resError) LogServerError(r *http.Request, err error) {
	var (
		method = r.Method
		url    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	re.logger.Error(err.Error(), "method", method, "url", url, "trace", trace)
}

func (re resError) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	re.LogServerError(r, err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (re resError) Error(w http.ResponseWriter, r *http.Request, status int, err error) {
	var (
		method = r.Method
		url    = r.URL.RequestURI()
	)
	re.logger.Error(err.Error(), "method", method, "url", url, "status", status)
	http.Error(w, http.StatusText(status), status)
}

func newErrorHandlers(logger *slog.Logger) ErrorHandler {
	return resError{
		logger: logger,
	}
}
