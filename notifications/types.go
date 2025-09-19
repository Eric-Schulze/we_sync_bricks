package notifications

import "github.com/eric-schulze/we_sync_bricks/internal/common/models"

// NotificationClient defines the interface that provider-specific clients must implement
type NotificationClient interface {
	// GetNotifications retrieves unread notifications for a user
	GetNotifications(user *models.User) ([]models.Notification, error)
	
	// ProcessNotification handles a specific notification based on its type
	ProcessNotification(notification models.Notification) error
	
	// GetProvider returns the provider this client handles
	GetProvider() string
}

// NotificationService orchestrates notification processing across providers
type NotificationService interface {
	// ProcessNotificationsForUser processes notifications for a specific user from all registered providers
	ProcessNotificationsForUser(userID int) error
	
	// RegisterClient registers a provider-specific notification client
	RegisterClient(client NotificationClient)
	
	// ProcessWebhookForUser processes a webhook notification for a specific user and provider
	ProcessWebhookForUser(userID int, provider string) error
}