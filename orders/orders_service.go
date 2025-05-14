package orders

import ()

type OrdersClient interface {
	GetOrders() (string, error)
}

type OrdersService struct {
	client OrdersClient
}

func NewOrdersService(c OrdersClient) *OrdersService {
	return &OrdersService{
		client: c,
	}
}

func (service *OrdersService) GetOrders() (string, error){
	var orders, err = service.client.GetOrders()
	if err != nil {
		return "", err
	}

	return orders, nil
}
