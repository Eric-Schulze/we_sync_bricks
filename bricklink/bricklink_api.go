package bricklink

import (
	"encoding/json"
	"fmt"

	"github.com/eric-schulze/we_sync_bricks/utils/oauth"
)

const ConsumerKey = "4ED3302C6D1644CEA64E511455D1467B"
const ConsumerSecret = "9A91A62C32E040EBA2B16694C5120C6F"
const Token = "BA8A79AD3C624A53AD3F67C2E5C21B1F"
const TokenSecret = "733B0F8729C64879BA4094CB2936FE05"

const BricklinkApiBaseUrl = "https://api.bricklink.com/api/store/v1"

type BLResponse struct {
	Meta 		BLResponseMeta    `json:"meta"`
	Data 		any 			   `json:"data"`
	responseStr string
}

type BLResponseMeta struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

type BLClient struct {
	oAuthClient *oauth.OAuthClient
}

func (client BLClient) BLGet(endpoint string) (*BLResponse, error) {
	path := BricklinkApiBaseUrl + endpoint
	response, err := client.oAuthClient.Get(path)
	if err != nil {
		return &BLResponse{}, err
	}

	blResponse, err := parseBLResponse(string(response))
	if err != nil {
		return &BLResponse{}, err
	}

	return blResponse, nil
}

func parseBLResponse(jsonStr string) (*BLResponse, error) {
    var response BLResponse
    err := json.Unmarshal([]byte(jsonStr), &response)
    if err != nil {
        return nil, fmt.Errorf("error parsing JSON: %v", err)
    }

	response.responseStr = jsonStr
	
    return &response, nil
}
