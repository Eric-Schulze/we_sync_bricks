package user

import (
	"errors"
	"fmt"
	"strings"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

type UserService struct {
	repo *UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo *UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(userID int64) (*models.User, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve user")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID (alias for GetUser)
func (s *UserService) GetUserByID(userID int64) (*models.User, error) {
	return s.GetUser(userID)
}

// GetProfile retrieves a user's profile (alias for GetUser for backward compatibility)
func (s *UserService) GetProfile(userID int64) (*models.User, error) {
	return s.GetUser(userID)
}

// UpdateUser updates a user's profile
func (s *UserService) UpdateUser(userID int64, req *models.UserUpdateRequest) (*models.User, error) {
	logger.Debug("Service: Updating user", "user_id", userID, "username", req.Username)

	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	if req.Email == "" || req.Username == "" {
		return nil, errors.New("email and username are required")
	}

	// Validate email format
	if !strings.Contains(req.Email, "@") {
		return nil, errors.New("invalid email format")
	}

	user, err := s.repo.UpdateUser(userID, req)
	if err != nil {
		logger.Error("Service: Failed to update user", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("Service: Successfully updated user", "user_id", userID, "username", req.Username)
	return user, nil
}

// UpdateProfile updates a user's profile (alias for UpdateUser for backward compatibility)
func (s *UserService) UpdateProfile(userID int64, req *models.UserUpdateRequest) (*models.User, error) {
	return s.UpdateUser(userID, req)
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(userID int64, req *models.PasswordChangeRequest) error {
	logger.Debug("Service: Changing password", "user_id", userID)

	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	// Validate request
	if req.CurrentPassword == "" || req.NewPassword == "" || req.ConfirmPassword == "" {
		return errors.New("all password fields are required")
	}

	if req.NewPassword != req.ConfirmPassword {
		return errors.New("new password and confirmation do not match")
	}

	if len(req.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters long")
	}

	err := s.repo.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		logger.Error("Service: Failed to change password", "user_id", userID, "error", err)
		return err
	}

	logger.Info("Service: Successfully changed password", "user_id", userID)
	return nil
}

// GetProfileWithCredentials retrieves a user's complete profile including OAuth credentials
func (s *UserService) GetProfileWithCredentials(userID int64) (*models.UserProfileData, error) {
	logger.Debug("Service: Getting profile with credentials", "user_id", userID)

	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Get user profile
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		logger.Error("Service: Failed to get user", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	// Get OAuth credentials
	bricklinkCreds, _ := s.repo.GetUserOAuthCredentials(userID, "bricklink")
	brickowlCreds, _ := s.repo.GetUserOAuthCredentials(userID, "brickowl")

	profileData := &models.UserProfileData{
		User:                 user,
		BricklinkCredentials: bricklinkCreds,
		BrickowlCredentials:  brickowlCreds,
	}

	logger.Debug("Service: Successfully retrieved profile with credentials", "user_id", userID)
	return profileData, nil
}

// UpdateAPICredentials saves or updates OAuth credentials for a provider
func (s *UserService) UpdateAPICredentials(userID int64, req *models.APIKeyUpdateRequest) error {
	logger.Debug("Service: Updating API credentials", "user_id", userID, "provider", req.Provider)

	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	// Validate provider
	if req.Provider != "bricklink" && req.Provider != "brickowl" {
		return errors.New("provider must be 'bricklink' or 'brickowl'")
	}

	// Validate required fields
	if req.ConsumerKey == "" || req.ConsumerSecret == "" || req.Token == "" || req.TokenSecret == "" {
		return errors.New("all API credential fields are required")
	}

	err := s.repo.UpsertUserOAuthCredentials(userID, req)
	if err != nil {
		logger.Error("Service: Failed to update API credentials", "user_id", userID, "provider", req.Provider, "error", err)
		return err
	}

	logger.Info("Service: Successfully updated API credentials", "user_id", userID, "provider", req.Provider)
	return nil
}

// DeleteAPICredentials removes OAuth credentials for a provider
func (s *UserService) DeleteAPICredentials(userID int64, provider string) error {
	logger.Debug("Service: Deleting API credentials", "user_id", userID, "provider", provider)

	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	if provider != "bricklink" && provider != "brickowl" {
		return errors.New("provider must be 'bricklink' or 'brickowl'")
	}

	err := s.repo.DeleteUserOAuthCredentials(userID, provider)
	if err != nil {
		logger.Error("Service: Failed to delete API credentials", "user_id", userID, "provider", provider, "error", err)
		return err
	}

	logger.Info("Service: Successfully deleted API credentials", "user_id", userID, "provider", provider)
	return nil
}

// GetUserOAuthCredentials retrieves OAuth credentials for a user and provider
func (s *UserService) GetUserOAuthCredentials(userID int64, provider string) (*models.UserOAuthCredential, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return s.repo.GetUserOAuthCredentials(userID, provider)
}