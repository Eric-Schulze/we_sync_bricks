package partial_minifigs

import (
	"time"
)

// PartialMinifigList represents a collection of partial minifigs based on partial_minifig_lists table
type PartialMinifigList struct {
	ID               int              `json:"id" db:"id"`
	Name             string           `json:"name" db:"name"`
	Description      *string          `json:"description" db:"description"`
	UserID           int64            `json:"user_id" db:"user_id"`
	PartialMinifigs  []PartialMinifig `json:"partial_minifigs" db:"-"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        *time.Time       `json:"updated_at" db:"updated_at"`
}

// PartialMinifig represents a single partial minifig based on partial_minifigs table
type PartialMinifig struct {
	ID                     int                   `json:"id" db:"id"`
	ReferenceID            *string               `json:"reference_id" db:"reference_id"`
	PartialMinifigListID   int                   `json:"partial_minifig_list_id" db:"partial_minifig_list_id"`
	ItemID                 int64                 `json:"item_id" db:"item_id"`
	Parts                  []PartialMinifigPart  `json:"parts,omitempty"`
	CreatedAt              time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt              *time.Time            `json:"updated_at" db:"updated_at"`
}

// PartialMinifigPart represents a part needed for a partial minifig based on partial_minifig_parts table
type PartialMinifigPart struct {
	ID                 int       `json:"id" db:"id"`
	PartialMinifigID   int       `json:"partial_minifig_id" db:"partial_minifig_id"`
	ItemID             int64     `json:"item_id" db:"item_id"`
	ColorID            int       `json:"color_id" db:"color_id"`
	QuantityNeeded     int       `json:"quantity_needed" db:"quantity_needed"`
	QuantityCollected  int       `json:"quantity_collected" db:"quantity_collected"`
	IsCollected        bool      `json:"is_collected" db:"is_collected"`
	CostPerPiece       float64   `json:"cost_per_piece" db:"cost_per_piece"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at" db:"updated_at"`
}