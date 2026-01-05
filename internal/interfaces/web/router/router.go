package router

import (
	"fmt"
	"net/http"

	"github.com/justinas/alice"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/middleware"
)

// Router struct manages the routing of HTTP requests, applying middleware and handling static and dynamic routes.
type Router struct {
	mux             *http.ServeMux
	middleware      *middleware.Middleware
	baseRoutes      alice.Chain
	dynamicRoutes   alice.Chain
	protectedRoutes alice.Chain
}

// New creates and returns a new Router instance with the provided middleware.
func New(m *middleware.Middleware) *Router {
	baseRoutes := alice.New(m.Recover, m.Logging, m.Headers)
	dynamicRoutes := alice.New(m.LoadSession, m.CsrfToken, m.CheckAllowedHosts, m.Authenticate)
	protectedRoutes := dynamicRoutes.Append(m.LoginRequired)
	return &Router{
		mux:             http.NewServeMux(),
		middleware:      m,
		baseRoutes:      baseRoutes,
		dynamicRoutes:   dynamicRoutes,
		protectedRoutes: protectedRoutes,
	}
}

// RegisterPublicHandler registers a public HTTP handler with CSRF protection for the specified method and URL path.
func (r *Router) RegisterPublicHandler(method, url string, handler http.HandlerFunc) {
	r.mux.Handle(fmt.Sprintf("%s %s", method, url), r.dynamicRoutes.ThenFunc(handler))
}

// RegisterPrivateHandler registers a private HTTP handler for the given method and URL without applying m.
func (r *Router) RegisterPrivateHandler(method, url string, handler http.HandlerFunc) {
	r.mux.Handle(fmt.Sprintf("%s %s", method, url), r.protectedRoutes.ThenFunc(handler))
}

// RegisterUnprotectedHandler registers an unprotected HTTP handler (without CSRF protection) for the specified method and URL path.
func (r *Router) RegisterUnprotectedHandler(method, url string, handler http.HandlerFunc) {
	r.mux.Handle(fmt.Sprintf("%s %s", method, url), r.baseRoutes.ThenFunc(handler))
}

// RegisterStaticFiles maps a static file handler to the specified URL path in the router.
func (r *Router) RegisterStaticFiles(url string, handler http.Handler) {
	r.mux.Handle(url, handler)
}

// Handlers assembles the app m and returns the configured HTTP handler for routing requests.
func (r *Router) Handlers() http.Handler {
	return r.baseRoutes.Then(r.mux)
}
