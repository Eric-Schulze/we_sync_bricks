package bricklink

import (
	"gorm.io/gorm"
)

type Item struct {
	gorm.Model
	No         string `gorm:"index;not null" json:"no"` // BrickLink item number (e.g., "3001")
	Name       string `json:"name"`
	Type       string `json:"type"` // PART, SET, MINIFIG, BOOK, etc.
	CategoryID int    `json:"category_id"`
}

func (Item) TableName() string {
	return "items"
}

func MigrateItem(db *gorm.DB) error {
	return db.AutoMigrate(&Item{})
}
