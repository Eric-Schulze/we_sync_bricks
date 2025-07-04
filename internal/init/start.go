package init

import (
	"context"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/dashboard"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/db"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/profile"
	"github.com/eric-schulze/we_sync_bricks/partial_minifigs"
)


func Start(ctx context.Context, w io.Writer, args []string) error {
	// Create a root context with cancellation
	rootCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Initialize logging system
	logFilePath := "logs/app.log"
	err := logger.InitializeDefaultLogger(logger.LogInfo, logFilePath)
	if err != nil {
		log.Printf("Failed to initialize logger: %v", err)
		// Fall back to standard logging
		log.SetPrefix("we_sync_bricks: ")
		log.SetFlags(0)
	} else {
		logger.Info("Logger initialized successfully", "log_file", logFilePath)
	}

	app, err := initApp(rootCtx)
	if err != nil {
		logger.Error("Failed to initialize app", "error", err)
		return err
	}

	// Create a new router
	router := initRouter(app)

	logger.Info("Starting server on port 4000")
	http.ListenAndServe(":4000", router)

	return nil
}

func initApp(context context.Context) (*models.App, error) {
	var app = models.App{Context: context}

	appConfig, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	dbService, err := db.NewDBService(&appConfig.DBConfig, &app.Context)
	if err != nil {
		return nil, err
	}

	app.DBService = &dbService

	// Initialize templates with custom functions
	// Collect all template files first
	var templateFiles []string
	err = filepath.Walk("web/templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".html" {
			templateFiles = append(templateFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	
	// Parse all templates at once to share the namespace
	templates := template.Must(template.New("").Funcs(GetCustomTemplateFunctions()).ParseFiles(templateFiles...))
	app.Templates = templates

	// Initialize service handlers
	jwtSecret := []byte("your-jwt-secret-key") // TODO: Move to config
	
	app.AuthHandler = auth.InitializeAuthHandler(dbService, templates, jwtSecret)
	app.ProfileHandler = profile.InitializeProfileHandler(dbService, templates, jwtSecret)
	app.PartialMinifigHandler = partial_minifigs.InitializePartialMinifigHandler(dbService, templates, jwtSecret)
	app.DashboardHandler = dashboard.InitializeDashboardHandler(dbService, templates, jwtSecret)

	return &app, nil
}

// func (app *App) retrieveInventory() string{
// 	var blClient = bricklink.NewBricklinkInventoryClient()

// 	var inventoryService = inventory.NewInventoryService(blClient)

// 	messages, err := inventoryService.GetInventory()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	return messages
// }

// func (app *App)  retrieveOrders() string{
// 	var blClient = bricklink.NewBricklinkOrdersClient()

// 	var ordersService = orders.NewOrdersService(blClient)

// 	messages, err := ordersService.GetOrders()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	return messages
// }
