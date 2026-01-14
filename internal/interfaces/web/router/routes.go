package router

import (
	"net/http"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web/handler"
)

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
	r.RegisterPrivateHandler(http.MethodPost, "/incomes", http.HandlerFunc(h.Private.IncomeHandler.CreateIncome))
	r.RegisterPrivateHandler(http.MethodDelete, "/incomes/{id}", http.HandlerFunc(h.Private.IncomeHandler.DeleteIncome))
	r.RegisterPrivateHandler(http.MethodPost, "/groups", http.HandlerFunc(h.Private.GroupHandler.CreateGroup))
	r.RegisterPrivateHandler(http.MethodPost, "/groups/edit", http.HandlerFunc(h.Private.GroupHandler.UpdateGroup))
	r.RegisterPrivateHandler(http.MethodDelete, "/groups/{id}", http.HandlerFunc(h.Private.GroupHandler.DeleteGroup))
	r.RegisterPrivateHandler(http.MethodPost, "/categories", http.HandlerFunc(h.Private.CategoryHandler.CreateCategory))
	r.RegisterPrivateHandler(http.MethodPost, "/categories/edit", http.HandlerFunc(h.Private.CategoryHandler.UpdateCategory))
	r.RegisterPrivateHandler(http.MethodDelete, "/groups/{groupID}/categories/{id}", http.HandlerFunc(h.Private.CategoryHandler.DeleteCategory))
	r.RegisterPrivateHandler(http.MethodPost, "/expenses", http.HandlerFunc(h.Private.ExpenseHandler.CreateExpense))
	r.RegisterPrivateHandler(http.MethodPost, "/expenses/edit", http.HandlerFunc(h.Private.ExpenseHandler.EditExpense))
	r.RegisterPrivateHandler(http.MethodDelete, "/expenses/{id}", http.HandlerFunc(h.Private.ExpenseHandler.DeleteExpense))
}
