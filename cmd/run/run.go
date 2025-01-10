// Copyright The Microcks Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package run

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	app "github.com/microcks/microcks-testcontainers-go-demo/internal"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/client"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/controller"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/service"
)

const (
	// DefaultApplicationPort represents the default port used for exposiing the application.
	DefaultApplicationPort = 9000
)

// Application is the application interface for starting/stopping it.
type Application interface {
	// Start this demo application using properties.
	Start() error
	// Stop this demo application.
	Stop() error
}

type App struct {
	appServices     app.ApplicationServices
	kafkaConsumer   *kafka.Consumer
	kafkaProducer   *kafka.Producer
	pastryAPIClient client.PastryAPI
	orderListener   service.OrderEventListener
	server          *http.Server

	AppService app.ApplicationServices
}

func NewApplication(applicationProperties *app.ApplicationProperties) *App {
	// Initialize kafka server.
	kafkaServer, err := applicationProperties.KafkaConfigMap.Get("bootstrap.servers", "unknown")
	if err != nil {
		fmt.Println("No bootstrap.servers specified for KafkaServer", err)
		os.Exit(1)
	}

	// Initialize your application
	fmt.Println("Starting Microcks TestContainers Go Demo application...")
	fmt.Printf("  Connecting to Kafka server: %s \n", kafkaServer)
	fmt.Printf("  Connecting to Microcks Pastries: %s \n", applicationProperties.PastriesBaseURL)

	// Prepare Kafka components we need.

	kafkaConsumer, err := kafka.NewConsumer(applicationProperties.KafkaConfigMap)
	if err != nil {
		fmt.Println("Error while connecting to Kafka broker", err)
		os.Exit(1)
	}
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServer,
	})
	if err != nil {
		fmt.Println("Error while connecting to Kafka broker", err)
		os.Exit(1)
	}

	// Prepare our own components and services.
	pastryAPIClient := client.NewPastryAPIClient(strings.ReplaceAll(applicationProperties.PastriesBaseURL, " ", "+"))
	orderPublisher := service.NewOrderEventPublisher(kafkaProducer, applicationProperties.OrderEventsCreatedTopic)
	orderService := service.NewOrderService(pastryAPIClient, orderPublisher)
	orderController := controller.NewOrderController(orderService)

	// Initialize and start the event listener.
	orderListener := service.NewOrderEventListener(kafkaConsumer, applicationProperties.OrderEventsReviewedTopic, orderService)

	// Define your HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/api/orders", orderController.CreateOrder)

	// Start your HTTP server
	fmt.Println("Microcks TestContainers Go Demo application is listening on localhost:9000")
	fmt.Println("")

	// go http.ListenAndServe(":9000", nil)
	server := &http.Server{Addr: ":9000", Handler: mux}

	return &App{
		appServices:     app.ApplicationServices{},
		kafkaConsumer:   kafkaConsumer,
		kafkaProducer:   kafkaProducer,
		pastryAPIClient: pastryAPIClient,
		orderListener:   orderListener,
		AppService: app.ApplicationServices{
			OrderService: orderService,
		},
		server: server,
	}
}

func (a *App) Start() error {
	_, err := a.orderListener.Listen(context.Background())
	if err != nil {
		fmt.Println("Error while starting consuming orders reviews", err)
		os.Exit(1)
	}

	return a.server.ListenAndServe()
}

func (a *App) Stop() error {
	fmt.Println("Stopping Microcks TestContainers Go Demo application...")
	a.orderListener.Stop()

	fmt.Println("Stopping Kafka producer & consumer...")
	if a.kafkaConsumer != nil && !a.kafkaConsumer.IsClosed() {
		if err := a.kafkaConsumer.Close(); err != nil {
			return err
		}
	}

	if a.kafkaProducer != nil && !a.kafkaProducer.IsClosed() {
		a.kafkaProducer.Close()
	}

	// Stop HTTP server
	return a.server.Shutdown(context.Background())
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Your HTTP request handler logic goes here
	// fmt.Fprintf(w, "Hello, World!")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"Hello": "World!"}`))
}
