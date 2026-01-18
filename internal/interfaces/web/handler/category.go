package handler

import (
	"errors"
	"net/http"

	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
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

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	var categoryForm form.CreateCategoryForm
	if err := h.app.Decoder.Decode(&categoryForm, r.PostForm); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	categoryForm.Validate()
	if !categoryForm.IsValid() {
		component := components.AddCategoryForm(&categoryForm, h.app.Config.Currency, categoryForm.StartMonth)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
		return
	}

	isRecurrent := categoryForm.Type == "recurrent"

	req := &usecase.CreateCategoryRequest{
		Name:        categoryForm.Name,
		Description: categoryForm.Description,
		IsRecurrent: isRecurrent,
		StartMonth:  categoryForm.StartMonth,
		EndMonth:    categoryForm.EndMonth,
		Budget:      categoryForm.Budget,
	}

	userID := h.app.Session.GetUserID(r.Context())
	_, err := h.category.Create(r.Context(), userID, categoryForm.GroupID, req)
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

	// Success - trigger UI update
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	var categoryForm form.UpdateCategoryForm
	if err := h.app.Decoder.Decode(&categoryForm, r.PostForm); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	categoryForm.Validate()
	if !categoryForm.IsValid() {
		component := components.EditCategoryForm(&categoryForm, h.app.Config.Currency)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
		return
	}

	isRecurrent := categoryForm.Type == "recurrent"

	req := &usecase.UpdateCategoryRequest{
		ID:          categoryForm.ID,
		Name:        categoryForm.Name,
		Description: categoryForm.Description,
		IsRecurrent: isRecurrent,
		StartMonth:  categoryForm.StartMonth,
		EndMonth:    categoryForm.EndMonth,
		Budget:      categoryForm.Budget,
	}

	userID := h.app.Session.GetUserID(r.Context())
	_, err := h.category.Update(r.Context(), userID, categoryForm.GroupID, req)
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

	// Success - trigger UI update
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("groupID")
	categoryID := r.PathValue("id")
	userID := h.app.Session.GetUserID(r.Context())

	if err := h.category.Delete(r.Context(), userID, groupID, categoryID); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
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
