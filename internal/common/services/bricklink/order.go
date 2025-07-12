package bricklink

import (
	"gorm.io/gorm"
	"time"
)

type Order struct {
	gorm.Model
	OrderID      int       `gorm:"uniqueIndex;not null" json:"order_id"`
	DateOrdered  time.Time `json:"date_ordered"`
	BuyerName    string    `json:"buyer_name"`
	Status       string    `json:"status"` // e.g. "Pending", "Paid"
	TotalPrice   float64   `json:"total_price"`
	CurrencyCode string    `json:"currency_code"`
}

func (Order) TableName() string {
	return "orders"
}

func MigrateOrder(db *gorm.DB) error {
	return db.AutoMigrate(&Order{})
}
