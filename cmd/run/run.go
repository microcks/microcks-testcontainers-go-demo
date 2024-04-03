package run

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/microcks/microcks-testcontainers-go-demo/internal/client"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/controller"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/service"
)

var close chan bool

func Run(pastryAPIBaseURL string) {
	// Setup signal hooks.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	close = make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Println("Caught signal " + sig.String())
		close <- true
	}()

	// Initialize your application
	fmt.Println("Starting Microcks TestContainers Go Demo application...")

	pastryAPIClient := client.NewPastryAPIClient(pastryAPIBaseURL)
	orderService := service.NewOrderService(pastryAPIClient)
	orderController := controller.NewOrderController(orderService)

	// Define your HTTP routes
	http.HandleFunc("/", handler)

	http.HandleFunc("/api/orders", orderController.CreateOrder)

	// Start your HTTP server
	http.ListenAndServe(":9000", nil)

	<-close
	fmt.Println("Exiting Microcks TestContainers Go Demo application.")
}

func Close() {
	close <- true
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Your HTTP request handler logic goes here
	fmt.Fprintf(w, "Hello, World!")
}
