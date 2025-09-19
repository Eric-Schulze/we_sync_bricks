package models

import (
	"time"
)

type User struct {
	ID        int64      `json:"id" db:"id"`
	Username  string     `json:"username" db:"username"`
	FirstName *string    `json:"first_name" db:"first_name"`
	LastName  *string    `json:"last_name" db:"last_name"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

// UserUpdateRequest represents the user profile update form data
type UserUpdateRequest struct {
	FirstName string `json:"first_name" form:"first_name"`
	LastName  string `json:"last_name" form:"last_name"`
	Email     string `json:"email" form:"email"`
	Username  string `json:"username" form:"username"`
}

// PasswordChangeRequest represents the password change form data
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" form:"current_password"`
	NewPassword     string `json:"new_password" form:"new_password"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
}

// APIKeyUpdateRequest represents the API key update form data
type APIKeyUpdateRequest struct {
	Provider       string `json:"provider" form:"provider"`
	ConsumerKey    string `json:"consumer_key" form:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret" form:"consumer_secret"`
	Token          string `json:"token" form:"token"`
	TokenSecret    string `json:"token_secret" form:"token_secret"`
}

// UserResponse represents the response data for user operations
type UserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UserOAuthCredential represents OAuth credentials for external APIs
type UserOAuthCredential struct {
	ID             int64     `json:"id" db:"id"`
	UserID         int64     `json:"user_id" db:"user_id"`
	Provider       string    `json:"provider" db:"provider"`
	ConsumerKey    string    `json:"consumer_key" db:"consumer_key"`
	ConsumerSecret string    `json:"-" db:"consumer_secret"` // Hide in JSON for security
	Token          string    `json:"token" db:"token"`
	TokenSecret    string    `json:"-" db:"token_secret"` // Hide in JSON for security
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// UserProfileData represents complete user profile data including OAuth credentials
type UserProfileData struct {
	User                 *User                `json:"user"`
	BricklinkCredentials *UserOAuthCredential `json:"bricklink_credentials,omitempty"`
	BrickowlCredentials  *UserOAuthCredential `json:"brickowl_credentials,omitempty"`
}
