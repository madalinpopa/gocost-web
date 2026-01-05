package private

import (
	"errors"
	"net/http"

	"github.com/madalinpopa/gocost-web/internal/app"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/form"
	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/madalinpopa/gocost-web/ui/templates/components"
)

type GroupHandler struct {
	app   app.ApplicationContext
	group usecase.GroupUseCase
}

func NewGroupHandler(app app.ApplicationContext, group usecase.GroupUseCase) GroupHandler {
	return GroupHandler{
		app:   app,
		group: group,
	}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	var groupForm form.CreateGroupForm
	if err := h.app.Decoder.Decode(&groupForm, r.PostForm); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	groupForm.Validate()
	if !groupForm.IsValid() {
		component := components.AddGroupForm(&groupForm)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
		return
	}

	userID := h.app.Session.GetUserID(r.Context())

	req := &usecase.CreateGroupRequest{
		Name:        groupForm.Name,
		Description: groupForm.Description,
		Order:       groupForm.Order,
	}

	_, err := h.group.Create(r.Context(), userID, req)
	if err != nil {
		errMessage, isUserFacing := translateGroupError(err)
		groupForm.AddNonFieldError(errMessage)
		component := components.AddGroupForm(&groupForm)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)

		if !isUserFacing {
			h.app.Logger.Error("failed to create group", "error", err)
		}
		return
	}

	// Success - trigger page refresh
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	var groupForm form.UpdateGroupForm
	if err := h.app.Decoder.Decode(&groupForm, r.PostForm); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusBadRequest, err)
		return
	}

	groupForm.Validate()
	if !groupForm.IsValid() {
		component := components.EditGroupForm(&groupForm)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)
		return
	}

	userID := h.app.Session.GetUserID(r.Context())

	req := &usecase.UpdateGroupRequest{
		ID:          groupForm.ID,
		Name:        groupForm.Name,
		Description: groupForm.Description,
		Order:       groupForm.Order,
	}

	_, err := h.group.Update(r.Context(), userID, req)
	if err != nil {
		errMessage, isUserFacing := translateGroupError(err)
		groupForm.AddNonFieldError(errMessage)
		component := components.EditGroupForm(&groupForm)
		h.app.Template.Render(w, r, component, http.StatusUnprocessableEntity)

		if !isUserFacing {
			h.app.Logger.Error("failed to update group", "error", err)
		}
		return
	}

	// Success - trigger page refresh
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	userID := h.app.Session.GetUserID(r.Context())
	groupID := r.PathValue("id")

	if err := h.group.Delete(r.Context(), userID, groupID); err != nil {
		h.app.Response.Handle.Error(w, r, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func translateGroupError(err error) (string, bool) {
	switch {
	case errors.Is(err, tracking.ErrEmptyName):
		return "Group name cannot be empty.", true
	case errors.Is(err, tracking.ErrNameTooLong):
		return "Group name is too long.", true
	case errors.Is(err, tracking.ErrDescriptionTooLong):
		return "Description is too long.", true
	case errors.Is(err, tracking.ErrInvalidOrder):
		return "Order must be non-negative.", true
	default:
		return "An unexpected error occurred. Please try again later.", false
	}
}
