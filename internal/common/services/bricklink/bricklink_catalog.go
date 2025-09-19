package bricklink

import (
	"errors"
	"strconv"
	"strings"

	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/eric-schulze/we_sync_bricks/utils/oauth"
)

const BricklinkCatalogEndpoint = "/items"

type BricklinkCatalogClient struct {
	BLClient
}

func NewBricklinkCatalogClient() BricklinkCatalogClient {
	// Create a new OAuth client for the Bricklink API
	client := oauth.NewOAuthClient(ConsumerKey, ConsumerSecret, Token, TokenSecret)

	return BricklinkCatalogClient{
		BLClient{
			oAuthClient: client,
		},
	}
}

func (blClient BricklinkCatalogClient) GetItem(item_id string) (*Item, error) {
	logger.Info("Bricklink Client: Starting GetItem request", "item_id", item_id)

	// Always use MINIFIG type since we're specifically dealing with minifigures
	endpoint := BricklinkCatalogEndpoint + "/MINIFIG/" + item_id
	logger.Info("Bricklink Client: Making API request", "item_id", item_id, "endpoint", endpoint)

	response, err := blClient.BLGet(endpoint)
	if err != nil {
		logger.Error("Bricklink Client: API request failed", "item_id", item_id, "endpoint", endpoint, "error", err)
		return nil, err
	}

	logger.Info("Bricklink Client: API request completed", "item_id", item_id, "response_code", response.Meta.Code, "response_message", response.Meta.Message)

	if response.Meta.Code != 200 {
		logger.Error("Bricklink Client: API returned error response", "item_id", item_id, "code", response.Meta.Code, "message", response.Meta.Message)

		// Check for specific error cases and provide user-friendly messages
		if response.Meta.Code == 400 && (response.Meta.Message == "PARAMETER_MISSING_OR_INVALID" ||
			response.Meta.Description != "") {
			return nil, errors.New("The ID '" + item_id + "' is not a valid minifigure ID. Please check the ID and try again.")
		}

		return nil, errors.New("Unable to find minifigure with ID '" + item_id + "'. Please verify the ID is correct.")
	}

	logger.Info("Bricklink Client: Successfully retrieved raw item data", "item_id", item_id, "response_length", len(response.responseStr))

	// Parse the response using our parsing function
	item, err := ParseBricklinkItem(response.responseStr)
	if err != nil {
		logger.Error("Bricklink Client: Failed to parse item response", "item_id", item_id, "error", err)
		return nil, err
	}

	logger.Info("Bricklink Client: Successfully parsed item", "item_id", item_id, "item_name", item.Name, "item_type", item.Type)
	return item, nil
}

func (blClient BricklinkCatalogClient) GetSubset(item_id string, subset_type string) (string, error) {
	response, err := blClient.BLGet(BricklinkCatalogEndpoint + "/" + subset_type + "/" + item_id + "subset")
	if err != nil {
		return "", err
	}

	if response.Meta.Code != 200 {
		logger.Error("Failed to get subset of item from Bricklink", "item", item_id, "subset_type", subset_type, "code", response.Meta.Code, "message", response.Meta.Message)
		return "", errors.New("Failed to get subset of " + subset_type + " from Bricklink: " + strconv.Itoa(response.Meta.Code) + " " + response.Meta.Message)
	}

	return response.responseStr, nil
}

// GetMinifigPictures retrieves image URLs for a minifigure
func (blClient BricklinkCatalogClient) GetMinifigPictures(item_id string) (string, error) {
	logger.Info("Bricklink Client: Starting GetMinifigPictures request", "item_id", item_id)

	endpoint := BricklinkCatalogEndpoint + "/MINIFIG/" + item_id + "/images/0"
	logger.Info("Bricklink Client: Making API request", "item_id", item_id, "endpoint", endpoint)

	response, err := blClient.BLGet(endpoint)
	if err != nil {
		logger.Error("Bricklink Client: Picture API request failed", "item_id", item_id, "endpoint", endpoint, "error", err)
		return "", err
	}

	logger.Info("Bricklink Client: Picture API request completed", "item_id", item_id, "response_code", response.Meta.Code, "response_message", response.Meta.Message)

	if response.Meta.Code != 200 {
		logger.Error("Bricklink Client: Picture API returned error response", "item_id", item_id, "code", response.Meta.Code, "message", response.Meta.Message, "description", response.Meta.Description)
		return "", errors.New("Failed to get pictures for minifigure " + item_id + ": " + strconv.Itoa(response.Meta.Code) + " " + response.Meta.Message)
	}

	logger.Info("Bricklink Client: Successfully retrieved minifig pictures", "item_id", item_id, "response_length", len(response.responseStr), "response_content", response.responseStr)
	return response.responseStr, nil
}

// GetPriceGuide retrieves price guide data for a minifigure (legacy method for backward compatibility)
func (blClient BricklinkCatalogClient) GetPriceGuide(item_id string, condition string, color_id int) (string, error) {
	return blClient.GetPriceGuideByType("MINIFIG", item_id, condition, color_id)
}

