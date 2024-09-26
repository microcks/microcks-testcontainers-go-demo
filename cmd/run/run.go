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
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

func Run(applicationProperties app.ApplicationProperties, applicationServices chan app.ApplicationServices) {
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
	if err != nil {
		fmt.Println("No bootstrap.servers specified for KafkaServer", err)
		os.Exit(1)
	}

	// Initialize your application
	fmt.Println("Starting Microcks TestContainers Go Demo application...")
	fmt.Printf("  Connecting to Kafka server: %s \n", kafkaServer)
	fmt.Printf("  Connecting to Microcks Pastries: %s \n", applicationProperties.PastriesBaseUrl)

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
	pastryAPIClient := client.NewPastryAPIClient(strings.Replace(applicationProperties.PastriesBaseUrl, " ", "+", -1))
	orderPublisher := service.NewOrderEventPublisher(kafkaProducer, applicationProperties.OrderEventsCreatedTopic)
	orderService := service.NewOrderService(pastryAPIClient, orderPublisher)
	orderController := controller.NewOrderController(orderService)

	// Initialize and start the event listener.
	orderListener := service.NewOrderEventListener(kafkaConsumer, applicationProperties.OrderEventsReviewedTopic, orderService)
	err = orderListener.Listen()
	if err != nil {
		fmt.Println("Error while starting consuming orders reviews", err)
		os.Exit(1)
	}

	// Provide ApplicationServices to exernal caller.
	services := app.ApplicationServices{
		OrderService: orderService,
	}
	applicationServices <- services // service to channel

	// Define your HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/api/orders", orderController.CreateOrder)

	// Start your HTTP server
	fmt.Println("Microcks TestContainers Go Demo application is listening on localhost:9000")
	fmt.Println("")

	//go http.ListenAndServe(":9000", nil)
	server := &http.Server{Addr: ":9000", Handler: mux}
	err = server.ListenAndServe()

	<-close
	orderListener.Stop()
	server.Shutdown(context.Background())
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
