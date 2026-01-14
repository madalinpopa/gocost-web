package public

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/madalinpopa/gocost-web/internal/app"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/form"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/madalinpopa/gocost-web/ui/templates/pages/public"
)

type LoginHandler struct {
	app  app.HandlerContext
	auth usecase.AuthUseCase
}

func NewLoginHandler(app app.HandlerContext, auth usecase.AuthUseCase) LoginHandler {
	return LoginHandler{
		app:  app,
		auth: auth,
	}
}

func (lh LoginHandler) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	data := lh.app.Template.GetData(r)
	page := public.LoginPage(data)
	lh.app.Template.Render(w, r, page, http.StatusOK)
}

func (lh LoginHandler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	loginForm := form.LoginForm{}
	page := public.LoginForm(loginForm)
	lh.app.Template.Render(w, r, page, http.StatusOK)
}

func (lh LoginHandler) SubmitLoginForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		lh.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	var loginForm form.LoginForm
	if err := lh.app.Decoder.Decode(&loginForm, r.PostForm); err != nil {
		lh.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	loginForm.Validate()
	if !loginForm.IsValid() {
		page := public.LoginForm(loginForm)
		lh.app.Template.Render(w, r, page, http.StatusUnprocessableEntity)
		return
	}

	req := &usecase.LoginRequest{
		EmailOrUsername: loginForm.Email,
		Password:        loginForm.Password,
	}

	resp, err := lh.auth.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			loginForm.AddNonFieldError("Invalid email or password.")
			page := public.LoginForm(loginForm)
			lh.app.Template.Render(w, r, page, http.StatusUnprocessableEntity)
			return
		}
		lh.app.Response.Handle.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	err = lh.app.Session.RenewToken(r.Context())
	if err != nil {
		lh.app.Response.Handle.Error(
			w, r,
			http.StatusInternalServerError,
			fmt.Errorf("failed to renew session token: %w", err),
		)
		return
	}

	lh.app.Session.SetUserID(r.Context(), resp.UserID)
	lh.app.Session.SetUsername(r.Context(), resp.Username)

	lh.app.Response.Htmx.Redirect(w, "/home")
}
