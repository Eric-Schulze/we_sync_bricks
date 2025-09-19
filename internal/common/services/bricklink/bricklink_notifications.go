package bricklink

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/eric-schulze/we_sync_bricks/internal/services"
)

// BLNotification represents a BrickLink-specific notification
type BLNotification struct {
	NotificationID int64     `json:"notification_id"`
	Type           string    `json:"type"`
	DateCreated    time.Time `json:"date_created"`
	ResourceType   string    `json:"resource_type"`
	ResourceID     string    `json:"resource_id"`
	Message        string    `json:"message"`
}

// BLNotificationResponse represents the response from Get Notifications API
type BLNotificationResponse struct {
	Meta BLResponseMeta    `json:"meta"`
	Data []BLNotification  `json:"data"`
}

// BricklinkNotificationsClient handles BrickLink-specific notification operations
type BricklinkNotificationsClient struct {
	clientManager *ClientManager
}

// NewBricklinkNotificationsClient creates a new BrickLink notifications client
func NewBricklinkNotificationsClient(clientManager *ClientManager) *BricklinkNotificationsClient {
	return &BricklinkNotificationsClient{
		clientManager: clientManager,
	}
}

// GetProvider returns the provider name this client handles
func (c *BricklinkNotificationsClient) GetProvider() string {
	return "bricklink"
}

// GetNotifications retrieves unread notifications for a user from BrickLink
func (c *BricklinkNotificationsClient) GetNotifications(user *models.User) ([]models.Notification, error) {
	logger.Info("Fetching BrickLink notifications for user", "user_id", user.ID)
	
	// Get base client for the user
	client, err := c.clientManager.GetBaseClient(user)
	if err != nil {
		logger.Error("Failed to create BrickLink client for user", "user_id", user.ID, "error", err)
		return nil, err
	}
	
	response, err := client.BLGet("/notifications")
	if err != nil {
		logger.Error("Failed to fetch notifications from BrickLink", "user_id", user.ID, "error", err)
		return nil, err
	}

	// Check if the API call was successful
	if response.Meta.Code != 200 {
		logger.Error("BrickLink API returned error", "user_id", user.ID, "code", response.Meta.Code, "message", response.Meta.Message)
		return nil, fmt.Errorf("BrickLink API error: %s", response.Meta.Message)
	}

	// Parse the notifications data
	var blNotifications []BLNotification
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		logger.Error("Failed to marshal notification data", "user_id", user.ID, "error", err)
		return nil, err
	}

	err = json.Unmarshal(dataBytes, &blNotifications)
	if err != nil {
		logger.Error("Failed to parse notifications", "user_id", user.ID, "error", err)
		return nil, err
	}

	// Convert BrickLink notifications to generic notifications
	notifications := make([]models.Notification, len(blNotifications))
	for i, blNotification := range blNotifications {
		notifications[i] = models.Notification{
			ID:           strconv.FormatInt(blNotification.NotificationID, 10),
			Provider:     "bricklink",
			Type:         models.NotificationType(blNotification.Type),
			ResourceType: blNotification.ResourceType,
			ResourceID:   blNotification.ResourceID,
			Message:      blNotification.Message,
			DateCreated:  blNotification.DateCreated,
			UserID:       user.ID,
		}
	}

	logger.Info("Successfully fetched BrickLink notifications", "user_id", user.ID, "count", len(notifications))
	return notifications, nil
}

// ProcessNotification handles a specific BrickLink notification based on its type
func (c *BricklinkNotificationsClient) ProcessNotification(notification models.Notification) error {
	logger.Info("Processing BrickLink notification", 
		"notification_id", notification.ID,
		"type", notification.Type,
		"resource_type", notification.ResourceType,
		"resource_id", notification.ResourceID,
		"user_id", notification.UserID)

	switch notification.Type {
	case models.TypeOrderNew:
		return c.processNewOrderNotification(notification)
	case models.TypeOrderStatusChanged:
		return c.processOrderStatusNotification(notification)
	case models.TypeOrderItemsChanged:
		return c.processOrderItemsNotification(notification)
	case models.TypeMessageNew:
		return c.processNewMessageNotification(notification)
	case models.TypeFeedbackNew:
		return c.processNewFeedbackNotification(notification)
	default:
		logger.Info("Unknown notification type, skipping", 
			"type", notification.Type,
			"notification_id", notification.ID)
		return nil
	}
}

