package webhooks

import (
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/bricklink"
	"github.com/eric-schulze/we_sync_bricks/notifications"
)

// WebhookService handles webhook-related business logic
type WebhookService struct {
	notificationService notifications.NotificationService
}

// NewWebhookService creates a new webhook service
func NewWebhookService(notificationService notifications.NotificationService) *WebhookService {
	return &WebhookService{
		notificationService: notificationService,
	}
}

// ProcessWebhookForUser processes a webhook for a specific user and provider
func (s *WebhookService) ProcessWebhookForUser(userID int, provider string) error {
	return s.notificationService.ProcessWebhookForUser(userID, provider)
}

// InitializeWebhookHandler creates a fully initialized webhook handler with all dependencies
func InitializeWebhookHandler(app *models.App) *WebhookHandler {
	// Create notification manager
	notificationManager := notifications.NewNotificationManager(*app.DBService)
	
	// Create BrickLink client manager for notifications
	clientManager := bricklink.NewClientManager(30*time.Minute, 100, *app.DBService)
	
	// Create and register BrickLink notifications client
	bricklinkClient := bricklink.NewBricklinkNotificationsClient(clientManager)
	notificationManager.RegisterClient(bricklinkClient)
	
	// Create webhook service
	service := NewWebhookService(notificationManager)
	
	// Create and return webhook handler
	return NewWebhookHandler(app, service.notificationService)
}