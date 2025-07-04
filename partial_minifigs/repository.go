package partial_minifigs

import (
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
func (repo *PartialMinifigRepository) GetPartialMinifigListByID(id int, userID int64) (*PartialMinifigList, error) {
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
	list.PartialMinifigs = items

	return &list, nil
}

// GetPartialMinifigItemsByPartialMinifigListID retrieves all items for a specific list
func (repo *PartialMinifigRepository) GetPartialMinifigItemsByPartialMinifigListID(listID int) ([]PartialMinifig, error) {
	sql := "SELECT id, reference_id, partial_minifig_list_id, item_id, created_at, updated_at FROM partial_minifigs WHERE partial_minifig_list_id = $1 ORDER BY created_at ASC;"

	// Use the helper function that works with the DBService interface
	items, err := db.CollectRowsToStructFromService[PartialMinifig](repo.db, sql, listID)
	if err != nil {
		logger.Error("Database Error while retrieving partial minifig items", "list_id", listID, "error", err)
		return nil, err
	}

	return items, nil
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
func (repo *PartialMinifigRepository) UpdatePartialMinifigList(id int, name, description string, userID int64) (*PartialMinifigList, error) {
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
