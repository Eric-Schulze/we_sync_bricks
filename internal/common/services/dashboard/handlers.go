package dashboard

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

type DashboardHandler struct {
	service   *DashboardService
	templates *template.Template
	jwtSecret []byte
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(service *DashboardService, templates *template.Template, jwtSecret []byte) *DashboardHandler {
	return &DashboardHandler{
		service:   service,
		templates: templates,
		jwtSecret: jwtSecret,
	}
}

// HandleDashboardPage displays the main dashboard page
func (h *DashboardHandler) HandleDashboardPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			// For now, use a default user ID if not authenticated
			// TODO: Redirect to login page when auth is fully implemented
			// http.Redirect(w, r, "/login", http.StatusSeeOther)
			// return
			user = &models.User{ID: 1} // Default user for development
		}

		// Get dashboard data
		dashboardData, err := h.service.GetDashboardData(user.ID)
		if err != nil {
			logger.Error("Error getting dashboard data", "user_id", user.ID, "error", err)
			http.Error(w, "Error loading dashboard", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Title":         "Dashboard",
			"CurrentPage":   "dashboard",
			"User":          user,
			"DashboardData": dashboardData,
		}

		logger.Info("Handling request to Dashboard page", "user_id", user.ID)

		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			if err := h.templates.ExecuteTemplate(w, "dashboard-content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// If not an HTMX request, render the full page
		// Execute the dashboard template
		if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
			logger.Error("Error handling request to Dashboard page", "user_id", user.ID, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// HandleDashboardStats returns just the statistics portion of the dashboard
func (h *DashboardHandler) HandleDashboardStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			user = &models.User{ID: 1} // Default user for development
		}

		// Get just the stats
		stats, err := h.service.GetUserStats(user.ID)
		if err != nil {
			logger.Error("Error getting dashboard stats", "user_id", user.ID, "error", err)
			http.Error(w, "Error loading stats", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Stats": stats,
		}

		// Return stats component
		if err := h.templates.ExecuteTemplate(w, "dashboard-stats", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// HandleDashboardActivity returns the activity feed
func (h *DashboardHandler) HandleDashboardActivity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			user = &models.User{ID: 1} // Default user for development
		}

		// Get limit from query parameter
		limitStr := r.URL.Query().Get("limit")
		limit := 10 // default
		if limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		// Get recent activity
		activity, err := h.service.GetRecentActivity(user.ID, limit)
		if err != nil {
			logger.Error("Error getting dashboard activity", "user_id", user.ID, "error", err)
			http.Error(w, "Error loading activity", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Activity": activity,
		}

		// Return activity component
		if err := h.templates.ExecuteTemplate(w, "dashboard-activity", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// HandleDashboardRecentLists returns recent lists for the dashboard
func (h *DashboardHandler) HandleDashboardRecentLists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			user = &models.User{ID: 1} // Default user for development
		}

		// Get limit from query parameter
		limitStr := r.URL.Query().Get("limit")
		limit := 10 // default
		if limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		// Get recent lists
		recentLists, err := h.service.GetRecentLists(user.ID, limit)
		if err != nil {
			logger.Error("Error getting recent lists", "user_id", user.ID, "error", err)
			http.Error(w, "Error loading recent lists", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Lists": recentLists,
		}

		// Return recent lists component
		if err := h.templates.ExecuteTemplate(w, "dashboard-recent-lists", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// HandleRefreshDashboard refreshes dashboard data
func (h *DashboardHandler) HandleRefreshDashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			user = &models.User{ID: 1} // Default user for development
		}

		// Refresh dashboard data
		err := h.service.RefreshDashboardData(user.ID)
		if err != nil {
			logger.Error("Error refreshing dashboard data", "user_id", user.ID, "error", err)
			http.Error(w, "Error refreshing dashboard", http.StatusInternalServerError)
			return
		}

		logger.Info("Dashboard data refreshed", "user_id", user.ID)

		// Return success response or redirect back to dashboard
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Refresh", "true")
			w.WriteHeader(http.StatusOK)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}
