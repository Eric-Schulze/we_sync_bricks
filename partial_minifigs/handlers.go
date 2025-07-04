package partial_minifigs

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/go-chi/chi/v5"
)

type PartialMinifigHandler struct {
	service   *PartialMinifigService
	templates *template.Template
	jwtSecret []byte
}

// NewPartialMinifigHandler creates a new partial minifig handler
func NewPartialMinifigHandler(service *PartialMinifigService, templates *template.Template, jwtSecret []byte) *PartialMinifigHandler {
	return &PartialMinifigHandler{
		service:   service,
		templates: templates,
		jwtSecret: jwtSecret,
	}
}

func (handler *PartialMinifigHandler) HandlePartialMinifigListsPage() http.HandlerFunc {
	// Use this closure to prepare any objects needed for the handler
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Get current user from context (set by auth middleware)
			user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			logger.Info("Handling request to Partial Minifig List", "user_id", user.ID, "email", user.Email)

			lists, err := handler.service.GetAllPartialMinifigLists(user)
			if err != nil {
				logger.Error("Error getting lists", "user_id", user.ID, "error", err)
				http.Error(w, "Error loading lists", http.StatusInternalServerError)
				return
			}

			data := map[string]interface{}{
				"Title":       "Partial Minifigs Lists",
				"CurrentPage": "partial_minifigs_lists",
				"Lists":       lists,
				"User":        user,
			}

			// Check if this is an HTMX request
			if r.Header.Get("HX-Request") == "true" {
				if err := handler.templates.ExecuteTemplate(w, "partial-minifigs-lists-content", data); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			// If not an HTMX request, render the full page
			if err := handler.templates.ExecuteTemplate(w, "partial-minifigs-lists-page", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		},
	)
}

func (handler *PartialMinifigHandler) HandlePartialMinifigListDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		vars := chi.URLParam(r, "id")
		id, err := strconv.Atoi(vars)
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		list, err := handler.service.GetPartialMinifigListByID(id, user)
		if err != nil {
			http.Error(w, "List not found", http.StatusNotFound)
			return
		}

		data := map[string]interface{}{
			"Title":       list.Name,
			"CurrentPage": "lists",
			"List":        list,
			"User":        user,
		}

		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			if err := handler.templates.ExecuteTemplate(w, "partial-minifig-list-content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// If not an HTMX request, render the full page
		if err := handler.templates.ExecuteTemplate(w, "partial-minifig-list-page", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *PartialMinifigHandler) HandleNewPartialMinifigListModal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{}

		if err := handler.templates.ExecuteTemplate(w, "new-partial-minfig-list-modal", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *PartialMinifigHandler) HandleCreatePartialMinifigList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		name := r.Form.Get("name")
		description := r.Form.Get("description")

		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		newList, err := handler.service.CreatePartialMinifigList(name, description, user)
		if err != nil {
			http.Error(w, "Error creating list", http.StatusInternalServerError)
			return
		}

		// For HTMX requests, redirect to the new list detail page
		if r.Header.Get("HX-Request") == "true" {
			// Tell HTMX to redirect to the new list
			w.Header().Set("HX-Redirect", "/partial-minifigs-lists/"+strconv.Itoa(newList.ID))
			w.WriteHeader(http.StatusOK)
			return
		}

		// For non-HTMX requests, do a regular redirect
		http.Redirect(w, r, "/partial-minifigs-lists/"+strconv.Itoa(newList.ID), http.StatusSeeOther)
	}
}

