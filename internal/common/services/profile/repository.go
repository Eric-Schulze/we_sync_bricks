package profile

import (
	"crypto/md5"
	"fmt"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/db"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/jackc/pgx/v5"
)

type ProfileRepository struct {
	db models.DBService
}

// NewProfileRepository creates a new profile repository
func NewProfileRepository(db models.DBService) *ProfileRepository {
	return &ProfileRepository{
		db: db,
	}
}

// GetUserProfile retrieves a user's profile by user ID
func (r *ProfileRepository) GetUserProfile(userID int64) (*models.User, error) {
	logger.Debug("Getting user profile", "user_id", userID)

	sql := `SELECT id, username, first_name, last_name, email, created_at, updated_at 
			FROM users WHERE id = $1`

	rows, err := r.db.Query(sql, userID)
	if err != nil {
		logger.Error("Failed to query user profile", "user_id", userID, "error", err)
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		logger.Error("User not found", "user_id", userID)
		return nil, fmt.Errorf("user not found")
	}

	var user models.User
	err = rows.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		logger.Error("Failed to scan user profile", "user_id", userID, "error", err)
		return nil, err
	}

	logger.Debug("Successfully retrieved user profile", "user_id", userID, "username", user.Username)
	return &user, nil
}

// UpdateUserProfile updates a user's profile information
func (r *ProfileRepository) UpdateUserProfile(userID int64, req *ProfileUpdateRequest) (*models.User, error) {
	logger.Debug("Updating user profile", "user_id", userID, "username", req.Username)

	sql := `UPDATE users 
			SET first_name = $1, last_name = $2, email = $3, username = $4, updated_at = CURRENT_TIMESTAMP
			WHERE id = $5 
			RETURNING id, username, first_name, last_name, email, created_at, updated_at`

	var firstName, lastName *string
	if req.FirstName != "" {
		firstName = &req.FirstName
	}
	if req.LastName != "" {
		lastName = &req.LastName
	}

	rows, err := r.db.Query(sql, firstName, lastName, req.Email, req.Username, userID)
	if err != nil {
		logger.Error("Failed to update user profile", "user_id", userID, "error", err)
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		logger.Error("User not found for update", "user_id", userID)
		return nil, fmt.Errorf("user not found")
	}

	var user models.User
	err = rows.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		logger.Error("Failed to scan updated user profile", "user_id", userID, "error", err)
		return nil, err
	}

	logger.Info("Successfully updated user profile", "user_id", userID, "username", req.Username)
	return &user, nil
}

// ChangePassword updates a user's password after verifying the current password
func (r *ProfileRepository) ChangePassword(userID int64, currentPassword, newPassword string) error {
	logger.Debug("Changing password for user", "user_id", userID)

	// First verify the current password
	sql := `SELECT password FROM users WHERE id = $1`
	storedPasswordHash, err := db.QueryRowToStructFromService[string](r.db, sql, userID)
	if err != nil {
		logger.Error("Failed to retrieve current password", "user_id", userID, "error", err)
		return fmt.Errorf("failed to verify current password")
	}

	// For now using simple MD5 (in production use bcrypt)
	currentPasswordHash := fmt.Sprintf("%x", md5.Sum([]byte(currentPassword)))
	if currentPasswordHash != storedPasswordHash {
		logger.Warn("Invalid current password provided", "user_id", userID)
		return fmt.Errorf("current password is incorrect")
	}

	// Update with new password
	newPasswordHash := fmt.Sprintf("%x", md5.Sum([]byte(newPassword)))
	updateSQL := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	_, err = r.db.ExecSQL(updateSQL, newPasswordHash, userID)
	if err != nil {
		logger.Error("Failed to update password", "user_id", userID, "error", err)
		return fmt.Errorf("failed to update password")
	}

	logger.Info("Successfully changed password", "user_id", userID)
	return nil
}

// GetUserOAuthCredentials retrieves OAuth credentials for a user and provider
func (r *ProfileRepository) GetUserOAuthCredentials(userID int64, provider string) (*UserOAuthCredential, error) {
	logger.Debug("Getting OAuth credentials", "user_id", userID, "provider", provider)

	sql := `SELECT id, user_id, provider, consumer_key, consumer_secret, token, token_secret, is_active
			FROM user_oauth_credentials 
			WHERE user_id = $1 AND provider = $2 AND is_active = true`

	credential, err := db.QueryRowToStructFromService[UserOAuthCredential](r.db, sql, userID, provider)
	if err != nil {
		if err == pgx.ErrNoRows {
			// It's OK if no credentials exist yet
			logger.Debug("No OAuth credentials found", "user_id", userID, "provider", provider)
			return nil, nil
		}
		// Log actual errors
		logger.Error("Error retrieving OAuth credentials", "user_id", userID, "provider", provider, "error", err)
		return nil, err
	}

	logger.Debug("Successfully retrieved OAuth credentials", "user_id", userID, "provider", provider)
	return &credential, nil
}

// UpsertUserOAuthCredentials creates or updates OAuth credentials for a user
func (r *ProfileRepository) UpsertUserOAuthCredentials(userID int64, req *APIKeyUpdateRequest) error {
	logger.Debug("Upserting OAuth credentials", "user_id", userID, "provider", req.Provider)

	sql := `INSERT INTO user_oauth_credentials 
			(user_id, provider, consumer_key, consumer_secret, token, token_secret, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (user_id, provider) 
			DO UPDATE SET 
				consumer_key = EXCLUDED.consumer_key,
				consumer_secret = EXCLUDED.consumer_secret,
				token = EXCLUDED.token,
				token_secret = EXCLUDED.token_secret,
				is_active = true,
				updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.ExecSQL(sql, userID, req.Provider, req.ConsumerKey, req.ConsumerSecret, req.Token, req.TokenSecret)
	if err != nil {
		logger.Error("Failed to upsert OAuth credentials", "user_id", userID, "provider", req.Provider, "error", err)
		return fmt.Errorf("failed to save API credentials")
	}

	logger.Info("Successfully saved OAuth credentials", "user_id", userID, "provider", req.Provider)
	return nil
}

// DeleteUserOAuthCredentials removes OAuth credentials for a user and provider
func (r *ProfileRepository) DeleteUserOAuthCredentials(userID int64, provider string) error {
	logger.Debug("Deleting OAuth credentials", "user_id", userID, "provider", provider)

	sql := `UPDATE user_oauth_credentials 
			SET is_active = false, updated_at = CURRENT_TIMESTAMP
			WHERE user_id = $1 AND provider = $2`

	_, err := r.db.ExecSQL(sql, userID, provider)
	if err != nil {
		logger.Error("Failed to delete OAuth credentials", "user_id", userID, "provider", provider, "error", err)
		return fmt.Errorf("failed to delete API credentials")
	}

	logger.Info("Successfully deleted OAuth credentials", "user_id", userID, "provider", provider)
	return nil
}
