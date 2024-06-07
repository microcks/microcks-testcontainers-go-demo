/*
 * Copyright The Microcks Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package service

import (
	"time"

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
	// Retrive an exsiting order.
	GetOrder(id string) *model.Order
	// Update an order that has been reviewed.
	UpdateReviewedOrder(event *model.OrderEvent) *model.Order
}

type orderService struct {
	pastryAPI           client.PastryAPI
	orderEventPublisher OrderEventPublisher
	ordersRepository    map[string]*model.Order
}

func NewOrderService(pastryAPI client.PastryAPI, orderEventPublisher OrderEventPublisher) OrderService {
	return &orderService{
		pastryAPI:           pastryAPI,
		orderEventPublisher: orderEventPublisher,
		ordersRepository:    make(map[string]*model.Order),
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

	// Everything is available! Create a new order.
	order := &model.Order{
		OrderInfo: *info,
		Id:        uuid.New().String(),
		Status:    model.CREATED,
	}

	// Persist and publish creation event.
	orderCreated := &model.OrderEvent{
		Timestamp:    1000 * time.Now().Unix(),
		Order:        *order,
		ChangeReason: "creation",
	}
	os.orderEventPublisher.PublishOrderEvent(orderCreated)
	os.ordersRepository[order.Id] = order

	return order, nil
}

// GetOrder allows retreiving an order by its id. May retur nil if unknown.
func (os *orderService) GetOrder(id string) *model.Order {
	return os.ordersRepository[id]
}

// UpdateReviewedOrder allows peristing an order review.
func (os *orderService) UpdateReviewedOrder(event *model.OrderEvent) *model.Order {
	os.ordersRepository[event.Order.Id] = &event.Order
	return &event.Order
}
