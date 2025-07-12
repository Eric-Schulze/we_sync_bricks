package bricklink

import (
	"errors"
	"strconv"

	"github.com/eric-schulze/we_sync_bricks/utils/oauth"
)

const BricklinkInventoriesEndpoint = "/inventories"

type BricklinkInventoryClient struct {
	BLClient
}

func NewBricklinkInventoryClient() BricklinkInventoryClient {
	// Create a new OAuth client for the Bricklink API
	client := oauth.NewOAuthClient(ConsumerKey, ConsumerSecret, Token, TokenSecret)

	return BricklinkInventoryClient{
		BLClient{
			oAuthClient: client,
		},
	}
}

func (blClient BricklinkInventoryClient) GetInventory() (string, error) {
	response, err := blClient.BLGet(BricklinkInventoriesEndpoint)
	if err != nil {
		return "", err
	}

	if response.Meta.Code != 200 {
		return "", errors.New("failed to get orders: " + strconv.Itoa(response.Meta.Code) + " " + response.Meta.Message)
	}

	return response.responseStr, nil
}
