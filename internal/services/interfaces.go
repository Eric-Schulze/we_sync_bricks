package services

import (
	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
)

// UserService defines the interface for user operations
type UserService interface {
	GetUserByID(userID int64) (*models.User, error)
	GetUser(userID int64) (*models.User, error)
	GetProfile(userID int64) (*models.User, error)
	UpdateUser(userID int64, req *models.UserUpdateRequest) (*models.User, error)
	UpdateProfile(userID int64, req *models.UserUpdateRequest) (*models.User, error)
	ChangePassword(userID int64, req *models.PasswordChangeRequest) error
	GetProfileWithCredentials(userID int64) (*models.UserProfileData, error)
	UpdateAPICredentials(userID int64, req *models.APIKeyUpdateRequest) error
	DeleteAPICredentials(userID int64, provider string) error
	GetUserOAuthCredentials(userID int64, provider string) (*models.UserOAuthCredential, error)
}

// OrdersService defines the interface for order operations
type OrdersService interface {
	GetOrders(userID int64) ([]models.Order, error)
	GetFilteredOrders(userID int64, filters models.OrderFilters) ([]models.Order, int64, error)
	GetOrdersCount(userID int64) (int64, error)
	SyncOrdersFromBrickLink(user *models.User) (models.OrderSync, error)
	SyncOrdersFromBrickOwl(user *models.User) (models.OrderSync, error) // Future implementation
}

// NotificationService defines the interface for notification operations
type NotificationService interface {
	ProcessNotificationsForUser(userID int) error
	RegisterClient(client interface{}) // NotificationClient interface
	ProcessWebhookForUser(userID int, provider string) error
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	Login(login, password string) (interface{}, error)
	ValidateToken(tokenString string) (*models.User, error)
	// Add other auth methods as needed
}