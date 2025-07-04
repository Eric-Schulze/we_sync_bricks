package bricklink

import (
	"errors"
	"strconv"

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

func (blClient BricklinkCatalogClient) GetItem(item_id string) (string, error) {
	logger.Info("Bricklink Client: Starting GetItem request", "item_id", item_id)
	
	endpoint := BricklinkCatalogEndpoint + "/" + item_id
	logger.Info("Bricklink Client: Making API request", "item_id", item_id, "endpoint", endpoint)
	
	response, err := blClient.BLGet(endpoint)
	if err != nil {
		logger.Error("Bricklink Client: API request failed", "item_id", item_id, "endpoint", endpoint, "error", err)
		return "", err
	}

	logger.Info("Bricklink Client: API request completed", "item_id", item_id, "response_code", response.Meta.Code, "response_message", response.Meta.Message)

	if response.Meta.Code != 200 {
		logger.Error("Bricklink Client: API returned error response", "item_id", item_id, "code", response.Meta.Code, "message", response.Meta.Message)
		return "", errors.New("Failed to get item from Bricklink: " + strconv.Itoa(response.Meta.Code) + " " + response.Meta.Message)
	}

	logger.Info("Bricklink Client: Successfully retrieved item data", "item_id", item_id, "response_length", len(response.responseStr))
	return response.responseStr, nil
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