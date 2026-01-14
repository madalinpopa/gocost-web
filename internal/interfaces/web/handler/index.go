package handler

import (
	"net/http"

	"github.com/madalinpopa/gocost-web/ui/templates/pages/public"
)

type IndexHandler struct {
	app HandlerContext
}

func NewIndexHandler(app HandlerContext) IndexHandler {
	return IndexHandler{
		app: app,
	}
}

func (ih IndexHandler) ShowIndexPage(w http.ResponseWriter, r *http.Request) {
	data := ih.app.Template.GetData(r)
	page := public.Index(data)
	ih.app.Template.Render(w, r, page, http.StatusOK)
}
