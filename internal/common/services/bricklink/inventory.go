package bricklink

import (
)

type Inventory struct {
	ItemID       uint   `gorm:"not null" json:"-"`
	Item         Item   `gorm:"foreignKey:ItemID" json:"item"`
	ColorID      int    `gorm:"not null" json:"color_id"`
	Color        Color  `gorm:"foreignKey:ColorID;references:ColorID" json:"color"`
	Quantity     int    `json:"quantity"`
	NewOrUsed    string `json:"new_or_used"`    // "N" or "U"
	Completeness string `json:"completeness"`   // "C", "I", etc.
	UnitPrice    float64 `json:"unit_price"`
}

func (Inventory) TableName() string {
	return "inventories"
}

