package bricklink

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/eric-schulze/we_sync_bricks/utils/oauth"
)

const BricklinkOrdersEndpoint = "/orders"

type BricklinkOrdersClient struct {
	BLClient
}

type BLOrder struct {
}

type BLOrders struct {
	Orders []BLOrder
	Count  int
}

func NewBricklinkOrdersClient() BricklinkOrdersClient {
	// Create a new OAuth client for the Bricklink API
	client := oauth.NewOAuthClient(ConsumerKey, ConsumerSecret, Token, TokenSecret)

	return BricklinkOrdersClient{
		BLClient{
			oAuthClient: client,
		},
	}
}

func (blClient BricklinkOrdersClient) GetOrders() (string, error) {
	response, err := blClient.BLGet(BricklinkOrdersEndpoint)
	if err != nil {
		return "", err
	}
	fmt.Println("retrieved orders")

	if response.Meta.Code != 200 {
		fmt.Println(response.Meta.Code)
		return "", errors.New("failed to get orders: " + strconv.Itoa(response.Meta.Code) + " " + response.Meta.Message)
	}

	return response.responseStr, nil
}
