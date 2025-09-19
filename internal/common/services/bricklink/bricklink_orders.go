package bricklink

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/eric-schulze/we_sync_bricks/utils/oauth"
)

const BricklinkOrdersEndpoint = "/orders"

type BricklinkOrdersClient struct {
	BLClient
}

type BLOrder struct {
	OrderID           int     `json:"order_id"`
	DateOrdered       string  `json:"date_ordered"`
	DateStatusChanged *string `json:"date_status_changed"`
	
	// Seller/Store Info
	SellerName        *string `json:"seller_name"`
	StoreName         *string `json:"store_name"`
	
	// Buyer Info
	BuyerName         string  `json:"buyer_name"`
	BuyerEmail        *string `json:"buyer_email"`
	BuyerOrderCount   *int    `json:"buyer_order_count"`
	
	// Order Details
	RequireInsurance  *bool   `json:"require_insurance"`
	Status            string  `json:"status"`
	IsInvoiced        *bool   `json:"is_invoiced"`
	Remarks           *string `json:"remarks"`
	TotalCount        *int    `json:"total_count"`
	UniqueCount       *int    `json:"unique_count"`
	TotalWeight       *string `json:"total_weight"`
	IsFiled           *bool   `json:"is_filed"`
	DriveThruSent     *bool   `json:"drive_thru_sent"`
	
	// Payment Info
	Payment           *BLOrderPayment `json:"payment"`
	
	// Shipping Info
	Shipping          *BLOrderShipping `json:"shipping"`
	
	// Cost Info
	Cost              *BLOrderCost `json:"cost"`
	DispCost          *BLOrderCost `json:"disp_cost"`
}

type BLOrderPayment struct {
	Method       *string `json:"method"`
	CurrencyCode *string `json:"currency_code"`
	DatePaid     *string `json:"date_paid"`
	Status       *string `json:"status"`
}

type BLOrderShipping struct {
	MethodID    *int                   `json:"method_id"`
	Method      *string                `json:"method"`
	TrackingLink *string               `json:"tracking_link"`
	Address     *BLOrderShippingAddress `json:"address"`
}

type BLOrderShippingAddress struct {
	Name        *BLOrderShippingName `json:"name"`
	Full        *string              `json:"full"`
	CountryCode *string              `json:"country_code"`
}

type BLOrderShippingName struct {
	Full *string `json:"full"`
}

type BLOrderCost struct {
	CurrencyCode *string `json:"currency_code"`
	Subtotal     *string `json:"subtotal"`
	GrandTotal   *string `json:"grand_total"`
	Etc1         *string `json:"etc1"`
	Etc2         *string `json:"etc2"`
	Insurance    *string `json:"insurance"`
	Shipping     *string `json:"shipping"`
	Credit       *string `json:"credit"`
	Coupon       *string `json:"coupon"`
	VatRate      *string `json:"vat_rate"`
	VatAmount    *string `json:"vat_amount"`
}

type BLOrdersResponse struct {
	Orders []BLOrder `json:"orders"`
}

type BricklinkOrdersOptions struct {
	UpdatedAfter *time.Time
	Status       string
	Direction    string
}

func NewBricklinkOrdersClient() BricklinkOrdersClient {
	client := oauth.NewOAuthClient(ConsumerKey, ConsumerSecret, Token, TokenSecret)

	return BricklinkOrdersClient{
		BLClient{
			oAuthClient: client,
		},
	}
}


func (blClient BricklinkOrdersClient) GetOrders() (string, error) {
	return blClient.GetOrdersWithOptions(nil)
}

func (blClient BricklinkOrdersClient) GetOrdersWithOptions(options *BricklinkOrdersOptions) (string, error) {
	endpoint := BricklinkOrdersEndpoint
	
	// Add query parameters if options are provided
	if options != nil {
		endpoint += "?"
		params := []string{}
		
		if options.UpdatedAfter != nil {
			// BrickLink expects date in YYYY-MM-DD format
			dateStr := options.UpdatedAfter.Format("2006-01-02")
			params = append(params, "updated_after="+dateStr)
		}
		
		if options.Status != "" {
			params = append(params, "status="+options.Status)
		}
		
		if options.Direction != "" {
			params = append(params, "direction="+options.Direction)
		} else {
			params = append(params, "direction=in") // Default to incoming orders
		}
		
		if len(params) > 0 {
			for i, param := range params {
				if i > 0 {
					endpoint += "&"
				}
				endpoint += param
			}
		}
	}

	logger.Info("Fetching orders from BrickLink", "endpoint", endpoint)

	response, err := blClient.BLGet(endpoint)
	if err != nil {
		logger.Error("Failed to get orders from BrickLink API", "error", err)
		return "", err
	}

	if response.Meta.Code != 200 {
		logger.Error("BrickLink API returned error", "code", response.Meta.Code, "message", response.Meta.Message)
		return "", errors.New("failed to get orders: " + strconv.Itoa(response.Meta.Code) + " " + response.Meta.Message)
	}

	logger.Info("Successfully retrieved orders from BrickLink", "response_length", len(response.responseStr))
	return response.responseStr, nil
}

func (blClient BricklinkOrdersClient) GetOrdersSince(lastSync time.Time) ([]BLOrder, error) {
	options := &BricklinkOrdersOptions{
		UpdatedAfter: &lastSync,
		Direction:    "in",
	}

	responseStr, err := blClient.GetOrdersWithOptions(options)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var blResponse BLResponse
	err = json.Unmarshal([]byte(responseStr), &blResponse)
	if err != nil {
		logger.Error("Failed to parse BrickLink orders response", "error", err)
		return nil, fmt.Errorf("error parsing orders JSON: %v", err)
	}

	// Convert the data interface{} to our order structure
	dataBytes, err := json.Marshal(blResponse.Data)
	if err != nil {
		logger.Error("Failed to marshal BrickLink data", "error", err)
		return nil, fmt.Errorf("error marshaling orders data: %v", err)
	}

	var orders []BLOrder
	err = json.Unmarshal(dataBytes, &orders)
	if err != nil {
		logger.Error("Failed to unmarshal orders data", "error", err)
		return nil, fmt.Errorf("error unmarshaling orders: %v", err)
	}

	logger.Info("Parsed BrickLink orders", "count", len(orders))
	return orders, nil
}
