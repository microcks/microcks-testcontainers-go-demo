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
package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	server "github.com/microcks/microcks-testcontainers-go-demo/cmd/run"
	app "github.com/microcks/microcks-testcontainers-go-demo/internal"
	"github.com/microcks/microcks-testcontainers-go-demo/internal/model"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	//kafkaTC "github.com/testcontainers/testcontainers-go/modules/kafka"
	kafkaTC "microcks.io/testcontainers-go/kafka"

	"github.com/testcontainers/testcontainers-go/network"
	"microcks.io/go-client"
	"microcks.io/testcontainers-go/ensemble"
	kafkaCon "microcks.io/testcontainers-go/ensemble/async/connection/kafka"
)

func setupEnsemble(ctx context.Context, t *testing.T, net *testcontainers.DockerNetwork, brokerUrl string) *ensemble.MicrocksContainersEnsemble {
	microcksEnsemble, err := ensemble.RunContainers(ctx,
		ensemble.WithMainArtifact("../../testdata/order-service-openapi.yaml"),
		ensemble.WithMainArtifact("../../testdata/order-events-asyncapi.yaml"),
		ensemble.WithMainArtifact("../../testdata/apipastries-openapi.yaml"),
		ensemble.WithSecondaryArtifact("../../testdata/apipastries-postman-collection.json"),
		ensemble.WithPostman(),
		ensemble.WithAsyncFeature(),
		ensemble.WithNetwork(net),
		ensemble.WithHostAccessPorts([]int{server.DefaultApplicationPort}),
		ensemble.WithKafkaConnection(kafkaCon.Connection{
			BootstrapServers: brokerUrl,
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := microcksEnsemble.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
	return microcksEnsemble
}

func TestOrderEventIsPublishedWhenOrderIsCreated(t *testing.T) {
	ctx := context.Background()

	// Common network and Kafka container.
	net, err := network.New(ctx, network.WithCheckDuplicate())
	if err != nil {
		require.NoError(t, err)
		return
	}

	kafkaContainer, err := kafkaTC.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		network.WithNetwork([]string{"kafka"}, net),

		testcontainers.WithEnv(map[string]string{
			"KAFKA_LISTENERS":                      "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092,CONTROLLER://0.0.0.0:9094,TC://0.0.0.0:19092",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "PLAINTEXT:PLAINTEXT,BROKER:PLAINTEXT,CONTROLLER:PLAINTEXT,TC:PLAINTEXT",
			"KAFKA_ADVERTISED_LISTENERS":           "PLAINTEXT://%s:%d,BROKER://%s:9092,TC://kafka:19092",
		}),
	)
	t.Cleanup(func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate Kafka container: %s", err)
		}
	})

	brokerURL, err := kafkaContainer.Brokers(ctx)
	if err != nil || len(brokerURL) == 0 {
		require.NoError(t, err)
		return
	}
	fmt.Println("BrokerURL is " + brokerURL[0])

	microcksEnsemble := setupEnsemble(ctx, t, net, "kafka:19092")

	baseApiUrl, err := microcksEnsemble.GetMicrocksContainer().RestMockEndpoint(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)

	applicationProperties := &app.ApplicationProperties{
		PastriesBaseUrl:          baseApiUrl,
		OrderEventsCreatedTopic:  "orders-created",
		OrderEventsReviewedTopic: "orders-reviewed",
		KafkaConfigMap: &kafka.ConfigMap{
			"bootstrap.servers": brokerURL[0],
			"group.id":          "order-service",
			"auto.offset.reset": "latest",
		},
	}

	appServicesChan := make(chan app.ApplicationServices)
	go server.Run(*applicationProperties, appServicesChan)
	defer server.Close()

	// Prepare a Microcks Tests.
	testRequest := client.TestRequest{
		ServiceId:          "Order Events API:0.1.0",
		RunnerType:         client.TestRunnerTypeASYNCAPISCHEMA,
		TestEndpoint:       "kafka://kafka:19092/orders-created",
		Timeout:            5000,
		FilteredOperations: &[]string{"SUBSCRIBE orders-created"},
	}

	// Prepare an application Order.
	info := model.OrderInfo{
		CustomerId: "123-456-789",
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

	appServices := <-appServicesChan

	testResultChan := make(chan *client.TestResult)
	go microcksEnsemble.GetMicrocksContainer().TestEndpointAsync(ctx, &testRequest, testResultChan)

	time.Sleep(750 * time.Millisecond)

	// Invoke the application to create an order.
	createdOrder, err := appServices.OrderService.PlaceOrder(&info)
	if err != nil {
		t.Error("OrderService raised an error while placing order: " + err.Error())
	}

	// You may check additional stuff on createdOrder...
	require.NotNil(t, createdOrder.Id)

	// Get the Microcks test result.
	testResult := <-testResultChan
	require.NoError(t, err)

	t.Logf("Test Result success is %t", testResult.Success)

	// Log TestResult raw structure.
	j, err := json.Marshal(testResult)
	t.Logf(string(j))

	require.True(t, testResult.Success)
}
