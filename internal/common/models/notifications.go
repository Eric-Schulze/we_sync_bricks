package models

import "time"

// NotificationType represents different types of notifications
type NotificationType string

const (
	TypeOrderNew           NotificationType = "ORDER_NEW"
	TypeOrderStatusChanged NotificationType = "ORDER_STATUS_CHANGED"
	TypeOrderItemsChanged  NotificationType = "ORDER_ITEMS_CHANGED"
	TypeMessageNew         NotificationType = "MESSAGE_NEW"
	TypeFeedbackNew        NotificationType = "FEEDBACK_NEW"
)

// Notification represents a generic notification from any provider
type Notification struct {
	ID           string           `json:"id"`
	Provider     string           `json:"provider"`
	Type         NotificationType `json:"type"`
	ResourceType string           `json:"resource_type"`
	ResourceID   string           `json:"resource_id"`
	Message      string           `json:"message"`
	DateCreated  time.Time        `json:"date_created"`
	UserID       int64            `json:"user_id"`
}