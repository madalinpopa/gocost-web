package response

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
	"github.com/madalinpopa/gocost-web/internal/infrastructure/session"
)

type ToastType string

const (
	Success  ToastType = "success"
	ErrorMsg ToastType = "error"
	Warning  ToastType = "warning"
	Info     ToastType = "info"
)

type ToastMessage struct {
	Type    ToastType
	Message string
	Style   string
	ID      string
}

type Data struct {
	Title       string
	CurrentYear int
	CSRFToken   string
	Toast       *ToastMessage
	User        session.AuthenticatedUser
	Version     string
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
	data   Data
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (t *Template) GetData(r *http.Request) Data {
	var user session.AuthenticatedUser
	if u, ok := r.Context().Value(session.AuthenticatedUserKey).(session.AuthenticatedUser); ok {
		user = u
	}

	return Data{
		Title:       "Go Cost - Expense Tracker",
		CurrentYear: time.Now().Year(),
		CSRFToken:   nosurf.Token(r),
		User:        user,
		Version:     t.config.Version,
	}
}
