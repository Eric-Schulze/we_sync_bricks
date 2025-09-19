package orders

import (
	"html/template"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/bricklink"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

type OrdersService struct {
	repo            *OrderRepository
	clientManager   *bricklink.ClientManager
}

func NewOrdersService(dbService models.DBService, clientManager *bricklink.ClientManager) *OrdersService {
	return &OrdersService{
		repo:          NewOrderRepository(dbService),
		clientManager: clientManager,
	}
}

// GetOrders retrieves all orders for a user
func (service *OrdersService) GetOrders(userID int64) ([]models.Order, error) {
	orderList, err := service.repo.GetAllOrders(userID)
	if err != nil {
		return nil, err
	}

	// Convert to interface slice
	orders := make([]models.Order, len(orderList))
	for i, order := range orderList {
		orders[i] = &order
	}
	return orders, nil
}

// GetFilteredOrders retrieves orders with filtering, sorting, and pagination
func (service *OrdersService) GetFilteredOrders(userID int64, filters models.OrderFilters) ([]models.Order, int64, error) {
	// Convert abstract filters to repository filters
	repoFilters := OrderFilters{
		Page:     filters.Page,
		Limit:    filters.Limit,
		Sort:     filters.Sort,
		Order:    filters.Order,
		Status:   filters.Status,
		Search:   filters.Search,
		DateFrom: filters.DateFrom,
		DateTo:   filters.DateTo,
	}

	orderList, count, err := service.repo.GetFilteredOrders(userID, repoFilters)
	if err != nil {
		return nil, 0, err
	}

	// Convert to interface slice
	orders := make([]models.Order, len(orderList))
	for i, order := range orderList {
		orders[i] = &order
	}
	return orders, count, nil
}

// GetOrdersCount returns the total count of orders for a user
func (service *OrdersService) GetOrdersCount(userID int64) (int64, error) {
	return service.repo.GetOrdersCount(userID)
}

// SyncOrdersFromBrickLink synchronizes orders from BrickLink API
func (service *OrdersService) SyncOrdersFromBrickLink(user *models.User) (models.OrderSync, error) {
	logger.Info("Starting order sync", "user_id", user.ID, "provider", "bricklink", "service", "OrdersService")

	// Create initial sync record
	orderSync, err := service.repo.CreateOrderSync(user.ID, "in_progress", 0, nil)
	if err != nil {
		logger.Error("Failed to create order sync record", "user_id", user.ID, "provider", "bricklink", "service", "OrdersService", "error", err)
		return nil, err
	}

	// Repository will handle the provider-specific client management

	// Get last sync time
	lastSyncTime, err := service.repo.GetLastSyncTime(user.ID)
	if err != nil {
		logger.Error("Failed to get last sync time", "user_id", user.ID, "error", err)
		// Continue with full sync if we can't get last sync time
	}

	// Get orders from repository - let repository handle provider-specific API calls
	var orders []models.Order
	if lastSyncTime != nil {
		logger.Info("Syncing orders since last sync", "user_id", user.ID, "last_sync", lastSyncTime.Format(time.RFC3339))
		orders, err = service.repo.SyncOrdersSince(user, *lastSyncTime)
	} else {
		logger.Info("Performing full order sync", "user_id", user.ID)
		// For full sync, get orders from the last 6 months to avoid overwhelming the system
		sixMonthsAgo := time.Now().AddDate(0, -6, 0)
		orders, err = service.repo.SyncOrdersSince(user, sixMonthsAgo)
	}

	if err != nil {
		errorMsg := "Failed to retrieve orders from BrickLink: " + err.Error()
		service.repo.UpdateOrderSync(orderSync.ID, "failed", 0, &errorMsg)
		logger.Error("Failed to retrieve orders from BrickLink", "user_id", user.ID, "error", err)
		return nil, err
	}

	logger.Info("Retrieved orders for sync", "user_id", user.ID, "provider", "bricklink", "service", "OrdersService", "count", len(orders))

	// Orders have been processed by repository
	successCount := len(orders)
	newOrdersCount := len(orders) // Repository should return this info in future

	// Update sync record with final status
	finalStatus := "completed"
	var errorMessage *string
	if successCount < len(blOrders) {
		finalStatus = "completed_with_errors"
		msg := "Some orders failed to sync"
		errorMessage = &msg
	}

	updatedSync, err := service.repo.UpdateOrderSync(orderSync.ID, finalStatus, newOrdersCount, errorMessage)
	if err != nil {
		logger.Error("Failed to update order sync record", "sync_id", orderSync.ID, "error", err)
		return orderSync, err // Return original sync record
	}

	logger.Info("Completed order sync", "user_id", user.ID, "provider", "bricklink", "service", "OrdersService", "total_orders", len(orders), "processed", successCount, "new_orders", newOrdersCount, "status", finalStatus)
	return updatedSync, nil
}

// InitializeOrdersHandler creates a fully initialized orders handler with all dependencies
func InitializeOrdersHandler(dbService models.DBService, templates *template.Template, jwtSecret []byte) *OrdersHandler {
	// Initialize client manager with 30-minute cache TTL and max 100 clients
	clientManager := bricklink.NewClientManager(30*time.Minute, 100, dbService)

	service := NewOrdersService(dbService, clientManager)
	handler := NewOrdersHandler(service, templates, jwtSecret)
	return handler
}
