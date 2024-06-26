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
package run

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

var close chan bool

func Run(applicationProperties app.ApplicationProperties) {
	// Setup signal hooks.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	close = make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Println("Caught signal " + sig.String())
		close <- true
	}()

	kafkaServer, err := applicationProperties.KafkaConfigMap.Get("bootstrap.servers", "unknown")
	fmt.Printf("%s", kafkaServer)

	// Initialize your application
	fmt.Println("Starting Microcks TestContainers Go Demo application...")
	fmt.Println("  Connecting to Kafka server")

	// Prepare Kafka components we need.
	kafkaConsumer, err := kafka.NewConsumer(applicationProperties.KafkaConfigMap)
	if err != nil {
		fmt.Println("Error while connecting to Kafka broker", err)
		os.Exit(1)
	}
	kafkaProducer, err := kafka.NewProducer(applicationProperties.KafkaConfigMap)
	if err != nil {
		fmt.Println("Error while connecting to Kafka broker", err)
		os.Exit(1)
	}

	// Prepare our own components and services.
	pastryAPIClient := client.NewPastryAPIClient(applicationProperties.PastriesBaseUrl)
	orderPublisher := service.NewOrderEventPublisher(kafkaProducer, applicationProperties.OrderEventsCreatedTopic)
	orderService := service.NewOrderService(pastryAPIClient, orderPublisher)
	orderController := controller.NewOrderController(orderService)

	// Initialize and start the event listener.
	orderListener := service.NewOrderEventListener(kafkaConsumer, applicationProperties.OrderEventsReviewedTopic, orderService)
	err = orderListener.Listen()
	if err != nil {
		fmt.Println("Error while starting consumling orders reviews", err)
		os.Exit(1)
	}

	// Define your HTTP routes
	http.HandleFunc("/", handler)

	http.HandleFunc("/api/orders", orderController.CreateOrder)

	// Start your HTTP server
	http.ListenAndServe(":9000", nil)

	<-close
	orderListener.Stop()
	fmt.Println("Exiting Microcks TestContainers Go Demo application.")
}

func Close() {
	close <- true
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Your HTTP request handler logic goes here
	//fmt.Fprintf(w, "Hello, World!")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(`{"Hello": "World!"}`))
}
