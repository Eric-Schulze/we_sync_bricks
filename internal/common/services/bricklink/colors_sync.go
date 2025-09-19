package bricklink

import (
	"fmt"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

// ColorsSyncService handles syncing colors from BrickLink API to database
type ColorsSyncService struct {
	clientManager *ClientManager
	dbService     models.DBService
}

// NewColorsSyncService creates a new colors sync service
func NewColorsSyncService(clientManager *ClientManager, dbService models.DBService) *ColorsSyncService {
	return &ColorsSyncService{
		clientManager: clientManager,
		dbService:     dbService,
	}
}

// SyncColors syncs all colors from BrickLink API to database
func (css *ColorsSyncService) SyncColors(user *models.User) error {
	logger.Info("Starting BrickLink colors sync", "user_id", user.ID)

	// Get BrickLink client for user
	client, err := css.clientManager.GetClient(user)
	if err != nil {
		logger.Error("Failed to get BrickLink client for colors sync", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to get BrickLink client: %w", err)
	}

	// Fetch colors from BrickLink API
	logger.Debug("Fetching colors from BrickLink API", "user_id", user.ID)
	responseStr, err := client.GetColors()
	if err != nil {
		logger.Error("Failed to fetch colors from BrickLink API", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to fetch colors from BrickLink: %w", err)
	}

	// Parse the response
	colors, err := ParseColorsResponse(responseStr)
	if err != nil {
		logger.Error("Failed to parse colors response", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to parse colors response: %w", err)
	}

	logger.Info("Fetched colors from BrickLink API", "user_id", user.ID, "color_count", len(colors))

	// Sync colors to database
	syncedCount, err := css.syncColorsToDatabase(colors)
	if err != nil {
		logger.Error("Failed to sync colors to database", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to sync colors to database: %w", err)
	}

	logger.Info("Successfully synced BrickLink colors", "user_id", user.ID, "total_colors", len(colors), "synced_count", syncedCount)
	return nil
}

// syncColorsToDatabase syncs the parsed colors to the database
func (css *ColorsSyncService) syncColorsToDatabase(colors []BricklinkColor) (int, error) {
	logger.Debug("Starting database sync for colors", "color_count", len(colors))

	syncedCount := 0

	for _, color := range colors {
		// Convert to database format
		dbColor := color.ToDatabase()

		// Check if color already exists
		existsSQL := `SELECT id FROM colors WHERE bricklink_id = $1`
		rows, err := css.dbService.CollectRowsToMap(existsSQL, color.ColorID)
		if err != nil {
			logger.Error("Failed to check if color exists", "color_id", color.ColorID, "error", err)
			continue
		}

		if len(rows) > 0 {
			// Color exists, update it
			updateSQL := `UPDATE colors 
						  SET name = $2, code = $3, type = $4, updated_at = CURRENT_TIMESTAMP 
						  WHERE bricklink_id = $1`
			
			_, err = css.dbService.ExecSQL(updateSQL, dbColor.BricklinkID, dbColor.Name, dbColor.Code, dbColor.Type)
			if err != nil {
				logger.Error("Failed to update existing color", "color_id", color.ColorID, "color_name", color.ColorName, "error", err)
				continue
			}

			logger.Debug("Updated existing color", "color_id", color.ColorID, "color_name", color.ColorName)
		} else {
			// Color doesn't exist, insert it
			insertSQL := `INSERT INTO colors (bricklink_id, rebrickable_id, name, code, type, created_at) 
						  VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`

			_, err = css.dbService.ExecSQL(insertSQL, dbColor.BricklinkID, dbColor.RebrickableID, dbColor.Name, dbColor.Code, dbColor.Type)
			if err != nil {
				logger.Error("Failed to insert new color", "color_id", color.ColorID, "color_name", color.ColorName, "error", err)
				continue
			}

			logger.Debug("Inserted new color", "color_id", color.ColorID, "color_name", color.ColorName)
		}

		syncedCount++
	}

	logger.Debug("Completed database sync for colors", "synced_count", syncedCount, "total_count", len(colors))
	return syncedCount, nil
}

// GetColorCount returns the number of colors in the database
func (css *ColorsSyncService) GetColorCount() (int, error) {
	countSQL := `SELECT COUNT(*) as count FROM colors`
	rows, err := css.dbService.CollectRowsToMap(countSQL)
	if err != nil {
		logger.Error("Failed to get color count", "error", err)
		return 0, err
	}

	if len(rows) > 0 {
		if count, ok := rows[0]["count"]; ok {
			if countInt, ok := count.(int64); ok {
				return int(countInt), nil
			}
		}
	}

	return 0, nil
}

// GetLastSyncInfo returns information about the last sync
func (css *ColorsSyncService) GetLastSyncInfo() (map[string]interface{}, error) {
	// Get total count
	count, err := css.GetColorCount()
	if err != nil {
		return nil, err
	}

	// Get last updated timestamp
	lastUpdatedSQL := `SELECT MAX(updated_at) as last_updated FROM colors WHERE updated_at IS NOT NULL`
	rows, err := css.dbService.CollectRowsToMap(lastUpdatedSQL)
	if err != nil {
		logger.Error("Failed to get last sync info", "error", err)
		return nil, err
	}

	result := map[string]interface{}{
		"total_colors": count,
		"last_updated": nil,
	}

	if len(rows) > 0 && rows[0]["last_updated"] != nil {
		result["last_updated"] = rows[0]["last_updated"]
	}

	return result, nil
}