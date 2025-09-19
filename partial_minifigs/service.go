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
	repo          PartialMinifigRepository
	clientManager *bricklink.ClientManager
}

// GetAllPartialMinifigLists returns all partial minifig lists for the specified user
func (s *PartialMinifigService) GetAllPartialMinifigLists(user *models.User) ([]PartialMinifigList, error) {
	return s.repo.GetAllPartialMinifigLists(user.ID)
}

// GetPartialMinifigListByID returns a specific list by ID for the specified user
func (s *PartialMinifigService) GetPartialMinifigListByID(id int64, user *models.User) (*PartialMinifigList, error) {
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
func (s *PartialMinifigService) UpdatePartialMinifigList(id int64, name, description string, user *models.User) (*PartialMinifigList, error) {
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
	item, err := client.GetItem(itemID)
	if err != nil {
		logger.Error("Service: Bricklink client returned error", "item_id", itemID, "user_id", user.ID, "error", err)
		return nil, err
	}

	logger.Info("Service: Successfully retrieved and parsed Bricklink item", "item_id", itemID, "user_id", user.ID, "item_name", item.Name, "item_type", item.Type)

	return item, nil
}

// GetMinifigPicture retrieves picture data for a minifigure
func (s *PartialMinifigService) GetMinifigPicture(itemID string, user *models.User) (string, error) {
	logger.Info("Service: Starting GetMinifigPicture", "item_id", itemID, "user_id", user.ID)

	if itemID == "" {
		logger.Warn("Service: GetMinifigPicture called with empty item ID", "user_id", user.ID)
		return "", errors.New("item ID cannot be empty")
	}

	logger.Info("Service: Getting Bricklink client for user", "item_id", itemID, "user_id", user.ID)

	// Get user-specific Bricklink client from the client manager
	client, err := s.clientManager.GetClient(user)
	if err != nil {
		logger.Error("Service: Failed to get Bricklink client for user", "item_id", itemID, "user_id", user.ID, "error", err)
		return "", err
	}

	logger.Info("Service: Calling user-specific Bricklink client to get pictures", "item_id", itemID, "user_id", user.ID)

	// Use the user-specific Bricklink client to get the pictures
	pictureData, err := client.GetMinifigPictures(itemID)
	if err != nil {
		logger.Error("Service: Bricklink client returned error for pictures", "item_id", itemID, "user_id", user.ID, "error", err)
		return "", err
	}

	logger.Info("Service: Successfully retrieved minifig pictures", "item_id", itemID, "user_id", user.ID)

	return pictureData, nil
}

// GetMinifigPricing retrieves pricing data for a minifigure
func (s *PartialMinifigService) GetMinifigPricing(itemID string, condition string, user *models.User) (string, error) {
	logger.Info("Service: Starting GetMinifigPricing", "item_id", itemID, "condition", condition, "user_id", user.ID)

	if itemID == "" {
		logger.Warn("Service: GetMinifigPricing called with empty item ID", "user_id", user.ID)
		return "", errors.New("item ID cannot be empty")
	}

	if condition == "" {
		condition = "N" // Default to new condition
	}

	logger.Info("Service: Getting Bricklink client for user", "item_id", itemID, "condition", condition, "user_id", user.ID)

	// Get user-specific Bricklink client from the client manager
	client, err := s.clientManager.GetClient(user)
	if err != nil {
		logger.Error("Service: Failed to get Bricklink client for user", "item_id", itemID, "condition", condition, "user_id", user.ID, "error", err)
		return "", err
	}

	logger.Info("Service: Calling user-specific Bricklink client to get pricing", "item_id", itemID, "condition", condition, "user_id", user.ID)

	// Use the user-specific Bricklink client to get the price guide with condition
	pricingData, err := client.GetPriceGuide(itemID, condition, 0)
	if err != nil {
		logger.Error("Service: Bricklink client returned error for pricing", "item_id", itemID, "condition", condition, "user_id", user.ID, "error", err)
		return "", err
	}

	logger.Info("Service: Successfully retrieved minifig pricing", "item_id", itemID, "condition", condition, "user_id", user.ID)

	return pricingData, nil
}

// GetMinifigParts retrieves parts data for a minifigure
func (s *PartialMinifigService) GetMinifigParts(itemID string, user *models.User) (string, error) {
	logger.Info("Service: Starting GetMinifigParts", "item_id", itemID, "user_id", user.ID)

	if itemID == "" {
		logger.Warn("Service: GetMinifigParts called with empty item ID", "user_id", user.ID)
		return "", errors.New("item ID cannot be empty")
	}

	logger.Info("Service: Getting Bricklink client for user", "item_id", itemID, "user_id", user.ID)

	// Get user-specific Bricklink client from the client manager
	client, err := s.clientManager.GetClient(user)
	if err != nil {
		logger.Error("Service: Failed to get Bricklink client for user", "item_id", itemID, "user_id", user.ID, "error", err)
		return "", err
	}

	logger.Info("Service: Calling user-specific Bricklink client to get parts", "item_id", itemID, "user_id", user.ID)

	// Use the user-specific Bricklink client to get the subset parts
	partsData, err := client.GetSubsetParts(itemID)
	if err != nil {
		logger.Error("Service: Bricklink client returned error for parts", "item_id", itemID, "user_id", user.ID, "error", err)
		return "", err
	}

	logger.Info("Service: Successfully retrieved minifig parts", "item_id", itemID, "user_id", user.ID)

	return partsData, nil
}

// GetPartPricing retrieves pricing data for a specific part with color
func (s *PartialMinifigService) GetPartPricing(itemType string, itemID string, colorID int, condition string, user *models.User) (string, error) {
	logger.Info("Service: Starting GetPartPricing", "item_type", itemType, "item_id", itemID, "color_id", colorID, "condition", condition, "user_id", user.ID)

	if itemID == "" {
		logger.Warn("Service: GetPartPricing called with empty item ID", "user_id", user.ID)
		return "", errors.New("item ID cannot be empty")
	}

	if condition == "" {
		condition = "N" // Default to new condition
	}

	if itemType == "" {
		itemType = "PART" // Default to part type
	}

	logger.Info("Service: Getting Bricklink client for user", "item_type", itemType, "item_id", itemID, "color_id", colorID, "condition", condition, "user_id", user.ID)

	// Get user-specific Bricklink client from the client manager
	client, err := s.clientManager.GetClient(user)
	if err != nil {
		logger.Error("Service: Failed to get Bricklink client for user", "item_type", itemType, "item_id", itemID, "color_id", colorID, "condition", condition, "user_id", user.ID, "error", err)
		return "", err
	}

	logger.Info("Service: Calling user-specific Bricklink client to get part pricing", "item_type", itemType, "item_id", itemID, "color_id", colorID, "condition", condition, "user_id", user.ID)

	// Use the user-specific Bricklink client to get the price guide with condition and color
	pricingData, err := client.GetPriceGuideByType(itemType, itemID, condition, colorID)
	if err != nil {
		logger.Error("Service: Bricklink client returned error for part pricing", "item_type", itemType, "item_id", itemID, "color_id", colorID, "condition", condition, "user_id", user.ID, "error", err)
		return "", err
	}

	logger.Info("Service: Successfully retrieved part pricing", "item_type", itemType, "item_id", itemID, "color_id", colorID, "condition", condition, "user_id", user.ID)

	return pricingData, nil
}

// GetPartPicture retrieves picture data for a specific part with color
func (s *PartialMinifigService) GetPartPicture(itemType string, itemID string, colorID int, user *models.User) (string, error) {
	logger.Info("Service: Starting GetPartPicture", "item_type", itemType, "item_id", itemID, "color_id", colorID, "user_id", user.ID)

	if itemID == "" {
		logger.Warn("Service: GetPartPicture called with empty item ID", "user_id", user.ID)
		return "", errors.New("item ID cannot be empty")
	}

	if itemType == "" {
		itemType = "PART" // Default to PART if not specified
	}

	logger.Info("Service: Getting Bricklink client for user", "item_type", itemType, "item_id", itemID, "color_id", colorID, "user_id", user.ID)

	// Get user-specific Bricklink client from the client manager
	client, err := s.clientManager.GetClient(user)
	if err != nil {
		logger.Error("Service: Failed to get Bricklink client for user", "item_type", itemType, "item_id", itemID, "color_id", colorID, "user_id", user.ID, "error", err)
		return "", err
	}

	logger.Info("Service: Calling user-specific Bricklink client to get part pictures", "item_type", itemType, "item_id", itemID, "color_id", colorID, "user_id", user.ID)

	// Use the user-specific Bricklink client to get the part pictures
	pictureData, err := client.GetPartPictures(itemType, itemID, colorID)
	if err != nil {
		logger.Error("Service: Bricklink client returned error for part pictures", "item_type", itemType, "item_id", itemID, "color_id", colorID, "user_id", user.ID, "error", err)
		return "", err
	}

	logger.Info("Service: Successfully retrieved part pictures", "item_type", itemType, "item_id", itemID, "color_id", colorID, "user_id", user.ID)

	return pictureData, nil
}

// CreatePartialMinifig creates a new partial minifig in a list for the specified user
func (s *PartialMinifigService) CreatePartialMinifig(listID int64, name, description string, user *models.User) (*PartialMinifigList, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	// TODO: Implement actual logic to save to database using repository with user validation
	newMinifig := PartialMinifig{
		ID:                   1, // TODO: Generate proper ID from database
		ReferenceID:          nil,
		PartialMinifigListID: listID,
		ItemID:               1, // TODO: Get actual item ID
		CreatedAt:            time.Now(),
		UpdatedAt:            nil,
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
func (s *PartialMinifigService) GetPartialMinifigByID(id int64, user *models.User) (*PartialMinifig, error) {
	// TODO: Implement actual logic to fetch from database using repository with user validation
	return &PartialMinifig{
		ID:                   id,
		ReferenceID:          nil,
		PartialMinifigListID: 1,
		ItemID:               1,
		CreatedAt:            time.Now(),
		UpdatedAt:            nil,
	}, nil
}

// UpdatePartialMinifig updates a partial minifig in a list for the specified user
func (s *PartialMinifigService) UpdatePartialMinifig(listID, itemID int64, name, description string, user *models.User) (*PartialMinifigList, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	// TODO: Implement actual logic to update in database using repository with user validation
	now := time.Now()
	updatedMinifig := PartialMinifig{
		ID:                   itemID,
		ReferenceID:          nil,
		PartialMinifigListID: listID,
		ItemID:               1, // TODO: Get actual item ID
		CreatedAt:            time.Now(),
		UpdatedAt:            &now,
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
func (s *PartialMinifigService) DeletePartialMinifig(listID, itemID int64, user *models.User) (*PartialMinifigList, error) {
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

// AddMinifigWithParts adds a new partial minifig with selected parts to a list
func (s *PartialMinifigService) AddMinifigWithParts(listID int64, minifigID string, minifigName string, referenceID *string, condition *string, notes *string, selectedParts []SelectedPart, user *models.User) (*PartialMinifigList, error) {
	logger.Info("Service: Starting AddMinifigWithParts", "list_id", listID, "minifig_id", minifigID, "parts_count", len(selectedParts), "user_id", user.ID)

	if minifigID == "" {
		logger.Warn("Service: AddMinifigWithParts called with empty minifig ID", "user_id", user.ID)
		return nil, errors.New("minifig ID cannot be empty")
	}

	if len(selectedParts) == 0 {
		logger.Warn("Service: AddMinifigWithParts called with no selected parts", "user_id", user.ID)
		return nil, errors.New("at least one part must be selected")
	}

	// Verify that the list exists and belongs to the user
	_, err := s.repo.GetPartialMinifigListByID(listID, user.ID)
	if err != nil {
		logger.Error("Service: Failed to get partial minifig list", "list_id", listID, "user_id", user.ID, "error", err)
		return nil, err
	}

	// Create the partial minifig entry
	partialMinifig, err := s.repo.CreatePartialMinifigWithParts(listID, minifigID, minifigName, referenceID, condition, notes, selectedParts, user.ID)
	if err != nil {
		logger.Error("Service: Failed to create partial minifig with parts", "list_id", listID, "minifig_id", minifigID, "user_id", user.ID, "error", err)
		return nil, err
	}

	logger.Info("Service: Successfully created partial minifig with parts", "list_id", listID, "minifig_id", minifigID, "partial_minifig_id", partialMinifig.ID, "user_id", user.ID)

	// Return the updated list
	return s.repo.GetPartialMinifigListByID(listID, user.ID)
}

// AddMinifigWithPartsNoReturn adds a new partial minifig with selected parts to a list without returning the updated list
func (s *PartialMinifigService) AddMinifigWithPartsNoReturn(listID int64, minifigID string, minifigName string, referenceID *string, condition *string, notes *string, selectedParts []SelectedPart, user *models.User) error {
	logger.Info("Service: Starting AddMinifigWithPartsNoReturn", "list_id", listID, "minifig_id", minifigID, "parts_count", len(selectedParts), "user_id", user.ID)

	if minifigID == "" {
		logger.Warn("Service: AddMinifigWithPartsNoReturn called with empty minifig ID", "user_id", user.ID)
		return errors.New("minifig ID cannot be empty")
	}

	if len(selectedParts) == 0 {
		logger.Warn("Service: AddMinifigWithPartsNoReturn called with no selected parts", "user_id", user.ID)
		return errors.New("at least one part must be selected")
	}

	// Verify that the list exists and belongs to the user
	_, err := s.repo.GetPartialMinifigListByID(listID, user.ID)
	if err != nil {
		logger.Error("Service: Failed to get partial minifig list", "list_id", listID, "user_id", user.ID, "error", err)
		return err
	}

	// Create the partial minifig entry
	partialMinifig, err := s.repo.CreatePartialMinifigWithParts(listID, minifigID, minifigName, referenceID, condition, notes, selectedParts, user.ID)
	if err != nil {
		logger.Error("Service: Failed to create partial minifig with parts", "list_id", listID, "minifig_id", minifigID, "user_id", user.ID, "error", err)
		return err
	}

	logger.Info("Service: Successfully created partial minifig with parts", "list_id", listID, "minifig_id", minifigID, "partial_minifig_id", partialMinifig.ID, "user_id", user.ID)

	// Don't return any data, just success
	return nil
}

// SelectedPart represents a part selected by the user to be added
type SelectedPart struct {
	PartNo    string  `json:"partNo"`
	PartName  string  `json:"partName"`
	Quantity  int     `json:"quantity"`
	ColorID   int     `json:"colorId"`
	ItemType  string  `json:"itemType"`
	Condition string  `json:"condition"`
	NewPrice  float64 `json:"newPrice,omitempty"`
	UsedPrice float64 `json:"usedPrice,omitempty"`
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
