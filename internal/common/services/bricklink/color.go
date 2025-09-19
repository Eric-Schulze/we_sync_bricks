package bricklink

import (
	"encoding/json"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

// BricklinkColorResponse represents the API response structure
type BricklinkColorResponse struct {
	Meta BLResponseMeta   `json:"meta"`
	Data []BricklinkColor `json:"data"`
}

// BricklinkColor represents a single color from BrickLink API
type BricklinkColor struct {
	ColorID   int    `json:"color_id"`
	ColorName string `json:"color_name"`
	ColorCode string `json:"color_code"`
	ColorType string `json:"color_type"`
}

// Color represents the colors table structure in our database
type Color struct {
	ID            int64      `db:"id"`
	BricklinkID   int        `db:"bricklink_id"`
	RebrickableID *int       `db:"rebrickable_id"`
	Name          string     `db:"name"`
	Code          *string    `db:"code"`
	Type          *string    `db:"type"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
}

// ToDatabase converts BricklinkColor to Color format
func (bc BricklinkColor) ToDatabase() Color {
	var code, colorType *string
	if bc.ColorCode != "" {
		code = &bc.ColorCode
	}
	if bc.ColorType != "" {
		colorType = &bc.ColorType
	}

	return Color{
		BricklinkID:   bc.ColorID,
		RebrickableID: nil, // No rebrickable mapping available from BrickLink API
		Name:          bc.ColorName,
		Code:          code,
		Type:          colorType,
		CreatedAt:     time.Now(),
	}
}

// ParseColorsResponse parses BrickLink colors API response
func ParseColorsResponse(responseStr string) ([]BricklinkColor, error) {
	logger.Debug("Parsing BrickLink colors response", "response_length", len(responseStr))

	var response BricklinkColorResponse
	err := json.Unmarshal([]byte(responseStr), &response)
	if err != nil {
		logger.Error("Failed to parse colors response JSON", "error", err)
		return nil, err
	}

	if response.Meta.Code != 200 {
		logger.Error("BrickLink colors API returned error", "code", response.Meta.Code, "message", response.Meta.Message)
		return nil, err
	}

	logger.Debug("Successfully parsed colors response", "color_count", len(response.Data))
	return response.Data, nil
}
