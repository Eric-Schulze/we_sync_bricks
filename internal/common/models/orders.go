package models

import "time"

// Order defines the common interface that all provider-specific order types must implement
type Order interface {
	GetID() string
	GetOrderID() string
	GetUserID() int64
	GetBuyerName() string
	GetSellerName() string
	GetStoreName() string
	GetStatus() string
	GetDateOrdered() time.Time
	GetTotalPrice() float64
	GetCurrencyCode() string
	GetTotalCount() int
	GetUniqueCount() int
	GetProvider() string // "bricklink", "brickowl", etc.
}

// OrderSync defines the common interface for order synchronization results
type OrderSync interface {
	GetUserID() int64
	GetProvider() string
	GetSyncStatus() string
	GetOrdersCount() int
	GetErrorMessage() *string
	GetLastSyncTime() time.Time
}

// OrderFilters defines common filtering options for orders across providers
type OrderFilters struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
	Status   string `form:"status"`
	Search   string `form:"search"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
	Provider string `form:"provider"` // Filter by provider
}

// GetDefaults returns OrderFilters with default values
func (f *OrderFilters) GetDefaults() *OrderFilters {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 25
	}
	if f.Sort == "" {
		f.Sort = "date_ordered"
	}
	if f.Order == "" {
		f.Order = "desc"
	}
	return f
}