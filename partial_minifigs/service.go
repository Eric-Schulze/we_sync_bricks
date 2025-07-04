package partial_minifigs

import (
	"errors"
	"html/template"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/bricklink"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

type PartialMinifigService struct {
	repo           PartialMinifigRepository
	clientManager  *bricklink.ClientManager
}

// GetAllPartialMinifigLists returns all partial minifig lists for the specified user
func (s *PartialMinifigService) GetAllPartialMinifigLists(user *models.User) ([]PartialMinifigList, error) {
	return s.repo.GetAllPartialMinifigLists(user.ID)
}

// GetPartialMinifigListByID returns a specific list by ID for the specified user
func (s *PartialMinifigService) GetPartialMinifigListByID(id int, user *models.User) (*PartialMinifigList, error) {
	return s.repo.GetPartialMinifigListByID(id, user.ID)
}

// CreatePartialMinifigList creates a new partial minifig list for the specified user
func (s *PartialMinifigService) CreatePartialMinifigList(name, description string, user *models.User) (*PartialMinifigList, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	// Create the list using the repository
	newList, err := s.repo.CreatePartialMinifigList(name, description, user.ID)
	if err != nil {
		return nil, err
	}

	// Initialize with empty minifigs list
	newList.PartialMinifigs = []PartialMinifig{}

	return newList, nil
}

// UpdatePartialMinifigList updates an existing list for the specified user
func (s *PartialMinifigService) UpdatePartialMinifigList(id int, name, description string, user *models.User) (*PartialMinifigList, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	// Update the list using the repository
	updatedList, err := s.repo.UpdatePartialMinifigList(id, name, description, user.ID)
	if err != nil {
		return nil, err
	}

	return updatedList, nil
}

// SearchBricklinkItems searches Bricklink for items matching the query
func (s *PartialMinifigService) SearchBricklinkItems(searchTerms string, user *models.User) ([]bricklink.Item, error) {
	if searchTerms == "" {
		return nil, errors.New("search terms cannot be empty")
	}

	// TODO: Implement BrickLink search integration when the method exists
	// For now, return mock data
	return []bricklink.Item{}, nil
}

// GetBricklinkItem gets a specific item from Bricklink by ID
func (s *PartialMinifigService) GetBricklinkItem(itemID string, user *models.User) (*bricklink.Item, error) {
	logger.Info("Service: Starting GetBricklinkItem", "item_id", itemID, "user_id", user.ID)
	
	if itemID == "" {
		logger.Warn("Service: GetBricklinkItem called with empty item ID", "user_id", user.ID)
		return nil, errors.New("item ID cannot be empty")
	}

	logger.Info("Service: Getting Bricklink client for user", "item_id", itemID, "user_id", user.ID)

	// Get user-specific Bricklink client from the client manager
	client, err := s.clientManager.GetClient(user)
	if err != nil {
		logger.Error("Service: Failed to get Bricklink client for user", "item_id", itemID, "user_id", user.ID, "error", err)
		return nil, err
	}

	logger.Info("Service: Calling user-specific Bricklink client to get item", "item_id", itemID, "user_id", user.ID)

	// Use the user-specific Bricklink client to get the item
	responseData, err := client.GetItem(itemID)
	if err != nil {
		logger.Error("Service: Bricklink client returned error", "item_id", itemID, "user_id", user.ID, "error", err)
		return nil, err
	}

	logger.Info("Service: Bricklink client returned data", "item_id", itemID, "user_id", user.ID, "response_length", len(responseData))

	// TODO: Parse the JSON response properly
	// For now, return a mock item with the ID
	mockItem := &bricklink.Item{
		No:   itemID,
		Name: "Mock Item - " + itemID,
		Type: "PART",
	}

	logger.Info("Service: Successfully created mock item", "item_id", itemID, "user_id", user.ID, "item_name", mockItem.Name, "item_type", mockItem.Type)

	return mockItem, nil
}

// CreatePartialMinifig creates a new partial minifig in a list for the specified user
func (s *PartialMinifigService) CreatePartialMinifig(listID int, name, description string, user *models.User) (*PartialMinifigList, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	// TODO: Implement actual logic to save to database using repository with user validation
	newMinifig := PartialMinifig{
		ID:                      1, // TODO: Generate proper ID from database
		ReferenceID:             nil,
		PartialMinifigListID: listID,
		ItemID:                  1, // TODO: Get actual item ID
		CreatedAt:               time.Now(),
		UpdatedAt:               nil,
	}

	return &PartialMinifigList{
		ID:              listID,
		Name:            "Sample List",
		Description:     nil,
		UserID:          user.ID,
		PartialMinifigs: []PartialMinifig{newMinifig},
		CreatedAt:       time.Now(),
		UpdatedAt:       nil,
	}, nil
}

// GetPartialMinifigByID returns a specific partial minifig by ID for the specified user
func (s *PartialMinifigService) GetPartialMinifigByID(id int, user *models.User) (*PartialMinifig, error) {
	// TODO: Implement actual logic to fetch from database using repository with user validation
	return &PartialMinifig{
		ID:                      id,
		ReferenceID:             nil,
		PartialMinifigListID: 1,
		ItemID:                  1,
		CreatedAt:               time.Now(),
		UpdatedAt:               nil,
	}, nil
}

// UpdatePartialMinifig updates a partial minifig in a list for the specified user
func (s *PartialMinifigService) UpdatePartialMinifig(listID, itemID int, name, description string, user *models.User) (*PartialMinifigList, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	// TODO: Implement actual logic to update in database using repository with user validation
	now := time.Now()
	updatedMinifig := PartialMinifig{
		ID:                      itemID,
		ReferenceID:             nil,
		PartialMinifigListID: listID,
		ItemID:                  1, // TODO: Get actual item ID
		CreatedAt:               time.Now(),
		UpdatedAt:               &now,
	}

	return &PartialMinifigList{
		ID:              listID,
		Name:            "Sample List",
		Description:     nil,
		UserID:          user.ID,
		PartialMinifigs: []PartialMinifig{updatedMinifig},
		CreatedAt:       time.Now(),
		UpdatedAt:       &now,
	}, nil
}

// DeletePartialMinifig deletes a partial minifig from a list for the specified user
func (s *PartialMinifigService) DeletePartialMinifig(listID, itemID int, user *models.User) (*PartialMinifigList, error) {
	// TODO: Implement actual logic to delete from database using repository with user validation
	now := time.Now()
	return &PartialMinifigList{
		ID:              listID,
		Name:            "Sample List",
		Description:     nil,
		UserID:          user.ID,
		PartialMinifigs: []PartialMinifig{}, // Empty after deletion
		CreatedAt:       time.Now(),
		UpdatedAt:       &now,
	}, nil
}

// InitializePartialMinifigHandler creates a fully initialized partial minifig handler with all dependencies
func InitializePartialMinifigHandler(dbService models.DBService, templates *template.Template, jwtSecret []byte) *PartialMinifigHandler {
	repo := PartialMinifigRepository{db: dbService}
	
	// Initialize client manager with 30-minute cache TTL and max 100 clients
	clientManager := bricklink.NewClientManager(30*time.Minute, 100, dbService)
	
	service := &PartialMinifigService{
		repo:          repo,
		clientManager: clientManager,
	}
	handler := NewPartialMinifigHandler(service, templates, jwtSecret)
	return handler
}
