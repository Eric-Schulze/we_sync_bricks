package init

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/dashboard"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/profile"
	"github.com/eric-schulze/we_sync_bricks/partial_minifigs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var templates *template.Template

func initRouter(app *models.App) *chi.Mux{
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Logger)     // Chi's built-in logger
	router.Use(middleware.Recoverer)  // Error recovery
	
	// Static files
	fileServer(router, "/static/", http.Dir("web/static/"))
	
	// Register service routes using injected handlers
	if authHandler, ok := app.AuthHandler.(*auth.AuthHandler); ok {
		authHandler.RegisterRoutes(router)
	}
	
	if profileHandler, ok := app.ProfileHandler.(*profile.ProfileHandler); ok {
		profileHandler.RegisterRoutes(router)
	}
	
	if partialMinifigHandler, ok := app.PartialMinifigHandler.(*partial_minifigs.PartialMinifigHandler); ok {
		partialMinifigHandler.RegisterRoutes(router)
	}
	
	if dashboardHandler, ok := app.DashboardHandler.(*dashboard.DashboardHandler); ok {
		dashboardHandler.RegisterRoutes(router)
	}

	return router
}

// Helper function for cleaner static file serving
func fileServer(router chi.Router, path string, root http.FileSystem) {
    if strings.ContainsAny(path, "{}*") {
        panic("FileServer does not permit any URL parameters.")
    }

    if path != "/" && path[len(path)-1] != '/' {
        router.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
        path += "/"
    }
    path += "*"

    router.Get(path, func(w http.ResponseWriter, r *http.Request) {
		// Cache static assets for 1 year
        if strings.HasPrefix(r.URL.Path, "/static/js/") || 
           strings.HasPrefix(r.URL.Path, "/static/css/") {
            w.Header().Set("Cache-Control", "public, max-age=31536000")
        }

        rctx := chi.RouteContext(r.Context())
        pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
        fs := http.StripPrefix(pathPrefix, http.FileServer(root))
        fs.ServeHTTP(w, r)
    })
}

// Dashboard handler moved to dashboard service package

// Middleware Adapters

// Example middleware for restricting access to admin users
// would be used like this:
// router.HandleFunc("/admin", adminOnly(handleAdmin)))

// func adminOnly(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if !currentUser(r).IsAdmin {
// 			http.NotFound(w, r)
// 			return
// 		}

// 		h.ServeHTTP(w, r)
// 	})
// }
