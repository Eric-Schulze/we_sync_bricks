package notifications

import (
	"fmt"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

// NotificationManager is the implementation of the NotificationService interface
type NotificationManager struct {
	clients   map[string]NotificationClient
	dbService models.DBService
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(dbService models.DBService) *NotificationManager {
	return &NotificationManager{
		clients:   make(map[string]NotificationClient),
		dbService: dbService,
	}
}

// RegisterClient registers a provider-specific notification client
func (m *NotificationManager) RegisterClient(client NotificationClient) {
	provider := client.GetProvider()
	m.clients[provider] = client
	logger.Info("Registered notification client", "provider", provider)
}

// ProcessNotificationsForUser processes notifications for a specific user from all registered providers
func (m *NotificationManager) ProcessNotificationsForUser(userID int) error {
	logger.Info("Processing notifications for user from all providers", "user_id", userID)
	
	// Get user from database
	user, err := m.getUserByID(int64(userID))
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", userID, err)
	}
	
	// Process notifications from each registered provider
	for provider, client := range m.clients {
		logger.Info("Processing notifications from provider", "user_id", userID, "provider", provider)
		
		notifications, err := client.GetNotifications(user)
		if err != nil {
			logger.Error("Failed to get notifications from provider", "user_id", userID, "provider", provider, "error", err)
			// Continue processing other providers even if one fails
			continue
		}
		
		// Process each notification
		for _, notification := range notifications {
			err := client.ProcessNotification(notification)
			if err != nil {
				logger.Error("Failed to process notification", 
					"user_id", userID, 
					"provider", provider,
					"notification_id", notification.ID,
					"error", err)
				// Continue processing other notifications even if one fails
				continue
			}
		}
		
		logger.Info("Successfully processed notifications from provider", 
			"user_id", userID, 
			"provider", provider, 
			"count", len(notifications))
	}
	
	return nil
}

// ProcessWebhookForUser processes a webhook notification for a specific user and provider
func (m *NotificationManager) ProcessWebhookForUser(userID int, provider string) error {
	logger.Info("Processing webhook for user and provider", "user_id", userID, "provider", provider)
	
	client, exists := m.clients[provider]
	if !exists {
		return fmt.Errorf("no client registered for provider: %s", provider)
	}
	
	// Get user from database
	user, err := m.getUserByID(int64(userID))
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", userID, err)
	}
	
	// Get notifications from the specific provider
	notifications, err := client.GetNotifications(user)
	if err != nil {
		return fmt.Errorf("failed to get notifications from %s for user %d: %w", provider, userID, err)
	}
	
	// Process each notification
	for _, notification := range notifications {
		err := client.ProcessNotification(notification)
		if err != nil {
			logger.Error("Failed to process notification", 
				"user_id", userID, 
				"provider", provider,
				"notification_id", notification.ID,
				"error", err)
			// Continue processing other notifications even if one fails
			continue
		}
	}
	
	logger.Info("Successfully processed webhook notifications", 
		"user_id", userID, 
		"provider", provider, 
		"count", len(notifications))
	
	return nil
}

// getUserByID retrieves a user by ID from the database using raw SQL
func (m *NotificationManager) getUserByID(userID int64) (*models.User, error) {
	query := "SELECT id, username, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1"
	rows, err := m.dbService.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user %d: %w", userID, err)
	}
	defer rows.Close()
	
	if !rows.Next() {
		return nil, fmt.Errorf("user %d not found", userID)
	}
	
	var user models.User
	err = rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user %d: %w", userID, err)
	}
	
	return &user, nil
}