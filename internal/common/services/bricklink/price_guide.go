package bricklink

import (
	"gorm.io/gorm"
)

type PriceGuide struct {
	gorm.Model
	ItemNo       string  `gorm:"index;not null" json:"item_no"`
	ColorID      int     `gorm:"not null" json:"color_id"`
	NewAvgPrice  float64 `json:"new_avg_price"`
	UsedAvgPrice float64 `json:"used_avg_price"`
	CurrencyCode string  `json:"currency_code"`
}

func (PriceGuide) TableName() string {
	return "price_guides"
}

func MigratePriceGuide(db *gorm.DB) error {
	return db.AutoMigrate(&PriceGuide{})
}
