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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	server "github.com/microcks/microcks-testcontainers-go-demo/cmd/run"
	"github.com/microcks/microcks-testcontainers-go-demo/internal"
)

const (
	defaultPastryAPIURL   = "http://localhost:9090/rest/API+Pastries/0.0.1"
	defaultKafkaBootstrap = "localhost:9092"
	shutdownTimeout       = 15 * time.Second
	defaultOrdersTopic    = "orders-created"
	defaultReviewedTopic  = "OrderEventsAPI-0.1.0-orders-reviewed"
)

// Config holds all application configuration.
type Config struct {
	PastryAPIURL   string
	KafkaBootstrap string
	OrdersTopic    string
	ReviewedTopic  string
}

// loadConfig loads configuration from environment variables with defaults.
func loadConfig() *Config {
	return &Config{
		PastryAPIURL:   getEnv("PASTRY_API_URL", defaultPastryAPIURL),
		KafkaBootstrap: getEnv("KAFKA_BOOTSTRAP_URL", defaultKafkaBootstrap),
		OrdersTopic:    getEnv("ORDERS_TOPIC", defaultOrdersTopic),
		ReviewedTopic:  getEnv("REVIEWED_TOPIC", defaultReviewedTopic),
	}
}

// getEnv retrieves an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	// Setup logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	// Load configuration
	config := loadConfig()

	// Create application properties
	appProps := &internal.ApplicationProperties{
		PastriesBaseURL:          config.PastryAPIURL,
		OrderEventsCreatedTopic:  config.OrdersTopic,
		OrderEventsReviewedTopic: config.ReviewedTopic,
		KafkaConfigMap: &kafka.ConfigMap{
			"bootstrap.servers": config.KafkaBootstrap,
			"group.id":          "order-service",
			"auto.offset.reset": "latest",
		},
	}

	// Create application
	app := server.NewApplication(appProps)

	// Setup error channel
	errChan := make(chan error, 1)

	// Start application in a goroutine
	go func() {
		log.Printf("Starting application with Pastry API URL: %s, Kafka bootstrap: %s",
			config.PastryAPIURL, config.KafkaBootstrap)
		if err := app.Start(); err != nil {
			errChan <- fmt.Errorf("failed to start application: %w", err)
		}
	}()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal or error
	var shutdownErr error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case err := <-errChan:
		log.Printf("Received error: %v", err)
		shutdownErr = err
	}

	// Graceful shutdown
	log.Println("Starting graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Stop the application
	shutdownChan := make(chan struct{})
	go func() {
		if err := app.Stop(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		close(shutdownChan)
	}()

	// Wait for shutdown or timeout
	select {
	case <-shutdownCtx.Done():
		return fmt.Errorf("shutdown timed out: %v", shutdownCtx.Err())
	case <-shutdownChan:
		log.Println("Graceful shutdown completed")
	}

	return shutdownErr
}
