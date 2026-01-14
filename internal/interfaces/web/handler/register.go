package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/madalinpopa/gocost-web/internal/domain/identity"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/form"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/madalinpopa/gocost-web/ui/templates/pages/public"
)

type RegisterHandler struct {
	app  HandlerContext
	auth usecase.AuthUseCase
}

func NewRegisterHandler(app HandlerContext, auth usecase.AuthUseCase) RegisterHandler {
	return RegisterHandler{
		app:  app,
		auth: auth,
	}
}

func (rh RegisterHandler) ShowRegisterPage(w http.ResponseWriter, r *http.Request) {
	data := rh.app.Template.GetData(r)
	page := public.RegisterPage(data)
	rh.app.Template.Render(w, r, page, http.StatusOK)
}

func (rh RegisterHandler) ShowRegisterForm(w http.ResponseWriter, r *http.Request) {
	registerForm := form.RegisterForm{}
	page := public.RegisterForm(registerForm)
	rh.app.Template.Render(w, r, page, http.StatusOK)
}

func (rh RegisterHandler) SubmitRegisterForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		rh.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	var registerForm form.RegisterForm
	if err := rh.app.Decoder.Decode(&registerForm, r.PostForm); err != nil {
		rh.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	registerForm.Validate()
	if !registerForm.IsValid() {
		component := public.RegisterForm(registerForm)
		rh.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
		return
	}

	registerRequest := usecase.RegisterUserRequest{
		EmailRequest:    usecase.EmailRequest{Email: registerForm.Email},
		UsernameRequest: usecase.UsernameRequest{Username: registerForm.Username},
		Password:        registerForm.Password,
	}

	// Register the user
	userResponse, err := rh.auth.Register(r.Context(), &registerRequest)
	if err != nil {
		errMessage, isUserFacing := translateError(err)
		registerForm.AddNonFieldError(errMessage)
		component := public.RegisterForm(registerForm)
		rh.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
		if !isUserFacing {
			rh.app.Response.Handle.Error(w, r, http.StatusInternalServerError, err)
		}
		return
	}

	// Renew session token
	err = rh.app.Session.RenewToken(r.Context())
	if err != nil {
		rh.app.Response.Handle.Error(
			w, r,
			http.StatusInternalServerError,
			fmt.Errorf("failed to renew session token: %w", err),
		)
		return
	}

	// Register the user in the session
	rh.app.Session.SetUserID(r.Context(), userResponse.ID)
	rh.app.Session.SetUsername(r.Context(), userResponse.Username)

	// Redirect to admin home
	rh.app.Response.Htmx.Redirect(w, "/home")
}

func translateError(err error) (string, bool) {
	switch {
	case errors.Is(err, identity.ErrUserAlreadyExists):
		return "An account with this email or username already exists.", true
	case errors.Is(err, identity.ErrInvalidEmailFormat):
		return "Please enter a valid email address.", true
	case errors.Is(err, identity.ErrPasswordTooShort):
		return "Password must be at least 8 characters long.", true
	default:
		return "An unexpected error occurred. Please try again later.", false
	}
}
