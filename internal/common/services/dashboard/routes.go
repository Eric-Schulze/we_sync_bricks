package dashboard

import (
	"net/http"

	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all dashboard-related routes
func (handler *DashboardHandler) RegisterRoutes(router chi.Router) {
	// Dashboard API routes for partial updates
	router.Route("/dashboard", func(r chi.Router) {
		// Add auth middleware to protect dashboard routes
		r.Use(auth.Middleware(handler.jwtSecret))

		// Main dashboard route (protected)
		r.Get("/", handler.HandleDashboardPage())

		// Dashboard components for HTMX updates
		r.Get("/stats", handler.HandleDashboardStats())
		r.Get("/activity", handler.HandleDashboardActivity())
		r.Get("/recent-lists", handler.HandleDashboardRecentLists())

		// Dashboard actions
		r.Post("/refresh", handler.HandleRefreshDashboard())
	})

	// Root route redirects to dashboard (also protected)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	})
}
