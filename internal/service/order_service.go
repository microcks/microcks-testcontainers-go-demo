package service

import (
	"github.com/google/uuid"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/client"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/model"
)

// UnavailablePastryError is raised by OrderService when a pastry is not available in inventory
type UnavailablePastryError struct {
	product string
}

func (e *UnavailablePastryError) Error() string {
	return e.product
}

// OrderService is the service interface for managing orders.
type OrderService interface {
	// Place a new order if valid. May return an UnavailableProductError.
	PlaceOrder(info *model.OrderInfo) (*model.Order, error)
}

type orderService struct {
	pastryAPI client.PastryAPI
}

func NewOrderService(pastryAPI client.PastryAPI) OrderService {
	return &orderService{
		pastryAPI: pastryAPI,
	}
}

// PlaceOrder allows checking inventory and save and order if products are available.
func (os *orderService) PlaceOrder(info *model.OrderInfo) (*model.Order, error) {
	// Check availability of pastries.
	for i := 0; i < len(info.ProductQuantities); i++ {
		var productQuantity = info.ProductQuantities[i]
		pastry, err := os.pastryAPI.GetPastry(productQuantity.ProductName)
		if (err != nil) || (pastry.Status != "available") {
			return nil, &UnavailablePastryError{product: productQuantity.ProductName}
		}
	}

	return &model.Order{
		OrderInfo: *info,
		Id:        uuid.New().String(),
		Status:    model.CREATED,
	}, nil
}
