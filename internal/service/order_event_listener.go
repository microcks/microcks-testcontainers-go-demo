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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/model"
	"github.com/pkg/errors"
)

// OrderEventListener is the service interface for listening order events.
type OrderEventListener interface {
	// Listen starts listening to the Kafka topic and returns a channel that will be closed when listening stops
	Listen(ctx context.Context) (<-chan struct{}, error)
	// Stop gracefully stops listening to the Kafka topic
	Stop()
}

type orderEventListener struct {
	kafkaConsumer *kafka.Consumer
	kafkaTopic    string
	orderService  OrderService
	done          chan struct{}
	wg            sync.WaitGroup
	mu            sync.Mutex
	isRunning     bool
}

// OrderEventListenerConfig Configuration options for the listener.
type OrderEventListenerConfig struct {
	MaxRetries     int
	RetryBackoff   time.Duration
	CommitInterval time.Duration
	MessageTimeout time.Duration
}

// Default configuration values.
var defaultConfig = OrderEventListenerConfig{
	MaxRetries:     3,
	RetryBackoff:   time.Second * 2,
	CommitInterval: time.Second * 5,
	MessageTimeout: time.Second * 30,
}

func NewOrderEventListener(kafkaConsumer *kafka.Consumer, kafkaTopic string, orderService OrderService) OrderEventListener {
	return &orderEventListener{
		kafkaConsumer: kafkaConsumer,
		kafkaTopic:    kafkaTopic,
		orderService:  orderService,
		done:          make(chan struct{}),
	}
}

func (l *orderEventListener) Listen(ctx context.Context) (<-chan struct{}, error) {
	l.mu.Lock()
	if l.isRunning {
		l.mu.Unlock()
		return nil, errors.New("listener is already running")
	}
	l.isRunning = true
	l.mu.Unlock()

	err := l.kafkaConsumer.Subscribe(l.kafkaTopic, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", l.kafkaTopic, err)
	}

	finished := make(chan struct{})
	l.wg.Add(1)

	// Start the message processing goroutine
	go func() {
		defer func() {
			l.mu.Lock()
			l.isRunning = false
			l.mu.Unlock()
			l.wg.Done()
			close(finished)
			log.Println("Kafka listener stopped")
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("Context cancelled, stopping listener")
				return
			case <-l.done:
				log.Println("Received stop signal, stopping listener")
				return
			default:
				message, err := l.kafkaConsumer.ReadMessage(defaultConfig.MessageTimeout)
				if err != nil {
					if err.(kafka.Error).Code() == kafka.ErrTimedOut {
						continue
					}
					log.Printf("Error reading message: %v", err)
					continue
				}

				if err := l.processMessage(message); err != nil {
					log.Printf("Error processing message: %v", err)
					// Implement retry logic with backoff
					if retryErr := l.retryProcessMessage(message); retryErr != nil {
						log.Printf("Failed to process message after retries: %v", retryErr)
						// Consider implementing dead letter queue here
					}
				}

				// Commit the message offset
				if _, err := l.kafkaConsumer.CommitMessage(message); err != nil {
					log.Printf("Failed to commit message: %v", err)
				}
			}
		}
	}()

	return finished, nil
}

func (l *orderEventListener) processMessage(message *kafka.Message) error {
	if message == nil {
		return errors.New("received nil message")
	}

	log.Printf("Processing message from topic %s [%d] at offset %v: key = %s\n",
		*message.TopicPartition.Topic, message.TopicPartition.Partition,
		message.TopicPartition.Offset, string(message.Key))

	var orderEvent model.OrderEvent
	if err := json.Unmarshal(message.Value, &orderEvent); err != nil {
		return fmt.Errorf("failed to unmarshal message value: %w", err)
	}

	order := l.orderService.UpdateReviewedOrder(&orderEvent)
	log.Printf("Order '%s' has been updated after review", order.ID)
	return nil
}

func (l *orderEventListener) retryProcessMessage(message *kafka.Message) error {
	var lastErr error
	for i := range defaultConfig.MaxRetries {
		time.Sleep(defaultConfig.RetryBackoff * time.Duration(i+1))

		if err := l.processMessage(message); err != nil {
			lastErr = err
			log.Printf("Retry %d/%d failed: %v", i+1, defaultConfig.MaxRetries, err)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed after %d retries, last error: %w", defaultConfig.MaxRetries, lastErr)
}

func (l *orderEventListener) Stop() {
	l.mu.Lock()
	if !l.isRunning {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()

	close(l.done)
	l.wg.Wait()

	if l.kafkaConsumer == nil {
		return
	}

	if l.kafkaConsumer.IsClosed() {
		return
	}

	if err := l.kafkaConsumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}
}
