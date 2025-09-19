package partial_minifigs

import (
	"encoding/json"
	"fmt"
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
			logger.Debug("Starting HandlePartialMinifigListsPage request", "method", r.Method, "url", r.URL.String(), "remote_addr", r.RemoteAddr)

			// Get current user from context (set by auth middleware)
			user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
			if !ok {
				logger.Error("User not found in context - authentication failed", "context_keys", r.Context())
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			logger.Debug("User authenticated successfully", "user_id", user.ID, "email", user.Email)

			logger.Debug("Calling service to get all partial minifig lists", "user_id", user.ID)
			lists, err := handler.service.GetAllPartialMinifigLists(user)
			if err != nil {
				logger.Error("Service error when getting lists", "user_id", user.ID, "error", err, "error_type", fmt.Sprintf("%T", err))
				http.Error(w, "Error loading lists", http.StatusInternalServerError)
				return
			}

			logger.Debug("Successfully retrieved lists from service", "user_id", user.ID, "list_count", len(lists))

			data := map[string]interface{}{
				"Title":       "Partial Minifigs Lists",
				"CurrentPage": "partial_minifigs_lists",
				"Lists":       lists,
				"User":        user,
			}

			logger.Debug("Template data prepared", "user_id", user.ID, "data_keys", getMapKeys(data))

			// Check if this is an HTMX request
			isHTMX := r.Header.Get("HX-Request") == "true"
			logger.Debug("Request type determined", "user_id", user.ID, "is_htmx", isHTMX)

			if isHTMX {
				logger.Debug("Executing HTMX template", "user_id", user.ID, "template", "partial-minifigs-lists-content")
				if err := handler.templates.ExecuteTemplate(w, "partial-minifigs-lists-content", data); err != nil {
					logger.Error("Template execution error for HTMX request", "user_id", user.ID, "template", "partial-minifigs-lists-content", "error_message", err.Error(), "error_type", fmt.Sprintf("%T", err))
					http.Error(w, err.Error(), http.StatusInternalServerError)
				} else {
					logger.Debug("HTMX template executed successfully", "user_id", user.ID)
				}
				return
			}

			// If not an HTMX request, render the full page
			logger.Debug("Executing full page template", "user_id", user.ID, "template", "partial_minifig_lists")
			if err := handler.templates.ExecuteTemplate(w, "partial_minifig_lists", data); err != nil {
				logger.Error("Template execution error for full page request", "user_id", user.ID, "template", "partial_minifig_lists", "error_message", err.Error(), "error_type", fmt.Sprintf("%T", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			logger.Debug("Full page template executed successfully", "user_id", user.ID)
		},
	)
}

func (handler *PartialMinifigHandler) HandlePartialMinifigListDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Starting HandlePartialMinifigListDetail request", "method", r.Method, "url", r.URL.String(), "remote_addr", r.RemoteAddr)

		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Error("User not found in context - authentication failed for list detail", "context_keys", r.Context())
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		logger.Debug("User authenticated successfully for list detail", "user_id", user.ID, "email", user.Email)

		vars := chi.URLParam(r, "id")
		logger.Debug("Extracting list ID from URL params", "user_id", user.ID, "id_param", vars)

		id, err := strconv.ParseInt(vars, 10, 64)
		if err != nil {
			logger.Error("Invalid list ID format", "user_id", user.ID, "id_param", vars, "error", err, "error_type", fmt.Sprintf("%T", err))
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		logger.Debug("Calling service to get partial minifig list by ID", "user_id", user.ID, "list_id", id)
		list, err := handler.service.GetPartialMinifigListByID(id, user)
		if err != nil {
			logger.Error("Service error when getting list by ID", "user_id", user.ID, "list_id", id, "error", err, "error_type", fmt.Sprintf("%T", err))
			http.Error(w, "List not found", http.StatusNotFound)
			return
		}

		logger.Debug("Successfully retrieved list from service", "user_id", user.ID, "list_id", id, "list_name", list.Name, "partial_minifigs_count", len(list.PartialMinifigs))

		data := map[string]interface{}{
			"Title":       list.Name,
			"CurrentPage": "lists",
			"List":        list,
			"User":        user,
		}

		logger.Debug("Template data prepared for list detail", "user_id", user.ID, "list_id", id, "data_keys", getMapKeys(data))

		// Check if this is an HTMX request
		isHTMX := r.Header.Get("HX-Request") == "true"
		logger.Debug("Request type determined for list detail", "user_id", user.ID, "list_id", id, "is_htmx", isHTMX)

		if isHTMX {
			logger.Debug("Executing HTMX template for list detail", "user_id", user.ID, "list_id", id, "template", "partial-minifig-list-content")
			if err := handler.templates.ExecuteTemplate(w, "partial-minifig-list-content", data); err != nil {
				logger.Error("Template execution error for HTMX list detail request", "user_id", user.ID, "list_id", id, "template", "partial-minifig-list-content", "error_message", err.Error(), "error_type", fmt.Sprintf("%T", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				logger.Debug("HTMX template executed successfully for list detail", "user_id", user.ID, "list_id", id)
			}
			return
		}

		// If not an HTMX request, render the full page
		logger.Debug("Executing full page template for list detail", "user_id", user.ID, "list_id", id, "template", "partial_minifig_list")
		if err := handler.templates.ExecuteTemplate(w, "partial_minifig_list", data); err != nil {
			logger.Error("Template execution error for full page list detail request", "user_id", user.ID, "list_id", id, "template", "partial_minifig_list", "error_message", err.Error(), "error_type", fmt.Sprintf("%T", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			logger.Debug("Full page template executed successfully for list detail", "user_id", user.ID, "list_id", id)
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

		name := r.FormValue("name")
		description := r.FormValue("description")

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
			w.Header().Set("HX-Redirect", "/partial-minifigs-lists/"+strconv.FormatInt(newList.ID, 10))
			w.WriteHeader(http.StatusOK)
			return
		}

		// For non-HTMX requests, do a regular redirect
		http.Redirect(w, r, "/partial-minifigs-lists/"+strconv.FormatInt(newList.ID, 10), http.StatusSeeOther)
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
		id, err := strconv.ParseInt(vars, 10, 64)
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
		id, err := strconv.ParseInt(vars, 10, 64)
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		description := r.FormValue("description")

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
		id, err := strconv.ParseInt(vars, 10, 64)
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
		listID, err := strconv.ParseInt(vars, 10, 64)
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		description := r.FormValue("description")

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

		listID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
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

		listID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid item ID", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		description := r.FormValue("description")

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

		listID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		itemID, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
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

		bricklinkId := r.FormValue("bricklink_id")
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
			w.Write([]byte(`<div class="error-message flex-auto break-all" style="color: red; padding: 10px; border: 1px solid red; border-radius: 4px; margin: 10px 0;">Error: ` + err.Error() + `</div>`))
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

func (handler *PartialMinifigHandler) HandleGetMinifigPicture() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Starting minifig picture request")

		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Warn("Unauthorized access attempt to minifig picture")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		bricklinkId := r.FormValue("bricklink_id")
		if bricklinkId == "" {
			logger.Warn("Minifig picture request without ID", "user_id", user.ID)
			http.Error(w, "Bricklink ID is required", http.StatusBadRequest)
			return
		}

		logger.Info("Getting minifig picture", "user_id", user.ID, "bricklink_id", bricklinkId)

		// Get picture data from Bricklink service
		pictureData, err := handler.service.GetMinifigPicture(bricklinkId, user)
		if err != nil {
			logger.Error("Failed to get minifig picture", "user_id", user.ID, "bricklink_id", bricklinkId, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Successfully retrieved minifig picture", "user_id", user.ID, "bricklink_id", bricklinkId)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(pictureData))
	}
}

func (handler *PartialMinifigHandler) HandleGetMinifigPricing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Starting minifig pricing request")

		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Warn("Unauthorized access attempt to minifig pricing")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		bricklinkId := r.FormValue("bricklink_id")
		condition := r.FormValue("condition")
		if bricklinkId == "" {
			logger.Warn("Minifig pricing request without ID", "user_id", user.ID)
			http.Error(w, "Bricklink ID is required", http.StatusBadRequest)
			return
		}

		// Default to new condition if not specified
		if condition == "" {
			condition = "N"
		}

		logger.Info("Getting minifig pricing", "user_id", user.ID, "bricklink_id", bricklinkId, "condition", condition)

		// Get pricing data from Bricklink service
		pricingData, err := handler.service.GetMinifigPricing(bricklinkId, condition, user)
		if err != nil {
			logger.Error("Failed to get minifig pricing", "user_id", user.ID, "bricklink_id", bricklinkId, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Successfully retrieved minifig pricing", "user_id", user.ID, "bricklink_id", bricklinkId)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(pricingData))
	}
}

func (handler *PartialMinifigHandler) HandleGetMinifigParts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Starting minifig parts request")

		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Warn("Unauthorized access attempt to minifig parts")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		bricklinkId := r.FormValue("bricklink_id")
		if bricklinkId == "" {
			logger.Warn("Minifig parts request without ID", "user_id", user.ID)
			http.Error(w, "Bricklink ID is required", http.StatusBadRequest)
			return
		}

		logger.Info("Getting minifig parts", "user_id", user.ID, "bricklink_id", bricklinkId)

		// Get parts data from Bricklink service
		partsData, err := handler.service.GetMinifigParts(bricklinkId, user)
		if err != nil {
			logger.Error("Failed to get minifig parts", "user_id", user.ID, "bricklink_id", bricklinkId, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Successfully retrieved minifig parts", "user_id", user.ID, "bricklink_id", bricklinkId)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(partsData))
	}
}

func (handler *PartialMinifigHandler) HandleGetPartPricing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Starting part pricing request")

		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Warn("Unauthorized access attempt to part pricing")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		itemType := r.FormValue("item_type")
		itemID := r.FormValue("item_id")
		colorIDStr := r.FormValue("color_id")
		condition := r.FormValue("condition")

		if itemID == "" {
			logger.Warn("Part pricing request without item ID", "user_id", user.ID)
			http.Error(w, "Item ID is required", http.StatusBadRequest)
			return
		}

		// Default to PART if no type specified
		if itemType == "" {
			itemType = "PART"
		}

		// Default to new condition if not specified
		if condition == "" {
			condition = "N"
		}

		// Parse color ID
		var colorID int
		if colorIDStr != "" {
			var err error
			colorID, err = strconv.Atoi(colorIDStr)
			if err != nil {
				logger.Warn("Invalid color ID format", "user_id", user.ID, "color_id", colorIDStr)
				http.Error(w, "Invalid color ID", http.StatusBadRequest)
				return
			}
		}

		logger.Info("Getting part pricing", "user_id", user.ID, "item_type", itemType, "item_id", itemID, "color_id", colorID, "condition", condition)

		// Get pricing data from Bricklink service
		pricingData, err := handler.service.GetPartPricing(itemType, itemID, colorID, condition, user)
		if err != nil {
			logger.Error("Failed to get part pricing", "user_id", user.ID, "item_type", itemType, "item_id", itemID, "color_id", colorID, "condition", condition, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Successfully retrieved part pricing", "user_id", user.ID, "item_type", itemType, "item_id", itemID, "color_id", colorID, "condition", condition)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(pricingData))
	}
}

func (handler *PartialMinifigHandler) HandleGetPartPicture() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Starting part picture request")

		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Warn("Unauthorized access attempt to part picture")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		itemType := r.FormValue("item_type")
		itemID := r.FormValue("item_id")
		colorIDStr := r.FormValue("color_id")

		if itemID == "" {
			logger.Warn("Part picture request without item ID", "user_id", user.ID)
			http.Error(w, "Item ID is required", http.StatusBadRequest)
			return
		}

		// Default to PART if no type specified
		if itemType == "" {
			itemType = "PART"
		}

		// Parse color ID
		var colorID int
		if colorIDStr != "" {
			var err error
			colorID, err = strconv.Atoi(colorIDStr)
			if err != nil {
				logger.Warn("Invalid color ID format", "user_id", user.ID, "color_id", colorIDStr)
				http.Error(w, "Invalid color ID", http.StatusBadRequest)
				return
			}
		}

		logger.Info("Getting part picture", "user_id", user.ID, "item_type", itemType, "item_id", itemID, "color_id", colorID)

		// Get picture data from Bricklink service
		pictureData, err := handler.service.GetPartPicture(itemType, itemID, colorID, user)
		if err != nil {
			logger.Error("Failed to get part picture", "user_id", user.ID, "item_type", itemType, "item_id", itemID, "color_id", colorID, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Successfully retrieved part picture", "user_id", user.ID, "item_type", itemType, "item_id", itemID, "color_id", colorID)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(pictureData))
	}
}

func (handler *PartialMinifigHandler) HandleAddMinifigWithParts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Starting add minifig with parts request")

		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Warn("Unauthorized access attempt to add minifig with parts")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		listIDStr := r.FormValue("list_id")
		minifigID := r.FormValue("minifig_id")
		minifigName := r.FormValue("minifig_name")
		selectedPartsJSON := r.FormValue("selected_parts")
		referenceID := r.FormValue("reference_id")
		condition := r.FormValue("condition")
		notes := r.FormValue("notes")

		if listIDStr == "" || minifigID == "" || selectedPartsJSON == "" {
			logger.Warn("Add minifig with parts request missing required fields", "user_id", user.ID)
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		listID, err := strconv.ParseInt(listIDStr, 10, 64)
		if err != nil {
			logger.Warn("Invalid list ID format", "user_id", user.ID, "list_id", listIDStr)
			http.Error(w, "Invalid list ID", http.StatusBadRequest)
			return
		}

		// Parse selected parts JSON
		var selectedParts []SelectedPart
		if err := json.Unmarshal([]byte(selectedPartsJSON), &selectedParts); err != nil {
			logger.Error("Failed to parse selected parts JSON", "user_id", user.ID, "error", err)
			http.Error(w, "Invalid selected parts data", http.StatusBadRequest)
			return
		}

		logger.Info("Adding minifig with parts", "user_id", user.ID, "list_id", listID, "minifig_id", minifigID, "parts_count", len(selectedParts))

		// Convert empty strings to nil pointers
		var refID, cond, notesPtr *string
		if referenceID != "" {
			refID = &referenceID
		}
		if condition != "" {
			cond = &condition
		}
		if notes != "" {
			notesPtr = &notes
		}

		// Add minifig with parts using the service
		err = handler.service.AddMinifigWithPartsNoReturn(listID, minifigID, minifigName, refID, cond, notesPtr, selectedParts, user)
		if err != nil {
			logger.Error("Failed to add minifig with parts", "user_id", user.ID, "list_id", listID, "minifig_id", minifigID, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Successfully added minifig with parts", "user_id", user.ID, "list_id", listID, "minifig_id", minifigID)

		// Return success status without data
		w.WriteHeader(http.StatusOK)
	}
}

func (handler *PartialMinifigHandler) HandleMinifigDetailsModal() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		itemID := r.FormValue("item_id")
		listIDStr := r.FormValue("list_id")
		partsJSON := r.FormValue("parts")

		if itemID == "" || listIDStr == "" {
			logger.Error("Missing required parameters", "item_id", itemID, "list_id", listIDStr)
			http.Error(w, "Missing required parameters", http.StatusBadRequest)
			return
		}

		listID, err := strconv.ParseInt(listIDStr, 10, 64)
		if err != nil {
			logger.Error("Invalid list_id", "list_id", listIDStr, "error", err)
			http.Error(w, "Invalid list_id", http.StatusBadRequest)
			return
		}

		// Parse parts JSON if provided
		var parts []map[string]interface{}
		if partsJSON != "" {
			if err := json.Unmarshal([]byte(partsJSON), &parts); err != nil {
				logger.Error("Failed to parse parts JSON", "error", err, "json", partsJSON)
				http.Error(w, "Invalid parts data", http.StatusBadRequest)
				return
			}
		}

		// Get the list to ensure user has access
		list, err := handler.service.GetPartialMinifigListByID(listID, user)
		if err != nil {
			logger.Error("Failed to get list", "list_id", listID, "user_id", user.ID, "error", err)
			http.Error(w, "List not found", http.StatusNotFound)
			return
		}

		logger.Info("Rendering minifig details modal", "user_id", user.ID, "item_id", itemID, "list_id", listID, "parts_count", len(parts))

		// Prepare template data
		data := map[string]interface{}{
			"ItemID": itemID,
			"List":   list,
			"Parts":  parts,
			"User":   user,
		}

		// Render the modal template
		if err := handler.templates.ExecuteTemplate(w, "add-minifig-details-modal", data); err != nil {
			logger.Error("Failed to execute modal template", "user_id", user.ID, "error", err)
			http.Error(w, "Template error", http.StatusInternalServerError)
		}
	})
}

// Helper function to get map keys for logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
