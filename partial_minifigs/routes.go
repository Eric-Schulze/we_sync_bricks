package partial_minifigs

import (
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/go-chi/chi/v5"
)

func (handler *PartialMinifigHandler) RegisterRoutes(router chi.Router) {
	router.Route("/partial-minifigs-lists", func(router chi.Router) {
		// Add auth middleware to protect all partial minifigs routes
		router.Use(auth.Middleware(handler.jwtSecret))

		router.Get("/", handler.HandlePartialMinifigListsPage())
		router.Post("/", handler.HandleCreatePartialMinifigList())
		router.Get("/new", handler.HandleNewPartialMinifigListModal())
		router.Get("/{id}", handler.HandlePartialMinifigListDetail())
		router.Put("/{id}", handler.HandleUpdatePartialMinifigList())
		router.Get("/{id}/edit", handler.HandleEditPartialMinifigListModal())

		router.Route("/{id}/partial-minifig", func(router chi.Router) {
			router.Post("/", handler.HandleCreatePartialMinifig())
			router.Get("/new", handler.HandleNewPartialMinifigModal())
			router.Put("/{itemId}", handler.HandleUpdatePartialMinifig())
			router.Delete("/{itemId}", handler.HandleDeletePartialMinifig())
			router.Get("/{itemId}/edit", handler.HandleEditPartialMinifigModal())
		})

		// Search endpoints
		router.Post("/search-bricklink", handler.HandleSearchBricklinkItem())

		// Additional Bricklink data endpoints
		router.Post("/minifig-picture", handler.HandleGetMinifigPicture())
		router.Post("/minifig-pricing", handler.HandleGetMinifigPricing())
		router.Post("/minifig-parts", handler.HandleGetMinifigParts())
		router.Post("/part-picture", handler.HandleGetPartPicture())
		router.Post("/part-pricing", handler.HandleGetPartPricing())
		router.Post("/add-minifig-with-parts", handler.HandleAddMinifigWithParts())
	})
}
