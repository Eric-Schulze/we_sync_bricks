package services

import (
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/bricklink"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/user"
	"github.com/eric-schulze/we_sync_bricks/notifications"
	"github.com/eric-schulze/we_sync_bricks/orders"
)

// ServiceProvider holds all application services
type ServiceProvider struct {
	User          UserService
	Orders        OrdersService
	Notifications NotificationService
	Auth          AuthService
}

// Global service provider instance
var Services *ServiceProvider

// InitializeServices creates and initializes all services
func InitializeServices(dbService models.DBService, jwtSecret []byte) *ServiceProvider {
	// Initialize repositories
	userRepo := user.NewUserRepository(dbService)
	authRepo := auth.NewAuthRepository(dbService)

	// Initialize core services
	userService := user.NewUserService(userRepo)
	authService := auth.NewAuthService(authRepo, jwtSecret)

	// Initialize BrickLink client manager
	clientManager := bricklink.NewClientManager(30*time.Minute, 100, dbService)

	// Initialize orders service
	ordersService := orders.NewOrdersService(dbService, clientManager)

	// Initialize notification manager
	notificationManager := notifications.NewNotificationManager(dbService)

	// Create BrickLink notifications client and register it
	bricklinkNotificationsClient := bricklink.NewBricklinkNotificationsClient(clientManager)
	notificationManager.RegisterClient(bricklinkNotificationsClient)

	// Create service provider
	provider := &ServiceProvider{
		User:          userService,
		Orders:        ordersService,
		Notifications: notificationManager,
		Auth:          authService,
	}

	// Set global instance
	Services = provider

	return provider
}

// GetServices returns the global service provider instance
func GetServices() *ServiceProvider {
	if Services == nil {
		panic("Services not initialized. Call InitializeServices first.")
	}
	return Services
}