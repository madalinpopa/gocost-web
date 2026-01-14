package handler

import (
	"net/http"

	"github.com/madalinpopa/gocost-web/internal/app"
	"github.com/madalinpopa/gocost-web/internal/usecase"
)

type LogoutHandler struct {
	app  app.HandlerContext
	auth usecase.AuthUseCase
}

func NewLogoutHandler(app app.HandlerContext, auth usecase.AuthUseCase) LogoutHandler {
	return LogoutHandler{
		app:  app,
		auth: auth,
	}
}

func (h *LogoutHandler) SubmitLogout(w http.ResponseWriter, r *http.Request) {
	userID := h.app.Session.GetUserID(r.Context())
	h.app.Logger.Info("user logout", "user_id", userID)

	if err := h.app.Session.RenewToken(r.Context()); err != nil {
		h.app.Response.Handle.ServerError(w, r, err)
		return
	}

	if err := h.app.Session.Destroy(r.Context()); err != nil {
		h.app.Response.Handle.ServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}