// processNewOrderNotification handles new order notifications
func (c *BricklinkNotificationsClient) processNewOrderNotification(notification models.Notification) error {
	logger.Info("Processing new order notification",
		"order_id", notification.ResourceID,
		"user_id", notification.UserID)

	// Get user using the global service provider
	user, err := services.GetServices().User.GetUserByID(notification.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", notification.UserID, err)
	}

	// Trigger order sync using the global service provider
	_, err = services.GetServices().Orders.SyncOrdersFromBrickLink(user)
	if err != nil {
		return fmt.Errorf("failed to sync orders after notification for user %d: %w", notification.UserID, err)
	}

	logger.Info("New order notification processed - order sync triggered", "order_id", notification.ResourceID, "user_id", notification.UserID)
	return nil
}

// processOrderStatusNotification handles order status change notifications
func (c *BricklinkNotificationsClient) processOrderStatusNotification(notification models.Notification) error {
	logger.Info("Processing order status change notification",
		"order_id", notification.ResourceID,
		"user_id", notification.UserID)

	// Get user using the global service provider
	user, err := services.GetServices().User.GetUserByID(notification.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", notification.UserID, err)
	}

	// Trigger order sync using the global service provider
	_, err = services.GetServices().Orders.SyncOrdersFromBrickLink(user)
	if err != nil {
		return fmt.Errorf("failed to sync orders after status change for user %d: %w", notification.UserID, err)
	}

	logger.Info("Order status change notification processed - order sync triggered", "order_id", notification.ResourceID, "user_id", notification.UserID)
	return nil
}

// processOrderItemsNotification handles order items change notifications
func (c *BricklinkNotificationsClient) processOrderItemsNotification(notification models.Notification) error {
	logger.Info("Processing order items change notification",
		"order_id", notification.ResourceID,
		"user_id", notification.UserID)

	// Get user using the global service provider
	user, err := services.GetServices().User.GetUserByID(notification.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", notification.UserID, err)
	}

	// Trigger order sync using the global service provider
	_, err = services.GetServices().Orders.SyncOrdersFromBrickLink(user)
	if err != nil {
		return fmt.Errorf("failed to sync orders after items change for user %d: %w", notification.UserID, err)
	}

	logger.Info("Order items change notification processed - order sync triggered", "order_id", notification.ResourceID, "user_id", notification.UserID)
	return nil
}

// processNewMessageNotification handles new message notifications
func (c *BricklinkNotificationsClient) processNewMessageNotification(notification models.Notification) error {
	logger.Info("Processing new message notification", 
		"message_id", notification.ResourceID,
		"user_id", notification.UserID)
	
	// TODO: Implement message handling
	// This would:
	// 1. Fetch the message details from BrickLink API
	// 2. Store the message in our database
	// 3. Trigger user notifications
	
	logger.Info("New message notification processed (placeholder)", "message_id", notification.ResourceID)
	return nil
}

// processNewFeedbackNotification handles new feedback notifications
func (c *BricklinkNotificationsClient) processNewFeedbackNotification(notification models.Notification) error {
	logger.Info("Processing new feedback notification", 
		"feedback_id", notification.ResourceID,
		"user_id", notification.UserID)
	
	// TODO: Implement feedback handling
	// This would:
	// 1. Fetch the feedback details from BrickLink API
	// 2. Store the feedback in our database
	// 3. Update seller metrics or trigger responses
	
	logger.Info("New feedback notification processed (placeholder)", "feedback_id", notification.ResourceID)
	return nil
}