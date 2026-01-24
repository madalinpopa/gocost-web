package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/form"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/madalinpopa/gocost-web/ui/templates/components"
)

type ExpenseHandler struct {
	app     HandlerContext
	expense usecase.ExpenseUseCase
}

func NewExpenseHandler(app HandlerContext, expense usecase.ExpenseUseCase) ExpenseHandler {
	return ExpenseHandler{
		app:     app,
		expense: expense,
	}
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	var expenseForm form.CreateExpenseForm
	if err := h.app.Decoder.Decode(&expenseForm, r.PostForm); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	expenseForm.Validate()
	if !expenseForm.IsValid() {
		component := components.AddExpenseForm(&expenseForm, h.app.Config.Currency)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
		return
	}

	spentAt, err := time.Parse("2006-01", expenseForm.Month)
	if err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	isPaid := expenseForm.PaymentStatus == "paid"
	var paidAt *time.Time
	if isPaid {
		now := time.Now()
		paidAt = &now
	}

	req := &usecase.CreateExpenseRequest{
		CategoryID:  expenseForm.CategoryID,
		Amount:      expenseForm.Amount,
		Description: expenseForm.Description,
		SpentAt:     spentAt,
		IsPaid:      isPaid,
		PaidAt:      paidAt,
	}

	userID := h.app.Session.GetUserID(r.Context())

	_, err = h.expense.Create(r.Context(), userID, req)
	if err != nil {
		errMessage, isUserFacing := translateExpenseError(err)
		expenseForm.AddNonFieldError(errMessage)
		component := components.AddExpenseForm(&expenseForm, h.app.Config.Currency)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)

		if !isUserFacing {
			h.app.Logger.Error("failed to create expense", "error", err)
		}
		return
	}

	// Success
	triggerDashboardRefresh(w, h.app.Response.Notify, web.Success, "Expense created.", "add-expense-modal")
	w.WriteHeader(http.StatusNoContent)
}

func (h *ExpenseHandler) EditExpense(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	var expenseForm form.UpdateExpenseForm
	if err := h.app.Decoder.Decode(&expenseForm, r.PostForm); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	expenseForm.Validate()
	if !expenseForm.IsValid() {
		component := components.EditExpenseForm(&expenseForm, h.app.Config.Currency)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
		return
	}

	userID := h.app.Session.GetUserID(r.Context())

	existing, err := h.expense.Get(r.Context(), userID, expenseForm.ID)
	if err != nil {
		errMessage, isUserFacing := translateExpenseError(err)
		if isUserFacing {
			expenseForm.AddNonFieldError(errMessage)
			component := components.EditExpenseForm(&expenseForm, h.app.Config.Currency)
			h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
			return
		}
		h.app.Logger.Error("failed to fetch expense for edit", "error", err)
		h.app.Response.Handle.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	isPaid := expenseForm.PaymentStatus == "paid"
	var paidAt *time.Time
	if isPaid {
		if existing.IsPaid && existing.PaidAt != nil {
			paidAt = existing.PaidAt
		} else {
			now := time.Now()
			paidAt = &now
		}
	}

	req := &usecase.UpdateExpenseRequest{
		ID:          expenseForm.ID,
		CategoryID:  expenseForm.CategoryID,
		Amount:      expenseForm.Amount,
		Description: expenseForm.Description,
		SpentAt:     existing.SpentAt,
		IsPaid:      isPaid,
		PaidAt:      paidAt,
	}

	_, err = h.expense.Update(r.Context(), userID, req)
	if err != nil {
		errMessage, isUserFacing := translateExpenseError(err)
		expenseForm.AddNonFieldError(errMessage)
		component := components.EditExpenseForm(&expenseForm, h.app.Config.Currency)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)

		if !isUserFacing {
			h.app.Logger.Error("failed to update expense", "error", err)
		}
		return
	}

	// Success
	triggerDashboardRefresh(w, h.app.Response.Notify, web.Success, "Expense updated.", "edit-expense-modal")
	w.WriteHeader(http.StatusNoContent)
}

func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	userID := h.app.Session.GetUserID(r.Context())
	expenseID := r.PathValue("id")

	if err := h.expense.Delete(r.Context(), userID, expenseID); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	triggerDashboardRefresh(w, h.app.Response.Notify, web.Success, "Expense deleted.", "")
	w.WriteHeader(http.StatusNoContent)
}

func translateExpenseError(err error) (string, bool) {
	switch {
	case errors.Is(err, money.ErrNegativeAmount):
		return "Amount cannot be negative.", true
	case errors.Is(err, expense.ErrExpenseDescriptionTooLong):
		return "Description is too long.", true
	case errors.Is(err, tracking.ErrCategoryNotFound):
		return "Category not found.", true
	case errors.Is(err, expense.ErrExpenseNotFound):
		return "Expense not found.", true
	default:
		return "An unexpected error occurred. Please try again later.", false
	}
}
