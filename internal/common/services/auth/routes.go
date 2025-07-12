package auth

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all auth-related routes
func (handler *AuthHandler) RegisterRoutes(router chi.Router) {
	// Public auth routes (no authentication required)
	router.Route("/auth", func(r chi.Router) {
		// Login routes
		r.Get("/login", handler.HandleLoginPage())
		r.Post("/login", handler.HandleLogin())

		// Registration routes
		r.Get("/register", handler.HandleRegisterPage())
		r.Post("/register", handler.HandleRegister())

		// Logout route
		r.Post("/logout", handler.HandleLogout())
		r.Get("/logout", handler.HandleLogout()) // Allow GET for simple logout links
	})
}
