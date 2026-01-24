package respond

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type notify struct {
	logger *slog.Logger
}

func NewNotify(l *slog.Logger) NotifyHandler {
	return notify{
		logger: l,
	}
}

func (n notify) Trigger(w http.ResponseWriter, events map[string]any) {
	if len(events) == 0 {
		return
	}

	notificationJSON, err := json.Marshal(events)
	if err != nil {
		n.logger.Error("failed to marshal notification", "error", err.Error())
		return
	}

	w.Header().Set("HX-Trigger", string(notificationJSON))
}

func (n notify) Toast(w http.ResponseWriter, t ToastType, message string) {
	n.Trigger(w, ToastEvent(t, message))
}
