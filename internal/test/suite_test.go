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
package test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	server "github.com/microcks/microcks-testcontainers-go-demo/cmd/run"
	app "github.com/microcks/microcks-testcontainers-go-demo/internal"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/model"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"microcks.io/go-client"
	"microcks.io/testcontainers-go/ensemble"
	kafkaCon "microcks.io/testcontainers-go/ensemble/async/connection/kafka"
	kafkaTC "microcks.io/testcontainers-go/kafka"
)

type BaseSuite struct {
	suite.Suite
	kafkaContainer   *kafkaTC.KafkaContainer
	microcksEnsemble *ensemble.MicrocksContainersEnsemble
	app              *server.App
}

func TestBaseSuite(t *testing.T) {
	suite.Run(t, new(BaseSuite))
}

func (s *BaseSuite) SetupTest() {
}

func (s *BaseSuite) TearDownTest() {
}

func (s *BaseSuite) SetupSuite() {
	ctx := context.Background()
	net, err := network.New(ctx)
	s.Require().NoError(err)

	kafkaContainer, err := kafkaTC.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		network.WithNetwork([]string{"kafka"}, net),

		testcontainers.WithEnv(map[string]string{
			"KAFKA_LISTENERS":                      "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094,TC://0.0.0.0:19092",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "PLAINTEXT:PLAINTEXT,BROKER:PLAINTEXT,CONTROLLER:PLAINTEXT,TC:PLAINTEXT",
			"KAFKA_ADVERTISED_LISTENERS":           "PLAINTEXT://%s:%d,BROKER://%s:9092,TC://kafka:19092",
		}),
	)
	s.Require().NoError(err)

	s.kafkaContainer = kafkaContainer

	// Configure and startup a new MicrocksContainersEnsemble.
	microcksEnsemble, err := ensemble.RunContainers(ctx,
		ensemble.WithMainArtifact("../../testdata/order-service-openapi.yaml"),
		ensemble.WithMainArtifact("../../testdata/order-events-asyncapi.yaml"),
		ensemble.WithMainArtifact("../../testdata/apipastries-openapi.yaml"),
		ensemble.WithSecondaryArtifact("../../testdata/order-service-postman-collection.json"),
		ensemble.WithSecondaryArtifact("../../testdata/apipastries-postman-collection.json"),
		ensemble.WithPostman(),
		ensemble.WithAsyncFeature(),
		ensemble.WithNetwork(net),
		ensemble.WithHostAccessPorts([]int{server.DefaultApplicationPort}),
		ensemble.WithKafkaConnection(kafkaCon.Connection{
			BootstrapServers: "kafka:19092",
		}),
	)
	s.Require().NoError(err)

	s.microcksEnsemble = microcksEnsemble

	brokerURL, err := kafkaContainer.Brokers(ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(brokerURL)

	fmt.Println("BrokerURL is " + brokerURL[0])

	// Configure and start the application.
	baseAPIURL, err := microcksEnsemble.GetMicrocksContainer().RestMockEndpoint(ctx, "API Pastries", "0.0.1")
	s.Require().NoError(err)

	reviewedTopic := microcksEnsemble.GetAsyncMinionContainer().KafkaMockTopic("Order Events API", "0.1.0", "PUBLISH orders-reviewed")

	applicationProperties := &app.ApplicationProperties{
		PastriesBaseURL:          baseAPIURL,
		OrderEventsCreatedTopic:  "orders-created",
		OrderEventsReviewedTopic: reviewedTopic,
		KafkaConfigMap: &kafka.ConfigMap{
			"bootstrap.servers": brokerURL[0],
			"group.id":          "order-service",
			"auto.offset.reset": "latest",
		},
	}

	appRun := server.NewApplication(applicationProperties)
	go func() {
		_ = appRun.Start()
	}()

	s.app = appRun
}

func (s *BaseSuite) TearDownSuite() {
	ctx := context.Background()

	err := s.app.Stop()
	s.Require().NoError(err)

	err = s.kafkaContainer.Terminate(ctx)
	s.Require().NoError(err)

	err = s.microcksEnsemble.Terminate(ctx)
	s.Require().NoError(err)
}

