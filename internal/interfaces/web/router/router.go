package router

import (
	"fmt"
	"net/http"

	"github.com/justinas/alice"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/handler"
)

// Router struct manages the routing of HTTP requests, applying middleware and handling static and dynamic routes.
type Router struct {
	mux             *http.ServeMux
	middleware      *web.Middleware
	baseRoutes      alice.Chain
	dynamicRoutes   alice.Chain
	protectedRoutes alice.Chain
}

// New creates and returns a new Router instance with the provided middleware.
func New(m *web.Middleware) *Router {
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

func (r *Router) RegisterRoutes(h handler.Handlers) {
	// Static file server
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	r.RegisterStaticFiles("/static/", http.StripPrefix("/static", fileServer))

	// Healthcheck
	r.RegisterUnprotectedHandler(http.MethodGet, "/health", handler.CheckHealthHandler)

	// Public pages
	r.RegisterPublicHandler(http.MethodGet, "/{$}", http.HandlerFunc(h.Public.IndexHandler.ShowIndexPage))
	r.RegisterPublicHandler(http.MethodGet, "/login", http.HandlerFunc(h.Public.LoginHandler.ShowLoginPage))
	r.RegisterPublicHandler(http.MethodGet, "/login/form", http.HandlerFunc(h.Public.LoginHandler.ShowLoginForm))
	r.RegisterPublicHandler(http.MethodPost, "/login", http.HandlerFunc(h.Public.LoginHandler.SubmitLoginForm))
	r.RegisterPublicHandler(http.MethodPost, "/logout", http.HandlerFunc(h.Public.LogoutHandler.SubmitLogout))
	r.RegisterPublicHandler(http.MethodGet, "/register", http.HandlerFunc(h.Public.RegisterHandler.ShowRegisterPage))
	r.RegisterPublicHandler(http.MethodGet, "/register/form", http.HandlerFunc(h.Public.RegisterHandler.ShowRegisterForm))
	r.RegisterPublicHandler(http.MethodPost, "/register", http.HandlerFunc(h.Public.RegisterHandler.SubmitRegisterForm))

	// Private pages
	r.RegisterPrivateHandler(http.MethodGet, "/home", http.HandlerFunc(h.Private.HomeHandler.ShowHomePage))
	r.RegisterPrivateHandler(http.MethodGet, "/home/groups", http.HandlerFunc(h.Private.HomeHandler.GetDashboardGroups))
	r.RegisterPrivateHandler(http.MethodGet, "/incomes", http.HandlerFunc(h.Private.IncomeHandler.ListIncomes))
	r.RegisterPrivateHandler(http.MethodGet, "/incomes/form", http.HandlerFunc(h.Private.IncomeHandler.GetCreateForm))
	r.RegisterPrivateHandler(http.MethodPost, "/incomes", http.HandlerFunc(h.Private.IncomeHandler.CreateIncome))
	r.RegisterPrivateHandler(http.MethodDelete, "/incomes/{id}", http.HandlerFunc(h.Private.IncomeHandler.DeleteIncome))
	r.RegisterPrivateHandler(http.MethodGet, "/groups/form", http.HandlerFunc(h.Private.GroupHandler.GetCreateForm))
	r.RegisterPrivateHandler(http.MethodPost, "/groups", http.HandlerFunc(h.Private.GroupHandler.CreateGroup))
	r.RegisterPrivateHandler(http.MethodPost, "/groups/edit", http.HandlerFunc(h.Private.GroupHandler.UpdateGroup))
	r.RegisterPrivateHandler(http.MethodDelete, "/groups/{id}", http.HandlerFunc(h.Private.GroupHandler.DeleteGroup))
	r.RegisterPrivateHandler(http.MethodGet, "/categories/form", http.HandlerFunc(h.Private.CategoryHandler.GetCreateForm))
	r.RegisterPrivateHandler(http.MethodPost, "/categories", http.HandlerFunc(h.Private.CategoryHandler.CreateCategory))
	r.RegisterPrivateHandler(http.MethodPost, "/categories/edit", http.HandlerFunc(h.Private.CategoryHandler.UpdateCategory))
	r.RegisterPrivateHandler(http.MethodDelete, "/groups/{groupID}/categories/{id}", http.HandlerFunc(h.Private.CategoryHandler.DeleteCategory))
	r.RegisterPrivateHandler(http.MethodGet, "/expenses/form", http.HandlerFunc(h.Private.ExpenseHandler.GetCreateForm))
	r.RegisterPrivateHandler(http.MethodPost, "/expenses", http.HandlerFunc(h.Private.ExpenseHandler.CreateExpense))
	r.RegisterPrivateHandler(http.MethodPost, "/expenses/edit", http.HandlerFunc(h.Private.ExpenseHandler.EditExpense))
	r.RegisterPrivateHandler(http.MethodDelete, "/expenses/{id}", http.HandlerFunc(h.Private.ExpenseHandler.DeleteExpense))
}
