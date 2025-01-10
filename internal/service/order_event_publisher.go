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
	// Listen to all the events on the default events channel
	go func() {
		for e := range kafkaProducer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				// The message delivery report, indicating success or permanent failure after retries have been exhausted.
				// Application level retries won't help since the client is already configured to do that.
				m := ev
				if m.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
				} else {
					fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
				}
			case kafka.Error:
				// Generic client instance-level errors, such as broker connection failures, authentication issues, etc.
				// These errors should generally be considered informational as the underlying client will automatically try to
				// recover from any errors encountered, the application does not need to take action on them.
				fmt.Printf("Error: %v\n", ev)
			default:
				fmt.Printf("Ignored event: %s\n", ev)
			}
		}
	}()

	return &orderEventPublisher{
		kafkaProducer: kafkaProducer,
		kafkaTopic:    kafkaTopic,
	}
}

// PublishOrderEvent.
func (oep *orderEventPublisher) PublishOrderEvent(event *model.OrderEvent) (*model.OrderEvent, error) {
	// Serailize OrderEvent in JSON.
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	// Publish on Kafka topic.
	err = oep.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &oep.kafkaTopic, Partition: kafka.PartitionAny},
		Value:          eventJSON,
	}, nil)
	if err != nil {
		return nil, err
	}

	numMsg := oep.kafkaProducer.Flush(750)
	fmt.Println("Send", numMsg, "Kafka messages")

	return event, nil
}