func waitFor(timeout time.Duration, fn func() error) error {
	start := time.Now()
	for {
		err := fn()
		if err == nil {
			return nil
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("timeout after %v: %w", timeout, err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (s *BaseSuite) TestOpenAPIContract() {
	// Test code goes here which can leverage the context
	// Prepare a Microcks Test.
	testRequest := client.TestRequest{
		ServiceId:    "Order Service API:0.1.0",
		RunnerType:   client.TestRunnerTypeOPENAPISCHEMA,
		TestEndpoint: fmt.Sprintf("http://host.testcontainers.internal:%d/api", server.DefaultApplicationPort),
		Timeout:      2000,
	}
	testResult, err := s.microcksEnsemble.GetMicrocksContainer().TestEndpoint(context.Background(), &testRequest)
	s.Require().NoError(err)

	s.T().Logf("Test Result success is %t", testResult.Success)

	// Log TestResult raw structure.
	j, err := json.Marshal(testResult)
	s.Require().NoError(err)
	s.T().Log(string(j))

	s.True(testResult.Success)
	s.Equal(1, len(*testResult.TestCaseResults)) //nolint:testifylint
}

func (s *BaseSuite) TestPostmanCollectionContract() {
	ctx := context.Background()
	// Test code goes here which can leverage the context
	// Prepare a Microcks Test.
	testRequest := client.TestRequest{
		ServiceId:    "Order Service API:0.1.0",
		RunnerType:   client.TestRunnerTypePOSTMAN,
		TestEndpoint: fmt.Sprintf("http://host.testcontainers.internal:%d/api", server.DefaultApplicationPort),
		Timeout:      2000,
	}
	testResult, err := s.microcksEnsemble.GetMicrocksContainer().TestEndpoint(ctx, &testRequest)
	s.Require().NoError(err)

	s.T().Logf("Test Result success is %t", testResult.Success)

	// Log TestResult raw structure.
	j, err := json.Marshal(testResult)
	s.Require().NoError(err)
	s.T().Log(string(j))

	s.True(testResult.Success)
	s.Equal(1, len(*testResult.TestCaseResults)) //nolint:testifylint
}

func (s *BaseSuite) TestOpenAPIContractAndBusinessConformance() {
	ctx := context.Background()
	// Test code goes here which can leverage the context
	// Prepare a Microcks Test.
	testRequest := client.TestRequest{
		ServiceId:    "Order Service API:0.1.0",
		RunnerType:   client.TestRunnerTypeOPENAPISCHEMA,
		TestEndpoint: fmt.Sprintf("http://host.testcontainers.internal:%d/api", server.DefaultApplicationPort),
		Timeout:      2000,
	}
	testResult, err := s.microcksEnsemble.GetMicrocksContainer().TestEndpoint(ctx, &testRequest)
	s.Require().NoError(err)

	s.T().Logf("Test Result success is %t", testResult.Success)

	// Log TestResult raw structure.
	j, err := json.Marshal(testResult)
	s.Require().NoError(err)
	s.T().Log(string(j))

	s.True(testResult.Success)
	s.Equal(1, len(*testResult.TestCaseResults)) //nolint:testifylint

	// You may also check business conformance.
	pairs, err := s.microcksEnsemble.GetMicrocksContainer().MessagesForTestCase(ctx, testResult, "POST /orders")
	s.Require().NoError(err)

	for _, pair := range *pairs {
		s.T().Logf("Got a responseBody %s", *pair.Response.Content)
	}
}

func (s *BaseSuite) TestOrderEventIsPublishedWhenOrderIsCreated() {
	ctx := context.Background()
	// Prepare a Microcks Tests.
	testRequest := client.TestRequest{
		ServiceId:          "Order Events API:0.1.0",
		RunnerType:         client.TestRunnerTypeASYNCAPISCHEMA,
		TestEndpoint:       "kafka://kafka:19092/orders-created",
		Timeout:            2000,
		FilteredOperations: &[]string{"SUBSCRIBE orders-created"},
	}

	// Prepare an application Order.
	info := model.OrderInfo{
		CustomerID: "123-456-789",
		ProductQuantities: []model.ProductQuantity{
			{
				ProductName: "Millefeuille",
				Quantity:    1,
			},
			{
				ProductName: "Eclair Cafe",
				Quantity:    1,
			},
		},
		TotalPrice: 8.4,
	}

	testResultChan := make(chan *client.TestResult)
	go func() {
		err := s.microcksEnsemble.GetMicrocksContainer().TestEndpointAsync(ctx, &testRequest, testResultChan)
		s.NoError(err)
	}()

	time.Sleep(500 * time.Millisecond)

	// Invoke the application to create an order.
	createdOrder, err := s.app.AppService.OrderService.PlaceOrder(&info)
	s.Require().NoError(err)

	// You may check additional stuff on createdOrder...
	s.NotNil(createdOrder)
	s.NotEmpty(createdOrder.ID)

	// Get the Microcks test result.
	testResult := <-testResultChan
	s.Require().NoError(err)

	s.T().Logf("Test Result success is %t", testResult.Success)

	// Log TestResult raw structure.
	j, err := json.Marshal(testResult)
	s.Require().NoError(err)
	s.T().Log(string(j))

	s.Require().True(testResult.Success)
	s.Equal(1, len(*testResult.TestCaseResults)) //nolint:testifylint

	// Check the content of the emitted event, read from Kafka topic.
	events, err := s.microcksEnsemble.GetMicrocksContainer().EventMessagesForTestCase(ctx, testResult, "SUBSCRIBE orders-created")
	s.Require().NoError(err)

	s.Equal(1, len(*events)) //nolint:testifylint

	message := (*events)[0].EventMessage
	var messageMap map[string]interface{}
	err = json.Unmarshal([]byte(*message.Content), &messageMap)
	s.Require().NoError(err)

	s.Equal("Creation", messageMap["changeReason"].(string))

	orderMap := messageMap["order"].(map[string]interface{})
	s.Equal("123-456-789", orderMap["customerId"].(string))
	s.InDelta(8.4, orderMap["totalPrice"].(float64), 0.01)

	productQuantities := orderMap["productQuantities"].([]interface{})
	s.Equal(2, len(productQuantities)) //nolint:testifylint
}

func (s *BaseSuite) TestEventIsConsumedAndProcessedByService() {
	err := waitFor(10*time.Second, func() error {
		order := s.app.AppService.OrderService.GetOrder("123-456-789")
		if order == nil {
			return errors.New("got no order '123-456-789' yet")
		}

		fmt.Printf("Order is %v\n", order)
		s.Equal("lbroudoux", order.CustomerID)
		s.Equal(model.VALIDATED, order.Status)
		s.Len(order.ProductQuantities, 2)
		return nil
	})
	s.Require().NoError(err)
}
