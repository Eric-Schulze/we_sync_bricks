package orders

import ()

type Orders interface {
	GetOrders() (string, error)
}

type OrdersService struct {
	orders Orders
}

func NewOrdersClient(c Orders) (*OrdersService, error) {
	return &OrdersService{
		orders: c,
	}, nil
}

func (service *OrdersService) GetOrders() (string, error){
	var orders, err = service.orders.GetOrders()
	if err != nil {
		return "", err
	}

	return orders, nil
}

