package orders

import (
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/go-chi/chi/v5"
)

func (handler *OrdersHandler) RegisterRoutes(router chi.Router) {
	router.Route("/orders", func(router chi.Router) {
		// Add auth middleware to protect all order routes
		router.Use(auth.Middleware(handler.jwtSecret))

		router.Get("/", handler.HandleOrdersPage())
		router.Post("/refresh", handler.HandleRefreshOrders())
		router.Get("/api", handler.HandleOrdersAPI())
	})
}