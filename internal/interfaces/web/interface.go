package web

import (
	"net/http"

	"github.com/a-h/templ"
)

type NotifyHandler interface {
	Trigger(w http.ResponseWriter, events map[string]any)
	Toast(w http.ResponseWriter, t ToastType, message string)
}

type HtmxHandler interface {
	Redirect(w http.ResponseWriter, url string)
	Location(w http.ResponseWriter, r *http.Request, url string, target string, swap string)
}

type ErrorHandler interface {
	ServerError(w http.ResponseWriter, r *http.Request, err error)
	Error(w http.ResponseWriter, r *http.Request, status int, err error)
	LogServerError(r *http.Request, err error)
}

type Templater interface {
	Render(w http.ResponseWriter, r *http.Request, c templ.Component, status int)
	GetData(r *http.Request) Data
}
