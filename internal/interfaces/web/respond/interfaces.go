package respond

import "net/http"

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
