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
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	run "github.com/microcks/microcks-testcontainers-go-demo/cmd/run"
	internal "github.com/microcks/microcks-testcontainers-go-demo/internal"
)

var close chan bool

func main() {
	pastryAPIBaseURL, present := os.LookupEnv("PASTRY_API_URL")
	if !present {
		pastryAPIBaseURL = "http://localhost:9090/rest/API+Pastries/0.0.1"
	}
	kafkaBootstrapURL, present := os.LookupEnv("KAFKA_BOOTSTRAP_URL")
	if !present {
		kafkaBootstrapURL = "localhost:9092"
	}

	applicationProperties := &internal.ApplicationProperties{
		PastriesBaseUrl:         pastryAPIBaseURL,
		OrderEventsCreatedTopic: "orders-created",
		//OrderEventsReviewedTopic: "orders-reviewed",
		OrderEventsReviewedTopic: "OrderEventsAPI-0.1.0-orders-reviewed",
		KafkaConfigMap: &kafka.ConfigMap{
			"bootstrap.servers": kafkaBootstrapURL,
			"group.id":          "order-service",
			"auto.offset.reset": "latest",
		},
	}

	appServicesChan := make(chan internal.ApplicationServices)
	go run.Run(*applicationProperties, appServicesChan)
	_ = <-appServicesChan

	// Setup signal hooks.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	close = make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Println("Caught signal " + sig.String())
		close <- true
	}()

	<-close
	fmt.Println("Exiting Microcks TestContainers Go Demo application main.")
}
