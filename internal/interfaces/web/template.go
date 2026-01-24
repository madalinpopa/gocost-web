package web

import (
	"bytes"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
	"github.com/madalinpopa/gocost-web/internal/config"
)

type Data struct {
	Title       string
	CurrentYear int
	CSRFToken   string
	Toast       *ToastMessage
	User        AuthenticatedUser
	Version     string
	Currency    string
}

func (d *Data) SetToast(toastType ToastType, message string) {
	id := uuid.New().String()

	if d.Toast != nil {
		d.Toast.Type = toastType
		d.Toast.Message = message
		d.Toast.ID = id
	} else {
		d.Toast = &ToastMessage{
			Type:    toastType,
			Message: message,
			ID:      id,
		}
	}
}

type Template struct {
	logger *slog.Logger
	config *config.Config
}

func NewTemplate(l *slog.Logger, c *config.Config) *Template {
	return &Template{
		logger: l,
		config: c,
	}
}

func (t *Template) Render(w http.ResponseWriter, r *http.Request, c templ.Component, status int) {
	var (
		method = r.Method
		url    = r.URL.RequestURI()
		trace  = string(debug.Stack())
		buff   bytes.Buffer
	)

	if err := c.Render(r.Context(), &buff); err != nil {
		t.logger.Error(err.Error(), "method", method, "url", url, "trace", trace)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)

	_, err := buff.WriteTo(w)
	if err != nil {
		t.logger.Error(err.Error(), "method", method, "url", url, "trace", trace)
	}
}

func (t *Template) GetData(r *http.Request) Data {
	var user AuthenticatedUser
	if u, ok := r.Context().Value(AuthenticatedUserKey).(AuthenticatedUser); ok {
		user = u
	}

	return Data{
		Title:       "Go Cost - Expense Tracker",
		CurrentYear: time.Now().Year(),
		CSRFToken:   nosurf.Token(r),
		User:        user,
		Version:     t.config.Version,
		Currency:    t.config.Currency,
	}
}
