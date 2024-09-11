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
package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/microcks/microcks-testcontainers-go-demo/internal/model"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/service"
)

type OrderController interface {
	CreateOrder(w http.ResponseWriter, r *http.Request)
}

type orderController struct {
	service service.OrderService
}

type unavailableProduct struct {
	ProductName string `json:"productName"`
	Details     string `json:"details"`
}

func NewOrderController(service service.OrderService) OrderController {
	return &orderController{
		service: service,
	}
}

func (oc *orderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Read OrderInfo from body.
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	info := model.OrderInfo{}
	json.Unmarshal([]byte(string(body)), &info)

	// Place a new order.
	order, err := oc.service.PlaceOrder(&info)
	if err != nil {
		// Manage unavialble product.
		var unavailableErr *service.UnavailablePastryError
		if errors.As(err, &unavailableErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(&unavailableProduct{
				ProductName: unavailableErr.Error(),
				Details:     "Pastry " + unavailableErr.Error() + " is not available",
			})
		} else {
			os.Exit(1)
		}
	} else {
		// Serialize order to JSON and write response.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(order)
	}
}
