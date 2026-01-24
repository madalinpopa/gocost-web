package respond

import (
	"log/slog"
	"net/http"
)

type RequestLogger struct {
	logger *slog.Logger
}

func NewRequestLogger(l *slog.Logger) RequestLogger {
	return RequestLogger{logger: l}
}

func (l RequestLogger) ErrorWithRequest(r *http.Request, err error, attrs ...any) {
	if err == nil {
		return
	}

	fields := []any{
		"method", r.Method,
		"url", r.URL.RequestURI(),
	}
	fields = append(fields, attrs...)

	l.logger.Error(err.Error(), fields...)
}
