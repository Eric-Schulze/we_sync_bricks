package profile

import (
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

type ProfileService struct {
	repo *ProfileRepository
}

// NewProfileService creates a new profile service
func NewProfileService(repo *ProfileRepository) *ProfileService {
	return &ProfileService{
		repo: repo,
	}
}

// GetProfile retrieves a user's profile
func (s *ProfileService) GetProfile(userID int64) (*models.User, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.repo.GetUserProfile(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve profile")
	}

	return user, nil
}

// UpdateProfile updates a user's profile
func (s *ProfileService) UpdateProfile(userID int64, req *ProfileUpdateRequest) (*models.User, error) {
	logger.Debug("Service: Updating user profile", "user_id", userID, "username", req.Username)

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

	user, err := s.repo.UpdateUserProfile(userID, req)
	if err != nil {
		logger.Error("Service: Failed to update profile", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	logger.Info("Service: Successfully updated user profile", "user_id", userID, "username", req.Username)
	return user, nil
}

// ChangePassword changes a user's password
func (s *ProfileService) ChangePassword(userID int64, req *PasswordChangeRequest) error {
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
func (s *ProfileService) GetProfileWithCredentials(userID int64) (*ProfileData, error) {
	logger.Debug("Service: Getting profile with credentials", "user_id", userID)

	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Get user profile
	user, err := s.repo.GetUserProfile(userID)
	if err != nil {
		logger.Error("Service: Failed to get user profile", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to retrieve profile: %w", err)
	}

	// Convert models.User to profile.User
	profileUser := &User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if user.FirstName != nil {
		profileUser.FirstName = *user.FirstName
	}
	if user.LastName != nil {
		profileUser.LastName = *user.LastName
	}
	if user.UpdatedAt != nil {
		profileUser.UpdatedAt = user.UpdatedAt.Format("2006-01-02 15:04:05")
	}

	// Get OAuth credentials
	bricklinkCreds, _ := s.repo.GetUserOAuthCredentials(userID, "bricklink")
	brickowlCreds, _ := s.repo.GetUserOAuthCredentials(userID, "brickowl")

	profileData := &ProfileData{
		User:                 profileUser,
		BricklinkCredentials: bricklinkCreds,
		BrickowlCredentials:  brickowlCreds,
	}

	logger.Debug("Service: Successfully retrieved profile with credentials", "user_id", userID)
	return profileData, nil
}

// UpdateAPICredentials saves or updates OAuth credentials for a provider
func (s *ProfileService) UpdateAPICredentials(userID int64, req *APIKeyUpdateRequest) error {
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
func (s *ProfileService) DeleteAPICredentials(userID int64, provider string) error {
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

// InitializeProfileHandler creates a fully initialized profile handler with all dependencies
func InitializeProfileHandler(dbService models.DBService, templates *template.Template, jwtSecret []byte) *ProfileHandler {
	repo := NewProfileRepository(dbService)
	service := NewProfileService(repo)
	handler := NewProfileHandler(service, templates, jwtSecret)
	return handler
}
