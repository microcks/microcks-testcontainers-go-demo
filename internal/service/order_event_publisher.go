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

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/model"
)

// OrderEventPublisher is the service interface for publishing order events.
type OrderEventPublisher interface {
	// Publish a new order event.
	PublishOrderEvent(event *model.OrderEvent) (*model.OrderEvent, error)
}

type orderEventPublisher struct {
	kafkaProducer *kafka.Producer
	kafkaTopic    string
}

func NewOrderEventPublisher(kafkaProducer *kafka.Producer, kafkaTopic string) OrderEventPublisher {
	return &orderEventPublisher{
		kafkaProducer: kafkaProducer,
		kafkaTopic:    kafkaTopic,
	}
}

// PublishOrderEvent
func (oep *orderEventPublisher) PublishOrderEvent(event *model.OrderEvent) (*model.OrderEvent, error) {
	// Serailize OrderEvent in JSON.
	eventJson, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	// Publish on Kafka topic.
	oep.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &oep.kafkaTopic, Partition: kafka.PartitionAny},
		Value:          eventJson,
	}, nil)

	numMsg := oep.kafkaProducer.Flush(50)
	fmt.Println("Send", numMsg, "Kafka messages")

	return event, nil
}
