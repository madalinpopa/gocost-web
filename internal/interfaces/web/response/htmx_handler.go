package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type htmx struct {
	logger     *slog.Logger
	errHandler ErrorHandler
}

func newHtmx(l *slog.Logger, h ErrorHandler) htmx {
	return htmx{
		logger:     l,
		errHandler: h,
	}
}

func (h htmx) Redirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusFound)
}

func (h htmx) Location(w http.ResponseWriter, r *http.Request, url string, target string, swap string) {
	locationObj := map[string]string{
		"path": url,
	}

	if target != "" {
		locationObj["target"] = target
	}

	if swap != "" {
		locationObj["swap"] = swap
	}

	locationJSON, err := json.Marshal(locationObj)
	if err != nil {
		h.errHandler.ServerError(w, r, err)
		return
	}
	w.Header().Set("HX-Location", string(locationJSON))
	w.WriteHeader(http.StatusOK)
}
