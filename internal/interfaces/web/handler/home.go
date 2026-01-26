package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/views"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/madalinpopa/gocost-web/ui/templates/components"
	"github.com/madalinpopa/gocost-web/ui/templates/pages/private"
)

type HomeHandler struct {
	app        HandlerContext
	incomeUC   usecase.IncomeUseCase
	expenseUC  usecase.ExpenseUseCase
	groupUC    usecase.GroupUseCase
	categoryUC usecase.CategoryUseCase
}

func NewHomeHandler(
	app HandlerContext,
	incomeUC usecase.IncomeUseCase,
	expenseUC usecase.ExpenseUseCase,
	groupUC usecase.GroupUseCase,
	categoryUC usecase.CategoryUseCase,
) HomeHandler {
	return HomeHandler{
		app:        app,
		incomeUC:   incomeUC,
		expenseUC:  expenseUC,
		groupUC:    groupUC,
		categoryUC: categoryUC,
	}
}

func (hh HomeHandler) ShowHomePage(w http.ResponseWriter, r *http.Request) {
	data := hh.app.Template.GetData(r)

	// Determine the month to display
	currentDate, prevDate, nextDate := web.GetMonthParam(r)
	monthStr := currentDate.Format("2006-01")

	dashboardData, err := hh.fetchDashboardData(r.Context(), data.User.ID, currentDate)
	if err != nil {
		hh.app.Errors.LogServerError(r, err)
		return
	}

	dashboardData.CurrentMonth = currentDate.Format("January 2006")
	dashboardData.CurrentMonthParam = monthStr
	dashboardData.PrevMonth = prevDate.Format("2006-01")
	dashboardData.NextMonth = nextDate.Format("2006-01")

	page := private.HomePage(data, dashboardData)
	hh.app.Template.Render(w, r, page, http.StatusOK)
}

func (hh HomeHandler) GetDashboardGroups(w http.ResponseWriter, r *http.Request) {
	currentDate, prevDate, nextDate := web.GetMonthParam(r)
	userID := hh.app.Session.GetUserID(r.Context())
	monthStr := currentDate.Format("2006-01")

	dashboardData, err := hh.fetchDashboardData(r.Context(), userID, currentDate)
	if err != nil {
		hh.app.Errors.LogServerError(r, err)
		return
	}

	dashboardData.CurrentMonth = currentDate.Format("January 2006")
	dashboardData.CurrentMonthParam = monthStr
	dashboardData.PrevMonth = prevDate.Format("2006-01")
	dashboardData.NextMonth = nextDate.Format("2006-01")

	// 1. Render the Groups List (Main Target)
	err = components.DashboardGroups(dashboardData.Groups, monthStr).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Render OOB Updates
	// Month Navigation
	err = components.MonthNavigation(dashboardData, true).Render(r.Context(), w)
	if err != nil {
		hh.app.Errors.LogServerError(r, err)
	}

	// Balance Display
	err = components.BalanceDisplay(dashboardData, true).Render(r.Context(), w)
	if err != nil {
		hh.app.Errors.LogServerError(r, err)
	}

	// Dashboard Actions
	err = components.DashboardActions(monthStr, true).Render(r.Context(), w)
	if err != nil {
		hh.app.Errors.LogServerError(r, err)
	}
}

func (hh HomeHandler) fetchDashboardData(ctx context.Context, userID string, date time.Time) (views.DashboardView, error) {
	monthStr := date.Format("2006-01")

	totalIncome, err := hh.incomeUC.Total(ctx, userID, monthStr)
	if err != nil {
		return views.DashboardView{}, err
	}

	totalExpenses, err := hh.expenseUC.Total(ctx, userID, monthStr)
	if err != nil {
		return views.DashboardView{}, err
	}

	// Fetch groups
	groupsDTO, err := hh.groupUC.List(ctx, userID)
	if err != nil {
		return views.DashboardView{}, err
	}

	// Fetch expenses for the user and month
	expensesDTO, err := hh.expenseUC.ListByMonth(ctx, userID, monthStr)
	if err != nil {
		return views.DashboardView{}, err
	}

	presenter := views.NewDashboardPresenter(hh.app.Config.Currency)
	return presenter.Present(totalIncome, totalExpenses, groupsDTO, expensesDTO, date), nil
}
