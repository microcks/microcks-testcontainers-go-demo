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
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/model"
)

// OrderEventListener is the service interface for listening order events.
type OrderEventListener interface {
	// Start listening to the Kafka topic.
	Listen() error
	// Stop listening to the Kafka topic.
	Stop()
}

type orderEventListener struct {
	kafkaConsumer   *kafka.Consumer
	kafkaTopic      string
	orderService    OrderService
	listenerHandler chan struct{}
}

func NewOrderEventListener(kafkaConsumer *kafka.Consumer, kafkaTopic string, orderService OrderService) OrderEventListener {
	return &orderEventListener{
		kafkaConsumer: kafkaConsumer,
		kafkaTopic:    kafkaTopic,
		orderService:  orderService,
	}
}

func (oel *orderEventListener) Listen() error {
	err := oel.kafkaConsumer.Subscribe(oel.kafkaTopic, nil)
	if err != nil {
		return err
	}

	// Handle incoming messages.
	stopChannel := make(chan struct{})
	oel.listenerHandler = stopChannel
	go func() {
		// This is no longer needed since we're now closing the channel in Close()
		// and handling the <-stopChannel case that ends the gorountine.
		//defer close(handler)
		for {
			select {
			case <-stopChannel:
				fmt.Println("Stopping Kafka listener goroutine...")
				return
			default:
				message, err := oel.kafkaConsumer.ReadMessage(100 * time.Millisecond)
				if err != nil {
					// Errors are informational and automatically handled by the consumer
					continue
				}
				fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
					*message.TopicPartition.Topic, string(message.Key), string(message.Value))

				// Transform message value in OrderEvent.
				orderEvent := model.OrderEvent{}
				err = json.Unmarshal(message.Value, &orderEvent)
				if err != nil {
					fmt.Println("Error while deserializing message value", err)
				} else {
					order := oel.orderService.UpdateReviewedOrder(&orderEvent)
					fmt.Println("Order '" + order.Id + "' has been updated after review")
				}
			}
		}
	}()

	return nil
}

func (oel *orderEventListener) Stop() {
	fmt.Println("Stopping Kafka consumer...")
	close(oel.listenerHandler)
	err := oel.kafkaConsumer.Close()
	if err != nil {
		fmt.Println("Got error while closing consumer: " + err.Error())
	}
}
