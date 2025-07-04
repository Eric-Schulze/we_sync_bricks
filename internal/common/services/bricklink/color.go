package bricklink

import (
	"gorm.io/gorm"
)

type Color struct {
	gorm.Model
	ColorID int    `gorm:"uniqueIndex;not null" json:"color_id"`
	Name    string `json:"name"`
	RGB     string `json:"rgb"`
	IsTrans bool   `json:"is_trans"`
}

func (Color) TableName() string {
	return "colors"
}

func MigrateColor(db *gorm.DB) error {
	return db.AutoMigrate(&Color{})
}
