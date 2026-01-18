package handler

import (
	"net/http"
	"time"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web/form"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/views"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/madalinpopa/gocost-web/ui/templates/components"
)

type IncomeHandler struct {
	app     HandlerContext
	income  usecase.IncomeUseCase
	expense usecase.ExpenseUseCase
}

func NewIncomeHandler(app HandlerContext, income usecase.IncomeUseCase, expense usecase.ExpenseUseCase) IncomeHandler {
	return IncomeHandler{
		app:     app,
		income:  income,
		expense: expense,
	}
}

func (h *IncomeHandler) CreateIncome(w http.ResponseWriter, r *http.Request) {
	var incomeForm form.CreateIncomeForm
	err := form.ParseAndValidateForm(r, h.app.Decoder, &incomeForm)
	if err != nil {
		h.app.Response.Handle.LogServerError(r, err)
		return
	}

	if !incomeForm.IsValid() {
		component := components.AddIncomeForm(&incomeForm, h.app.Config.Currency)
		if err := component.Render(r.Context(), w); err != nil {
			h.app.Response.Handle.LogServerError(r, err)
		}
		return
	}

	date, _ := time.Parse("2006-01-02", incomeForm.Date)

	req := &usecase.CreateIncomeRequest{
		Amount:     incomeForm.Amount,
		Source:     incomeForm.Description,
		ReceivedAt: date,
	}

	userID := h.app.Session.GetUserID(r.Context())
	if userID == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	_, err = h.income.Create(r.Context(), userID, req)
	if err != nil {
		h.app.Response.Handle.LogServerError(r, err)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *IncomeHandler) ListIncomes(w http.ResponseWriter, r *http.Request) {
	userID := h.app.Session.GetUserID(r.Context())
	month := r.URL.Query().Get("month")

	if month == "" {
		month = time.Now().Format("2006-01")
	}

	incomes, err := h.income.ListByMonth(r.Context(), userID, month)
	if err != nil {
		h.app.Response.Handle.LogServerError(r, err)
		return
	}

	err = components.IncomeList(incomes, h.app.Config.Currency).Render(r.Context(), w)
	if err != nil {
		h.app.Response.Handle.LogServerError(r, err)
	}
}

func (h *IncomeHandler) DeleteIncome(w http.ResponseWriter, r *http.Request) {
	userID := h.app.Session.GetUserID(r.Context())
	id := r.PathValue("id")

	// Get income to know the month for balance updated
	inc, err := h.income.Get(r.Context(), userID, id)
	if err != nil {
		h.app.Response.Handle.LogServerError(r, err)
		return
	}
	monthStr := inc.ReceivedAt.Format("2006-01")

	err = h.income.Delete(r.Context(), userID, id)
	if err != nil {
		h.app.Response.Handle.LogServerError(r, err)
		return
	}

	// Calculate new balance for the month
	totalIncome, err := h.income.Total(r.Context(), userID, monthStr)
	if err != nil {
		h.app.Response.Handle.LogServerError(r, err)
		return
	}

	totalExpenses, err := h.expense.Total(r.Context(), userID, monthStr)
	if err != nil {
		h.app.Response.Handle.LogServerError(r, err)
		return
	}

	balance := totalIncome - totalExpenses

	dashboardView := views.DashboardView{
		Balance:           balance,
		Currency:          h.app.Config.Currency,
		CurrentMonthParam: monthStr, // Needed for the link in BalanceDisplay
	}

	// Render BalanceDisplay OOB
	err = components.BalanceDisplay(dashboardView, true).Render(r.Context(), w)
	if err != nil {
		h.app.Response.Handle.LogServerError(r, err)
	}
}
