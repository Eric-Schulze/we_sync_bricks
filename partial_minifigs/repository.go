package partial_minifigs

import (
	"errors"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/db"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

type PartialMinifigRepository struct {
	db models.DBService
}

// NewPartialMinifigRepository creates a new repository instance
func NewPartialMinifigRepository(dbService models.DBService) *PartialMinifigRepository {
	return &PartialMinifigRepository{
		db: dbService,
	}
}

// GetAllPartialMinifigLists retrieves all partial minifig lists for a specific user from the database
func (repo *PartialMinifigRepository) GetAllPartialMinifigLists(userID int64) ([]PartialMinifigList, error) {
	sql := "SELECT id, name, description, user_id, created_at, updated_at FROM partial_minifig_lists WHERE user_id = $1 ORDER BY created_at DESC;"

	// Use the helper function that works with the DBService interface
	lists, err := db.CollectRowsToStructFromService[PartialMinifigList](repo.db, sql, userID)
	if err != nil {
		logger.Error("Database Error while retrieving partial minifig lists", "user_id", userID, "error", err)
		return nil, err
	}

	return lists, nil
}

// GetPartialMinifigListByID retrieves a specific partial minifig list with its items for a specific user
func (repo *PartialMinifigRepository) GetPartialMinifigListByID(id int64, userID int64) (*PartialMinifigList, error) {
	sql := "SELECT id, name, description, user_id, created_at, updated_at FROM partial_minifig_lists WHERE id = $1 AND user_id = $2;"

	// Use the helper function that works with the DBService interface
	list, err := db.QueryRowToStructFromService[PartialMinifigList](repo.db, sql, id, userID)
	if err != nil {
		logger.Error("Database Error while retrieving partial minifig list", "id", id, "user_id", userID, "error", err)
		return nil, err
	}

	// Get items for this list
	items, err := repo.GetPartialMinifigItemsByPartialMinifigListID(id)
	if err != nil {
		logger.Error("Database Error while retrieving list items", "list_id", id, "error", err)
		return &list, err // Return list even if items fail
	}

	// Load parts for each partial minifig
	for i := range items {
		parts, err := repo.GetPartialMinifigParts(items[i].ID)
		if err != nil {
			logger.Error("Database Error while retrieving parts for partial minifig", "partial_minifig_id", items[i].ID, "error", err)
			// Continue with empty parts rather than failing
			items[i].Parts = []PartialMinifigPart{}
		} else {
			items[i].Parts = parts
		}
	}

	list.PartialMinifigs = items

	return &list, nil
}

// GetPartialMinifigItemsByPartialMinifigListID retrieves all items for a specific list
func (repo *PartialMinifigRepository) GetPartialMinifigItemsByPartialMinifigListID(listID int64) ([]PartialMinifig, error) {
	sql := `SELECT 
		pm.id, 
		pm.reference_id, 
		pm.condition, 
		pm.notes, 
		pm.partial_minifig_list_id, 
		pm.item_id, 
		i.bricklink_id,
		ib.name as item_name,
		pm.created_at, 
		pm.updated_at 
	FROM partial_minifigs pm
	LEFT JOIN items i ON pm.item_id = i.id
	LEFT JOIN items_bricklink ib ON i.bricklink_id = ib.id
	WHERE pm.partial_minifig_list_id = $1 
	ORDER BY pm.created_at ASC;`

	// Use the helper function that works with the DBService interface
	items, err := db.CollectRowsToStructFromService[PartialMinifig](repo.db, sql, listID)
	if err != nil {
		logger.Error("Database Error while retrieving partial minifig items", "list_id", listID, "error", err)
		return nil, err
	}

	return items, nil
}

// GetPartialMinifigParts retrieves all parts for a specific partial minifig
func (repo *PartialMinifigRepository) GetPartialMinifigParts(partialMinifigID int64) ([]PartialMinifigPart, error) {
	sql := `SELECT 
		pmp.id, 
		pmp.partial_minifig_id, 
		pmp.item_id, 
		pmp.color_id, 
		pmp.condition,
		pmp.quantity_needed, 
		pmp.quantity_collected, 
		pmp.is_collected, 
		pmp.cost_per_piece, 
		i.bricklink_id,
		ib.name as part_name,
		pmp.created_at, 
		pmp.updated_at 
	FROM partial_minifig_parts pmp
	LEFT JOIN items i ON pmp.item_id = i.id
	LEFT JOIN items_bricklink ib ON i.bricklink_id = ib.id
	WHERE pmp.partial_minifig_id = $1 
	ORDER BY pmp.created_at ASC;`

	// Use the helper function that works with the DBService interface
	parts, err := db.CollectRowsToStructFromService[PartialMinifigPart](repo.db, sql, partialMinifigID)
	if err != nil {
		logger.Error("Database Error while retrieving partial minifig parts", "partial_minifig_id", partialMinifigID, "error", err)
		return nil, err
	}

	return parts, nil
}

// CreatePartialMinifigList creates a new partial minifig list in the database
func (repo *PartialMinifigRepository) CreatePartialMinifigList(name, description string, userID int64) (*PartialMinifigList, error) {
	sql := `INSERT INTO partial_minifig_lists (name, description, user_id, created_at) 
			VALUES ($1, NULLIF($2, ''), $3, CURRENT_TIMESTAMP) 
			RETURNING id, name, description, user_id, created_at, updated_at;`

	// Use the helper function that works with the DBService interface
	newList, err := db.QueryRowToStructFromService[PartialMinifigList](repo.db, sql, name, description, userID)
	if err != nil {
		logger.Error("Database Error while creating partial minifig list", "name", name, "user_id", userID, "error", err)
		return nil, err
	}

	return &newList, nil
}

// UpdatePartialMinifigList updates an existing partial minifig list in the database
func (repo *PartialMinifigRepository) UpdatePartialMinifigList(id int64, name, description string, userID int64) (*PartialMinifigList, error) {
	sql := `UPDATE partial_minifig_lists 
			SET name = $1, description = NULLIF($2, ''), updated_at = CURRENT_TIMESTAMP 
			WHERE id = $3 AND user_id = $4
			RETURNING id, name, description, user_id, created_at, updated_at;`

	// Use the helper function that works with the DBService interface
	updatedList, err := db.QueryRowToStructFromService[PartialMinifigList](repo.db, sql, name, description, id, userID)
	if err != nil {
		logger.Error("Database Error while updating partial minifig list", "id", id, "name", name, "user_id", userID, "error", err)
		return nil, err
	}

	// Get items for this list
	items, err := repo.GetPartialMinifigItemsByPartialMinifigListID(id)
	if err != nil {
		logger.Error("Database Error while retrieving list items after update", "list_id", id, "error", err)
		return &updatedList, err // Return updated list even if items fail
	}
	updatedList.PartialMinifigs = items

	return &updatedList, nil
}

// CreatePartialMinifigWithParts creates a new partial minifig with its associated parts
func (repo *PartialMinifigRepository) CreatePartialMinifigWithParts(listID int64, minifigID string, referenceID *string, condition *string, notes *string, selectedParts []SelectedPart, userID int64) (*PartialMinifig, error) {
	logger.Info("Repository: Starting CreatePartialMinifigWithParts", "list_id", listID, "minifig_id", minifigID, "parts_count", len(selectedParts), "user_id", userID)

	// First, find or create an items record for the minifig
	itemID, err := repo.FindOrCreateItem(minifigID)
	if err != nil {
		logger.Error("Repository: Failed to find or create item", "minifig_id", minifigID, "error", err)
		return nil, err
	}

	// Now create the partial minifig entry
	sql := `INSERT INTO partial_minifigs (reference_id, condition, notes, partial_minifig_list_id, item_id, created_at) 
			VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP) 
			RETURNING id, reference_id, condition, notes, partial_minifig_list_id, item_id, created_at, updated_at;`

	partialMinifig, err := db.QueryRowToStructFromService[PartialMinifig](repo.db, sql, referenceID, condition, notes, listID, itemID)
	if err != nil {
		logger.Error("Database Error while creating partial minifig", "list_id", listID, "minifig_id", minifigID, "item_id", itemID, "user_id", userID, "error", err)
		return nil, err
	}

	logger.Info("Repository: Created partial minifig", "partial_minifig_id", partialMinifig.ID, "list_id", listID, "item_id", itemID)

	// Now create the associated parts
	for _, part := range selectedParts {
		err := repo.CreatePartialMinifigPart(partialMinifig.ID, part)
		if err != nil {
			logger.Error("Repository: Failed to create part", "partial_minifig_id", partialMinifig.ID, "part_no", part.PartNo, "error", err)
			// Continue with other parts even if one fails
		}
	}

	logger.Info("Repository: Successfully created partial minifig with parts", "partial_minifig_id", partialMinifig.ID, "parts_count", len(selectedParts))

	return &partialMinifig, nil
}

// CreatePartialMinifigPart creates a single part entry for a partial minifig
func (repo *PartialMinifigRepository) CreatePartialMinifigPart(partialMinifigID int64, part SelectedPart) error {
	// First, find or create an item for this part
	itemID, err := repo.FindOrCreateItem(part.PartNo)
	if err != nil {
		logger.Error("Repository: Failed to find or create item for part", "part_no", part.PartNo, "error", err)
		return err
	}

	sql := `INSERT INTO partial_minifig_parts 
			(partial_minifig_id, item_id, color_id, condition, quantity_needed, quantity_collected, is_collected, cost_per_piece, created_at) 
			VALUES ($1, $2, $3, $4, $5, 0, false, $6, CURRENT_TIMESTAMP);`

	// Use the new price if available, otherwise use used price
	costPerPiece := part.NewPrice
	if costPerPiece == 0 && part.UsedPrice > 0 {
		costPerPiece = part.UsedPrice
	}

	// Use colorID as-is, defaulting to 0 if not provided
	colorID := part.ColorID
	if colorID < 0 {
		colorID = 0
	}

	// Set condition to empty string if not provided
	condition := part.Condition
	if condition == "" {
		condition = "N" // Default to new condition
	}

	_, err = repo.db.ExecSQL(sql, partialMinifigID, itemID, colorID, condition, part.Quantity, costPerPiece)
	if err != nil {
		logger.Error("Database Error while creating partial minifig part", "partial_minifig_id", partialMinifigID, "part_no", part.PartNo, "item_id", itemID, "error", err)
		return err
	}

	logger.Info("Repository: Created partial minifig part", "partial_minifig_id", partialMinifigID, "part_no", part.PartNo, "item_id", itemID, "quantity", part.Quantity)
	return nil
}

// FindOrCreateBricklinkItem finds an existing BrickLink item or creates a new one with minimal data
func (repo *PartialMinifigRepository) FindOrCreateBricklinkItem(bricklinkID string) error {
	logger.Info("Repository: Finding or creating BrickLink item", "bricklink_id", bricklinkID)

	// First, check if the item already exists in items_bricklink
	findSQL := `SELECT id FROM items_bricklink WHERE id = $1;`

	rows, err := repo.db.CollectRowsToMap(findSQL, bricklinkID)
	if err != nil {
		logger.Error("Database Error while finding BrickLink item", "bricklink_id", bricklinkID, "error", err)
		return err
	}

	if len(rows) > 0 {
		logger.Info("Repository: Found existing BrickLink item", "bricklink_id", bricklinkID)
		return nil
	}

	// If not found, create a new entry in items_bricklink with minimal required data
	createSQL := `INSERT INTO items_bricklink (id, name, type, created_at) 
				  VALUES ($1, $2, $3, CURRENT_TIMESTAMP);`

	// Use generic values since we don't have the actual item data yet
	name := "Unknown Item " + bricklinkID
	itemType := "PART" // Default type

	_, err = repo.db.ExecSQL(createSQL, bricklinkID, name, itemType)
	if err != nil {
		logger.Error("Database Error while creating BrickLink item", "bricklink_id", bricklinkID, "error", err)
		return err
	}

	logger.Info("Repository: Created new BrickLink item", "bricklink_id", bricklinkID)
	return nil
}

// FindOrCreateItem finds an existing item by BrickLink ID or creates a new one
func (repo *PartialMinifigRepository) FindOrCreateItem(bricklinkID string) (int64, error) {
	logger.Info("Repository: Finding or creating item", "bricklink_id", bricklinkID)

	// First, try to find an existing item using CollectRowsToMap
	findSQL := `SELECT id FROM items WHERE bricklink_id = $1;`

	rows, err := repo.db.CollectRowsToMap(findSQL, bricklinkID)
	if err != nil {
		logger.Error("Database Error while finding item", "bricklink_id", bricklinkID, "error", err)
		return 0, err
	}

	if len(rows) > 0 {
		if id, ok := rows[0]["id"]; ok {
			if itemID, ok := id.(int64); ok {
				logger.Info("Repository: Found existing item", "bricklink_id", bricklinkID, "item_id", itemID)
				return itemID, nil
			}
		}
	}

	// If not found, ensure the BrickLink item exists first (for FK constraint)
	err = repo.FindOrCreateBricklinkItem(bricklinkID)
	if err != nil {
		logger.Error("Repository: Failed to find or create BrickLink item", "bricklink_id", bricklinkID, "error", err)
		return 0, err
	}

	// Now create the main item entry using CollectRowsToMap for the INSERT
	createSQL := `INSERT INTO items (bricklink_id, created_at) 
				  VALUES ($1, CURRENT_TIMESTAMP) 
				  RETURNING id;`

	createRows, err := repo.db.CollectRowsToMap(createSQL, bricklinkID)
	if err != nil {
		logger.Error("Database Error while creating item", "bricklink_id", bricklinkID, "error", err)
		return 0, err
	}

	if len(createRows) > 0 {
		if id, ok := createRows[0]["id"]; ok {
			if itemID, ok := id.(int64); ok {
				logger.Info("Repository: Created new item", "bricklink_id", bricklinkID, "item_id", itemID)
				return itemID, nil
			}
		}
	}

	return 0, errors.New("failed to create item")
}
