package web

import (
	"net/http"

	"github.com/a-h/templ"
)

type Templater interface {
	Render(w http.ResponseWriter, r *http.Request, c templ.Component, status int)
	GetData(r *http.Request) Data
}
