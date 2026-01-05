package response

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

func (n notify) Toast(w http.ResponseWriter, t ToastType, message string) {

	// Create a notification payload matching the required structure
	notification := map[string]any{
		"showToast": map[string]string{
			"level":   string(t),
			"message": message,
		},
	}

	// Convert the notification to JSON
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		n.logger.Error("failed to marshal notification", "error", err.Error())
		return
	}

	// Set the HX-Trigger header with the JSON content
	w.Header().Set("HX-Trigger", string(notificationJSON))
}
