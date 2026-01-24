package web

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type notify struct {
	logger *slog.Logger
}

func newNotify(l *slog.Logger) notify {
	return notify{
		logger: l,
	}
}

func ToastEvent(t ToastType, message string) map[string]any {
	return map[string]any{
		"showToast": map[string]string{
			"level":   string(t),
			"message": message,
		},
	}
}

func (n notify) Trigger(w http.ResponseWriter, events map[string]any) {
	if len(events) == 0 {
		return
	}

	// Create a notification payload matching the required structure
	notificationJSON, err := json.Marshal(events)
	if err != nil {
		n.logger.Error("failed to marshal notification", "error", err.Error())
		return
	}

	// Set the HX-Trigger header with the JSON content
	w.Header().Set("HX-Trigger", string(notificationJSON))
}

func (n notify) Toast(w http.ResponseWriter, t ToastType, message string) {
	n.Trigger(w, ToastEvent(t, message))
}
