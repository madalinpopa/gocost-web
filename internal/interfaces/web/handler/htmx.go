package handler

import (
	"net/http"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
)

func triggerDashboardRefresh(w http.ResponseWriter, notify web.NotifyHandler, toastType web.ToastType, message string, closeModalID string) {
	if notify == nil {
		return
	}

	events := web.ToastEvent(toastType, message)
	events["dashboard:refresh"] = true
	if closeModalID != "" {
		events["close-modal"] = map[string]string{"id": closeModalID}
	}

	notify.Trigger(w, events)
}