// GetPriceGuideByType retrieves price guide data for any item type (MINIFIG, PART, etc.)
func (blClient BricklinkCatalogClient) GetPriceGuideByType(item_type string, item_id string, condition string, color_id int) (string, error) {
	logger.Info("Bricklink Client: Starting GetPriceGuideByType request", "item_type", item_type, "item_id", item_id, "condition", condition, "color_id", color_id)

	endpoint := BricklinkCatalogEndpoint + "/" + item_type + "/" + item_id + "/price"

	// Build query parameters
	var params []string
	if condition != "" {
		params = append(params, "new_or_used="+condition)
	}
	if color_id > 0 {
		params = append(params, "color_id="+strconv.Itoa(color_id))
	}
	// Add guide_type=sold to get total_quantity data
	params = append(params, "guide_type=sold")

	if len(params) > 0 {
		endpoint += "?" + strings.Join(params, "&")
	}

	logger.Info("Bricklink Client: Making API request", "item_type", item_type, "item_id", item_id, "condition", condition, "endpoint", endpoint)

	response, err := blClient.BLGet(endpoint)
	if err != nil {
		logger.Error("Bricklink Client: Price guide API request failed", "item_type", item_type, "item_id", item_id, "endpoint", endpoint, "error", err)
		return "", err
	}

	logger.Info("Bricklink Client: Price guide API request completed", "item_type", item_type, "item_id", item_id, "response_code", response.Meta.Code, "response_message", response.Meta.Message)

	if response.Meta.Code != 200 {
		logger.Error("Bricklink Client: Price guide API returned error response", "item_type", item_type, "item_id", item_id, "code", response.Meta.Code, "message", response.Meta.Message)
		return "", errors.New("Failed to get price guide for " + item_type + " " + item_id + ": " + strconv.Itoa(response.Meta.Code) + " " + response.Meta.Message)
	}

	logger.Info("Bricklink Client: Successfully retrieved price guide", "item_type", item_type, "item_id", item_id, "response_length", len(response.responseStr))
	return response.responseStr, nil
}

// GetSubsetParts retrieves the parts that make up a minifigure
func (blClient BricklinkCatalogClient) GetSubsetParts(item_id string) (string, error) {
	logger.Info("Bricklink Client: Starting GetSubsetParts request", "item_id", item_id)

	endpoint := BricklinkCatalogEndpoint + "/MINIFIG/" + item_id + "/subsets"
	logger.Info("Bricklink Client: Making API request", "item_id", item_id, "endpoint", endpoint)

	response, err := blClient.BLGet(endpoint)
	if err != nil {
		logger.Error("Bricklink Client: Subset parts API request failed", "item_id", item_id, "endpoint", endpoint, "error", err)
		return "", err
	}

	logger.Info("Bricklink Client: Subset parts API request completed", "item_id", item_id, "response_code", response.Meta.Code, "response_message", response.Meta.Message)

	if response.Meta.Code != 200 {
		logger.Error("Bricklink Client: Subset parts API returned error response", "item_id", item_id, "code", response.Meta.Code, "message", response.Meta.Message)
		return "", errors.New("Failed to get subset parts for minifigure " + item_id + ": " + strconv.Itoa(response.Meta.Code) + " " + response.Meta.Message)
	}

	logger.Info("Bricklink Client: Successfully retrieved subset parts", "item_id", item_id, "response_length", len(response.responseStr))
	return response.responseStr, nil
}

// GetPartPictures retrieves image URLs for a specific part with color
func (blClient BricklinkCatalogClient) GetPartPictures(item_type string, item_id string, color_id int) (string, error) {
	logger.Info("Bricklink Client: Starting GetPartPictures request", "item_type", item_type, "item_id", item_id, "color_id", color_id)

	endpoint := BricklinkCatalogEndpoint + "/" + item_type + "/" + item_id + "/images/" + strconv.Itoa(color_id)
	logger.Info("Bricklink Client: Making API request", "item_type", item_type, "item_id", item_id, "color_id", color_id, "endpoint", endpoint)

	response, err := blClient.BLGet(endpoint)
	if err != nil {
		logger.Error("Bricklink Client: Part picture API request failed", "item_type", item_type, "item_id", item_id, "color_id", color_id, "endpoint", endpoint, "error", err)
		return "", err
	}

	logger.Info("Bricklink Client: Part picture API request completed", "item_type", item_type, "item_id", item_id, "color_id", color_id, "response_code", response.Meta.Code, "response_message", response.Meta.Message)

	if response.Meta.Code != 200 {
		logger.Error("Bricklink Client: Part picture API returned error response", "item_type", item_type, "item_id", item_id, "color_id", color_id, "code", response.Meta.Code, "message", response.Meta.Message, "description", response.Meta.Description)
		return "", errors.New("Failed to get pictures for part " + item_id + " color " + strconv.Itoa(color_id) + ": " + strconv.Itoa(response.Meta.Code) + " " + response.Meta.Message)
	}

	logger.Info("Bricklink Client: Successfully retrieved part pictures", "item_type", item_type, "item_id", item_id, "color_id", color_id, "response_length", len(response.responseStr))
	return response.responseStr, nil
}

// GetColors retrieves all colors from BrickLink API
func (blClient BricklinkCatalogClient) GetColors() (string, error) {
	logger.Debug("Bricklink Client: Starting GetColors request")

	endpoint := "/colors"
	logger.Debug("Bricklink Client: Making API request", "endpoint", endpoint)

	response, err := blClient.BLGet(endpoint)
	if err != nil {
		logger.Error("Bricklink Client: Colors API request failed", "endpoint", endpoint, "error", err)
		return "", err
	}

	logger.Debug("Bricklink Client: Colors API request completed", "response_code", response.Meta.Code, "response_message", response.Meta.Message)

	if response.Meta.Code != 200 {
		logger.Error("Bricklink Client: Colors API returned error response", "code", response.Meta.Code, "message", response.Meta.Message, "description", response.Meta.Description)
		return "", errors.New("Failed to get colors from BrickLink: " + strconv.Itoa(response.Meta.Code) + " " + response.Meta.Message)
	}

	logger.Debug("Bricklink Client: Successfully retrieved colors", "response_length", len(response.responseStr))
	return response.responseStr, nil
}
