package bricklink

import (
	"encoding/json"
	"errors"

	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

// BricklinkItemResponse represents the structure of a Bricklink API item response
type BricklinkItemResponse struct {
	No         string `json:"no"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	CategoryID int    `json:"category_id"`
}

// ParseBricklinkItem parses a Bricklink API JSON response into an Item struct
func ParseBricklinkItem(jsonResponse string) (*Item, error) {
	logger.Debug("Parsing Bricklink item response", "response_length", len(jsonResponse))

	// Parse the full BL response first
	var blResponse BLResponse
	err := json.Unmarshal([]byte(jsonResponse), &blResponse)
	if err != nil {
		logger.Error("Failed to parse Bricklink response", "error", err)
		return nil, errors.New("failed to parse API response")
	}

	// Check if response was successful
	if blResponse.Meta.Code != 200 {
		logger.Error("Bricklink API returned error", "code", blResponse.Meta.Code, "message", blResponse.Meta.Message)
		return nil, errors.New("API error: " + blResponse.Meta.Message)
	}

	// Extract the data portion and parse it as BricklinkItemResponse
	dataBytes, err := json.Marshal(blResponse.Data)
	if err != nil {
		logger.Error("Failed to marshal Bricklink data", "error", err)
		return nil, errors.New("failed to process API data")
	}

	var itemResponse BricklinkItemResponse
	err = json.Unmarshal(dataBytes, &itemResponse)
	if err != nil {
		logger.Error("Failed to parse Bricklink item data", "error", err)
		return nil, errors.New("failed to parse item data")
	}

	// Convert to our Item struct
	item := &Item{
		No:         itemResponse.No,
		Name:       itemResponse.Name,
		Type:       itemResponse.Type,
		CategoryID: itemResponse.CategoryID,
	}

	logger.Info("Successfully parsed Bricklink item", "item_no", item.No, "item_name", item.Name, "item_type", item.Type)
	return item, nil
}
