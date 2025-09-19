package profile

import (
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all profile-related routes
func (handler *ProfileHandler) RegisterRoutes(router chi.Router) {
	// Profile routes (require authentication)
	router.Route("/profile", func(r chi.Router) {
		// Add auth middleware to protect all profile routes
		r.Use(auth.Middleware(handler.jwtSecret))

		r.Get("/", handler.HandleProfilePage())
		r.Get("/edit", handler.HandleEditProfilePage())
		r.Post("/edit", handler.HandleUpdateProfile())
		r.Put("/", handler.HandleUpdateProfile()) // Alternative for REST compliance

		// Password management
		r.Post("/change-password", handler.HandleChangePassword())

		// API key management
		r.Post("/api-keys", handler.HandleUpdateAPIKeys())
		r.Delete("/api-keys", handler.HandleDeleteAPIKeys())
	})

	// Account Settings routes (require authentication)
	router.Route("/account-settings", func(r chi.Router) {
		// Add auth middleware to protect all account settings routes
		r.Use(auth.Middleware(handler.jwtSecret))

		r.Get("/", handler.HandleAccountSettingsPage())
	})
}
