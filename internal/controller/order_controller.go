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

type OrderController struct {
	service service.OrderService
}

type unavailableProduct struct {
	ProductName string `json:"productName"`
	Details     string `json:"details"`
}

func NewOrderController(service service.OrderService) *OrderController {
	return &OrderController{
		service: service,
	}
}

func (oc *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
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
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&unavailableProduct{
				ProductName: unavailableErr.Error(),
				Details:     "Pastry " + unavailableErr.Error() + " is not available",
			})
		} else {
			os.Exit(1)
		}
	} else {
		// Serialize order to JSON and write response
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}
