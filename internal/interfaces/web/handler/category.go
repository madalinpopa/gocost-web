package handler

import (
	"errors"
	"net/http"

	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/form"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/madalinpopa/gocost-web/ui/templates/components"
)

type CategoryHandler struct {
	app      HandlerContext
	category usecase.CategoryUseCase
}

func NewCategoryHandler(app HandlerContext, category usecase.CategoryUseCase) CategoryHandler {
	return CategoryHandler{
		app:      app,
		category: category,
	}
}

func (h *CategoryHandler) GetCreateForm(w http.ResponseWriter, r *http.Request) {
	groupID, err := web.GetRequiredQueryParam(r, "group-id")
	if err != nil {
		h.app.Errors.Error(w, r, http.StatusBadRequest, err)
		return
	}

	categoryStart, err := web.GetRequiredQueryParam(r, "category-start")
	if err != nil {
		h.app.Errors.Error(w, r, http.StatusBadRequest, err)
		return
	}

	categoryForm := &form.CreateCategoryForm{
		GroupID:    groupID,
		StartMonth: categoryStart,
	}

	component := components.AddCategoryForm(categoryForm, h.app.Config.Currency, categoryStart)
	h.app.Template.Render(w, r, component, http.StatusOK)
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.app.Errors.Error(w, r, http.StatusBadRequest, err)
		return
	}

	var categoryForm form.CreateCategoryForm
	if err := h.app.Decoder.Decode(&categoryForm, r.PostForm); err != nil {
		h.app.Errors.Error(w, r, http.StatusBadRequest, err)
		return
	}

	categoryForm.Validate()
	if !categoryForm.IsValid() {
		component := components.AddCategoryForm(&categoryForm, h.app.Config.Currency, categoryForm.StartMonth)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
		return
	}

	isRecurrent := categoryForm.Type == "recurrent"

	userID := h.app.Session.GetUserID(r.Context())
	currency := h.app.Session.GetCurrency(r.Context())

	req := &usecase.CreateCategoryRequest{
		GroupID:     categoryForm.GroupID,
		UserID:      userID,
		Currency:    currency,
		Name:        categoryForm.Name,
		Description: categoryForm.Description,
		IsRecurrent: isRecurrent,
		StartMonth:  categoryForm.StartMonth,
		EndMonth:    categoryForm.EndMonth,
		Budget:      categoryForm.ParsedBudget(),
	}

	_, err := h.category.Create(r.Context(), req)
	if err != nil {
		errMessage, isUserFacing := translateCategoryError(err)
		categoryForm.AddNonFieldError(errMessage)
		component := components.AddCategoryForm(&categoryForm, h.app.Config.Currency, categoryForm.StartMonth)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)

		if !isUserFacing {
			h.app.Logger.Error("failed to create category", "error", err)
		}
		return
	}

	// Success - trigger dashboard refresh and toast.
	triggerDashboardRefresh(w, h.app.Notify, web.Success, "Category created successfully.", "add-category-modal")
	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.app.Errors.Error(w, r, http.StatusBadRequest, err)
		return
	}

	var categoryForm form.UpdateCategoryForm
	if err := h.app.Decoder.Decode(&categoryForm, r.PostForm); err != nil {
		h.app.Errors.Error(w, r, http.StatusBadRequest, err)
		return
	}

	categoryForm.Validate()
	if !categoryForm.IsValid() {
		component := components.EditCategoryForm(&categoryForm, h.app.Config.Currency)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
		return
	}

	isRecurrent := categoryForm.Type == "recurrent"

	userID := h.app.Session.GetUserID(r.Context())
	currency := h.app.Session.GetCurrency(r.Context())

	req := &usecase.UpdateCategoryRequest{
		ID:           categoryForm.ID,
		GroupID:      categoryForm.GroupID,
		UserID:       userID,
		Currency:     currency,
		Name:         categoryForm.Name,
		Description:  categoryForm.Description,
		IsRecurrent:  isRecurrent,
		StartMonth:   categoryForm.StartMonth,
		EndMonth:     categoryForm.EndMonth,
		CurrentMonth: categoryForm.CurrentMonth,
		Budget:       categoryForm.ParsedBudget(),
	}

	_, err := h.category.Update(r.Context(), req)
	if err != nil {
		errMessage, isUserFacing := translateCategoryError(err)
		categoryForm.AddNonFieldError(errMessage)
		component := components.EditCategoryForm(&categoryForm, h.app.Config.Currency)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)

		if !isUserFacing {
			h.app.Logger.Error("failed to update category", "error", err)
		}
		return
	}

	// Success - trigger dashboard refresh and toast.
	triggerDashboardRefresh(w, h.app.Notify, web.Success, "Category updated successfully.", "edit-category-modal")
	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("groupID")
	categoryID := r.PathValue("id")
	userID := h.app.Session.GetUserID(r.Context())

	if err := h.category.Delete(r.Context(), userID, groupID, categoryID); err != nil {
		h.app.Errors.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	triggerDashboardRefresh(w, h.app.Notify, web.Success, "Category deleted successfully.", "")
	w.WriteHeader(http.StatusNoContent)
}

func translateCategoryError(err error) (string, bool) {
	switch {
	case errors.Is(err, tracking.ErrEmptyName):
		return "Category name cannot be empty.", true
	case errors.Is(err, tracking.ErrNameTooLong):
		return "Category name is too long.", true
	case errors.Is(err, tracking.ErrDescriptionTooLong):
		return "Description is too long.", true
	case errors.Is(err, tracking.ErrInvalidMonth):
		return "Invalid month format.", true
	case errors.Is(err, tracking.ErrEndMonthBeforeStartMonth):
		return "End month must be after start month.", true
	case errors.Is(err, tracking.ErrEndMonthNotAllowed):
		return "End month is only allowed for recurrent categories.", true
	case errors.Is(err, tracking.ErrCategoryNameExists):
		return "Category name already exists in this group.", true
	case errors.Is(err, tracking.ErrCategoryGroupMismatch):
		return "Category does not belong to this group.", true
	case errors.Is(err, tracking.ErrGroupNotFound):
		return "Group not found.", true
	default:
		return "An unexpected error occurred. Please try again later.", false
	}
}
