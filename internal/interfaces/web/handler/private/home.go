package private

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/madalinpopa/gocost-web/internal/app"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/params"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/views"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/madalinpopa/gocost-web/ui/templates/components"
	"github.com/madalinpopa/gocost-web/ui/templates/pages/private"
)

type HomeHandler struct {
	app        app.HandlerContext
	incomeUC   usecase.IncomeUseCase
	expenseUC  usecase.ExpenseUseCase
	groupUC    usecase.GroupUseCase
	categoryUC usecase.CategoryUseCase
}

func NewHomeHandler(
	app app.HandlerContext,
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
	currentDate, prevDate, nextDate := params.GetMonthParam(r)
	monthStr := currentDate.Format("2006-01")

	dashboardData, err := hh.fetchDashboardData(r.Context(), data.User.ID, currentDate)
	if err != nil {
		hh.app.Response.Handle.LogServerError(r, err)
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
	currentDate, prevDate, nextDate := params.GetMonthParam(r)
	userID := hh.app.Session.GetUserID(r.Context())
	monthStr := currentDate.Format("2006-01")

	dashboardData, err := hh.fetchDashboardData(r.Context(), userID, currentDate)
	if err != nil {
		hh.app.Response.Handle.LogServerError(r, err)
		return
	}

	dashboardData.CurrentMonth = currentDate.Format("January 2006")
	dashboardData.CurrentMonthParam = monthStr
	dashboardData.PrevMonth = prevDate.Format("2006-01")
	dashboardData.NextMonth = nextDate.Format("2006-01")

	// 1. Render the Groups List (Main Target)
	err = components.GroupsList(dashboardData.Groups, monthStr).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Render OOB Updates
	// Month Navigation
	err = components.MonthNavigation(dashboardData, true).Render(r.Context(), w)
	if err != nil {
		hh.app.Response.Handle.LogServerError(r, err)
	}

	// Balance Display
	err = components.BalanceDisplay(dashboardData, true).Render(r.Context(), w)
	if err != nil {
		hh.app.Response.Handle.LogServerError(r, err)
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

	balance := totalIncome - totalExpenses

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

	// Create a map of expenses by category for easier lookup
	expensesByCategory := make(map[string][]views.ExpenseView)
	categorySpent := make(map[string]float64)

	for _, exp := range expensesDTO {
		status := views.StatusUnpaid
		paidAt := ""
		if exp.IsPaid {
			status = views.StatusPaid
			if exp.PaidAt != nil {
				paidAt = exp.PaidAt.Format("2006-01-02")
			}
		}
		expensesByCategory[exp.CategoryID] = append(expensesByCategory[exp.CategoryID], views.ExpenseView{
			ID:          exp.ID,
			Amount:      exp.Amount,
			Currency:    hh.app.Config.Currency,
			Description: exp.Description,
			Status:      status,
			SpentAt:     exp.SpentAt.Format("2006-01-02"),
			PaidAt:      paidAt,
		})
		categorySpent[exp.CategoryID] += exp.Amount
	}

	// Round spent amounts to avoid floating point precision issues
	for id, spent := range categorySpent {
		categorySpent[id] = math.Round(spent*100) / 100
	}

	// Map to views
	var groupViews []views.GroupView
	for _, grp := range groupsDTO {
		var categoryViews []views.CategoryView

		for _, cat := range grp.Categories {
			catType := views.TypeMonthly
			if cat.IsRecurrent {
				catType = views.TypeRecurrent
			}

			// Check if category is active for this month
			// StartMonth is required. EndMonth is optional.
			// Format is YYYY-MM
			start, _ := time.Parse("2006-01", cat.StartMonth)
			var end time.Time
			if cat.EndMonth != "" {
				end, _ = time.Parse("2006-01", cat.EndMonth)
			}

			// Logic to show category:
			// If recurrent: show if date >= start (and <= end if end is set)
			// If monthly: show if date == start

			showCategory := false
			currentMonthStart, _ := time.Parse("2006-01", monthStr)

			if cat.IsRecurrent {
				if !currentMonthStart.Before(start) {
					if cat.EndMonth == "" || !currentMonthStart.After(end) {
						showCategory = true
					}
				}
			} else {
				if cat.StartMonth == monthStr {
					showCategory = true
				}
			}

			if showCategory {
				categoryViews = append(categoryViews, views.CategoryView{
					ID:          cat.ID,
					Name:        cat.Name,
					Type:        catType,
					Description: cat.Description,
					StartMonth:  cat.StartMonth,
					EndMonth:    cat.EndMonth,
					Budget:      cat.Budget,
					Spent:       categorySpent[cat.ID],
					Currency:    hh.app.Config.Currency,
					Expenses:    expensesByCategory[cat.ID],
				})
			}
		}

		groupViews = append(groupViews, views.GroupView{
			ID:          grp.ID,
			Name:        grp.Name,
			Description: grp.Description,
			Order:       grp.Order,
			Categories:  categoryViews,
		})
	}

	return views.DashboardView{
		TotalIncome:   totalIncome,
		TotalExpenses: totalExpenses,
		Balance:       balance,
		Currency:      hh.app.Config.Currency,
		Groups:        groupViews,
	}, nil
}