func (handler *PartialMinifigHandler) HandleEditPartialMinifigListModal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		vars := chi.URLParam(r, "id")
		id, err := strconv.Atoi(vars)
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		list, err := handler.service.GetPartialMinifigListByID(id, user)
		if err != nil {
			http.Error(w, "List not found", http.StatusNotFound)
			return
		}

		data := map[string]interface{}{
			"List": list,
			"User": user,
		}

		if err := handler.templates.ExecuteTemplate(w, "edit-list-modal", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *PartialMinifigHandler) HandleUpdatePartialMinifigList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		vars := chi.URLParam(r, "id")
		id, err := strconv.Atoi(vars)
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		name := r.Form.Get("name")
		description := r.Form.Get("description")

		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		list, err := handler.service.UpdatePartialMinifigList(id, name, description, user)
		if err != nil {
			http.Error(w, "List not found", http.StatusNotFound)
			return
		}

		// Return updated list detail
		data := map[string]interface{}{
			"Title":       list.Name,
			"CurrentPage": "lists",
			"List":        list,
			"User":        user,
		}

		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			if err := handler.templates.ExecuteTemplate(w, "partial-minifig-list-content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// If not an HTMX request, render the full page
		if err := handler.templates.ExecuteTemplate(w, "partial-minifig-list-page", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *PartialMinifigHandler) HandleNewPartialMinifigModal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := chi.URLParam(r, "id")
		id, err := strconv.Atoi(vars)
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		data := map[string]interface{}{
			"ListID": id,
		}

		if err := handler.templates.ExecuteTemplate(w, "new-item-modal", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *PartialMinifigHandler) HandleCreatePartialMinifig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		vars := chi.URLParam(r, "id")
		listID, err := strconv.Atoi(vars)
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		name := r.Form.Get("name")
		description := r.Form.Get("description")

		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		list, err := handler.service.CreatePartialMinifig(listID, name, description, user)
		if err != nil {
			http.Error(w, "Error creating partial minifig", http.StatusInternalServerError)
			return
		}

		// Return updated list items section
		data := map[string]interface{}{
			"List": list,
			"User": user,
		}

		if err := handler.templates.ExecuteTemplate(w, "list-items-section", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *PartialMinifigHandler) HandleEditPartialMinifigModal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		listID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		itemID, err := strconv.Atoi(chi.URLParam(r, "itemId"))
		if err != nil {
			http.Error(w, "Invalid item ID", http.StatusBadRequest)
			return
		}

		item, err := handler.service.GetPartialMinifigByID(itemID, user)
		if err != nil {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}

		data := map[string]interface{}{
			"ListID": listID,
			"Item":   item,
			"User":   user,
		}

		if err := handler.templates.ExecuteTemplate(w, "edit-item-modal", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *PartialMinifigHandler) HandleUpdatePartialMinifig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		listID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		itemID, err := strconv.Atoi(chi.URLParam(r, "itemId"))
		if err != nil {
			http.Error(w, "Invalid item ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		name := r.Form.Get("name")
		description := r.Form.Get("description")

		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		list, err := handler.service.UpdatePartialMinifig(listID, itemID, name, description, user)
		if err != nil {
			http.Error(w, "Error updating partial minifig", http.StatusInternalServerError)
			return
		}

		// Return updated list items section
		data := map[string]interface{}{
			"List": list,
			"User": user,
		}

		if err := handler.templates.ExecuteTemplate(w, "list-items-section", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *PartialMinifigHandler) HandleDeletePartialMinifig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		listID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		itemID, err := strconv.Atoi(chi.URLParam(r, "itemId"))
		if err != nil {
			http.Error(w, "Invalid item ID", http.StatusBadRequest)
			return
		}

		list, err := handler.service.DeletePartialMinifig(listID, itemID, user)
		if err != nil {
			http.Error(w, "Error deleting partial minifig", http.StatusInternalServerError)
			return
		}

		// Return updated list items section
		data := map[string]interface{}{
			"List": list,
			"User": user,
		}

		if err := handler.templates.ExecuteTemplate(w, "list-items-section", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (handler *PartialMinifigHandler) HandleSearchBricklinkItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Starting Bricklink item search request")
		
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Warn("Unauthorized access attempt to Bricklink search")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		logger.Info("User authenticated for Bricklink search", "user_id", user.ID, "email", user.Email)

		if err := r.ParseForm(); err != nil {
			logger.Error("Error parsing form for Bricklink search", "user_id", user.ID, "error", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		bricklinkId := r.Form.Get("bricklink_id")
		if bricklinkId == "" {
			logger.Warn("Bricklink search attempted without ID", "user_id", user.ID)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="error-message" style="color: red; padding: 10px; border: 1px solid red; border-radius: 4px; margin: 10px 0;">Error: Bricklink ID is required</div>`))
			return
		}

		logger.Info("Searching for Bricklink item", "user_id", user.ID, "bricklink_id", bricklinkId)

		// Search for the item using the Bricklink service
		itemData, err := handler.service.GetBricklinkItem(bricklinkId, user)
		if err != nil {
			logger.Error("Failed to retrieve Bricklink item", "user_id", user.ID, "bricklink_id", bricklinkId, "error", err)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="error-message" style="color: red; padding: 10px; border: 1px solid red; border-radius: 4px; margin: 10px 0;">Error: ` + err.Error() + `</div>`))
			return
		}

		logger.Info("Successfully retrieved Bricklink item", "user_id", user.ID, "bricklink_id", bricklinkId, "item_name", itemData.Name, "item_type", itemData.Type)

		data := map[string]interface{}{
			"Item": itemData,
		}

		// Return the item data as a partial template
		if err := handler.templates.ExecuteTemplate(w, "bricklink-item-result", data); err != nil {
			logger.Error("Failed to execute template for Bricklink item result", "user_id", user.ID, "bricklink_id", bricklinkId, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Bricklink item search completed successfully", "user_id", user.ID, "bricklink_id", bricklinkId)
	}
}