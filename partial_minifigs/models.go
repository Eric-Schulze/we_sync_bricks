package partial_minifigs

import (
	"time"
)

// PartialMinifigList represents a collection of partial minifigs based on partial_minifig_lists table
type PartialMinifigList struct {
	ID                  int64            `json:"id" db:"id"`
	Name                string           `json:"name" db:"name"`
	Description         *string          `json:"description" db:"description"`
	UserID              int64            `json:"user_id" db:"user_id"`
	PartialMinifigs     []PartialMinifig `json:"partial_minifigs" db:"-"`
	PartialMinifigCount int64            `json:"partial_minifig_count" db:"-"`
	CreatedAt           time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt           *time.Time       `json:"updated_at" db:"updated_at"`
}

// PartialMinifig represents a single partial minifig based on partial_minifigs table
type PartialMinifig struct {
	ID                   int64                `json:"id" db:"id"`
	ReferenceID          *string              `json:"reference_id" db:"reference_id"`
	Condition            *string              `json:"condition" db:"condition"`
	Notes                *string              `json:"notes" db:"notes"`
	PartialMinifigListID int64                `json:"partial_minifig_list_id" db:"partial_minifig_list_id"`
	ItemID               int64                `json:"item_id" db:"item_id"`
	BricklinkID          *string              `json:"bricklink_id" db:"bricklink_id"`
	ItemName             *string              `json:"item_name" db:"item_name"`
	Parts                []PartialMinifigPart `json:"parts,omitempty" db:"-"`
	CreatedAt            time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt            *time.Time           `json:"updated_at" db:"updated_at"`
}

// PartialMinifigPart represents a part needed for a partial minifig based on partial_minifig_parts table
type PartialMinifigPart struct {
	ID                int64      `json:"id" db:"id"`
	PartialMinifigID  int64      `json:"partial_minifig_id" db:"partial_minifig_id"`
	ItemID            int64      `json:"item_id" db:"item_id"`
	ColorID           int64      `json:"color_id" db:"color_id"`
	BricklinkID       *string    `json:"bricklink_id" db:"bricklink_id"`
	PartName          *string    `json:"part_name" db:"part_name"`
	ColorName         *string    `json:"color_name" db:"color_name"`
	ColorCode         *string    `json:"color_code" db:"color_code"`
	Condition         *string    `json:"condition" db:"condition"`
	QuantityNeeded    int        `json:"quantity_needed" db:"quantity_needed"`
	QuantityCollected int        `json:"quantity_collected" db:"quantity_collected"`
	IsCollected       bool       `json:"is_collected" db:"is_collected"`
	CostPerPiece      float64    `json:"cost_per_piece" db:"cost_per_piece"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at" db:"updated_at"`
}
