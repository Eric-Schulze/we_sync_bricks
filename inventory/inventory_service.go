package inventory

import ()

type InventoryClient interface {
	GetInventory() (string, error)
}

type InventoryService struct {
	client InventoryClient
}

func NewInventoryService(c InventoryClient) *InventoryService {
	return &InventoryService{
		client: c,
	}
}

func (service *InventoryService) GetInventory() (string, error) {
	var inventory, err = service.client.GetInventory()
	if err != nil {
		return "", err
	}

	return inventory, nil
}
