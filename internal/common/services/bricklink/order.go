package bricklink

import (
	"strconv"
	"time"
)

type Order struct {
	ID                 int64      `json:"id" db:"id"`
	UserID             int64      `json:"user_id" db:"user_id"`
	BricklinkOrderID   int        `json:"bricklink_order_id" db:"bricklink_order_id"`
	
	// Order Dates
	DateOrdered        time.Time  `json:"date_ordered" db:"date_ordered"`
	DateStatusChanged  *time.Time `json:"date_status_changed" db:"date_status_changed"`
	
	// Seller Information
	SellerName         *string    `json:"seller_name" db:"seller_name"`
	StoreName          *string    `json:"store_name" db:"store_name"`
	
	// Buyer Information  
	BuyerName          string     `json:"buyer_name" db:"buyer_name"`
	BuyerEmail         *string    `json:"buyer_email" db:"buyer_email"`
	BuyerOrderCount    *int       `json:"buyer_order_count" db:"buyer_order_count"` // Previous orders from this buyer
	
	// Order Details
	Status             string     `json:"status" db:"status"`
	IsInvoiced         *bool      `json:"is_invoiced" db:"is_invoiced"`
	RequireInsurance   *bool      `json:"require_insurance" db:"require_insurance"`
	Remarks            *string    `json:"remarks" db:"remarks"`
	TotalCount         *int       `json:"total_count" db:"total_count"`
	UniqueCount        *int       `json:"unique_count" db:"unique_count"`
	TotalWeight        *string    `json:"total_weight" db:"total_weight"` // Stored as string from API
	IsFiled            *bool      `json:"is_filed" db:"is_filed"`
	DriveThruSent      *bool      `json:"drive_thru_sent" db:"drive_thru_sent"`
	
	// Payment Information
	PaymentMethod      *string    `json:"payment_method" db:"payment_method"`
	PaymentCurrencyCode *string   `json:"payment_currency_code" db:"payment_currency_code"`
	PaymentStatus      *string    `json:"payment_status" db:"payment_status"`
	DatePaid           *time.Time `json:"date_paid" db:"date_paid"`
	
	// Shipping Information
	ShippingMethodID   *int       `json:"shipping_method_id" db:"shipping_method_id"`
	ShippingMethod     *string    `json:"shipping_method" db:"shipping_method"`
	TrackingLink       *string    `json:"tracking_link" db:"tracking_link"`
	ShippingAddressName *string   `json:"shipping_address_name" db:"shipping_address_name"`
	ShippingAddressFull *string   `json:"shipping_address_full" db:"shipping_address_full"`
	ShippingCountryCode *string   `json:"shipping_country_code" db:"shipping_country_code"`
	
	// Cost Information
	CurrencyCode       string     `json:"currency_code" db:"currency_code"`
	Subtotal           *float64   `json:"subtotal" db:"subtotal"`
	TotalPrice         float64    `json:"total_price" db:"total_price"` // grand_total
	Etc1               *float64   `json:"etc1" db:"etc1"`
	Etc2               *float64   `json:"etc2" db:"etc2"`
	InsuranceCost      *float64   `json:"insurance_cost" db:"insurance_cost"`
	ShippingCost       *float64   `json:"shipping_cost" db:"shipping_cost"`
	CreditAmount       *float64   `json:"credit_amount" db:"credit_amount"`
	CouponAmount       *float64   `json:"coupon_amount" db:"coupon_amount"`
	VatRate            *float64   `json:"vat_rate" db:"vat_rate"`
	VatAmount          *float64   `json:"vat_amount" db:"vat_amount"`
	
	// Display Cost Information (separate from actual cost)
	DisplayCurrencyCode *string   `json:"display_currency_code" db:"display_currency_code"`
	DisplaySubtotal    *float64   `json:"display_subtotal" db:"display_subtotal"`
	DisplayGrandTotal  *float64   `json:"display_grand_total" db:"display_grand_total"`
	DisplayEtc1        *float64   `json:"display_etc1" db:"display_etc1"`
	DisplayEtc2        *float64   `json:"display_etc2" db:"display_etc2"`
	DisplayInsurance   *float64   `json:"display_insurance" db:"display_insurance"`
	DisplayShipping    *float64   `json:"display_shipping" db:"display_shipping"`
	DisplayCredit      *float64   `json:"display_credit" db:"display_credit"`
	DisplayCoupon      *float64   `json:"display_coupon" db:"display_coupon"`
	DisplayVatRate     *float64   `json:"display_vat_rate" db:"display_vat_rate"`
	DisplayVatAmount   *float64   `json:"display_vat_amount" db:"display_vat_amount"`
	
	// System fields
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at" db:"updated_at"`
}

type OrderSync struct {
	ID           int64      `json:"id" db:"id"`
	UserID       int64      `json:"user_id" db:"user_id"`
	LastSyncTime time.Time  `json:"last_sync_time" db:"last_sync_time"`
	SyncStatus   string     `json:"sync_status" db:"sync_status"` // "in_progress", "completed", "failed"
	OrdersCount  int        `json:"orders_count" db:"orders_count"`
	ErrorMessage *string    `json:"error_message" db:"error_message"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
}

// Implementation of models.Order interface for BrickLink Order
func (o *Order) GetID() string {
	return strconv.FormatInt(o.ID, 10)
}

func (o *Order) GetOrderID() string {
	return strconv.Itoa(o.BricklinkOrderID)
}

func (o *Order) GetUserID() int64 {
	return o.UserID
}

func (o *Order) GetBuyerName() string {
	return o.BuyerName
}

func (o *Order) GetSellerName() string {
	if o.SellerName != nil {
		return *o.SellerName
	}
	return ""
}

func (o *Order) GetStoreName() string {
	if o.StoreName != nil {
		return *o.StoreName
	}
	return ""
}

func (o *Order) GetStatus() string {
	return o.Status
}

func (o *Order) GetDateOrdered() time.Time {
	return o.DateOrdered
}

func (o *Order) GetTotalPrice() float64 {
	return o.TotalPrice
}

func (o *Order) GetCurrencyCode() string {
	return o.CurrencyCode
}

func (o *Order) GetTotalCount() int {
	if o.TotalCount != nil {
		return *o.TotalCount
	}
	return 0
}

func (o *Order) GetUniqueCount() int {
	if o.UniqueCount != nil {
		return *o.UniqueCount
	}
	return 0
}

func (o *Order) GetProvider() string {
	return "bricklink"
}

// Implementation of models.OrderSync interface for BrickLink OrderSync
func (os *OrderSync) GetUserID() int64 {
	return os.UserID
}

func (os *OrderSync) GetProvider() string {
	return "bricklink"
}

func (os *OrderSync) GetSyncStatus() string {
	return os.SyncStatus
}

func (os *OrderSync) GetOrdersCount() int {
	return os.OrdersCount
}

func (os *OrderSync) GetErrorMessage() *string {
	return os.ErrorMessage
}

func (os *OrderSync) GetLastSyncTime() time.Time {
	return os.LastSyncTime
}
