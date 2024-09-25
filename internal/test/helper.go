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
package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	server "github.com/microcks/microcks-testcontainers-go-demo/cmd/run"
	app "github.com/microcks/microcks-testcontainers-go-demo/internal"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"microcks.io/testcontainers-go/ensemble"

	kafkaTC "microcks.io/testcontainers-go/kafka"
	//kafkaTC "github.com/testcontainers/testcontainers-go/modules/kafka"
	kafkaCon "microcks.io/testcontainers-go/ensemble/async/connection/kafka"
)

type TestContext struct {
	microcksEnsemble *ensemble.MicrocksContainersEnsemble
	kafkaContainer   *kafkaTC.KafkaContainer
	appServices      *app.ApplicationServices
}

func SetupTestContext(t *testing.T) (*TestContext, error) {
	fmt.Println("SetupTestContext called")
	ctx := context.Background()

	// Common network.
	net, err := network.New(ctx, network.WithCheckDuplicate())
	if err != nil {
		require.NoError(t, err)
		return nil, err
	}

	// Configure and staratup a new KafkaContainer.
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

	// Configure and staratup a new MicrocksContainersEnsemble.
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
			BootstrapServers: "kafka:19092",
		}),
	)
	t.Cleanup(func() {
		if err := microcksEnsemble.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	brokerURL, err := kafkaContainer.Brokers(ctx)
	if err != nil || len(brokerURL) == 0 {
		return nil, err
	}
	fmt.Println("BrokerURL is " + brokerURL[0])

	// Configure and start the application.
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

	appServices := <-appServicesChan

	return &TestContext{
		microcksEnsemble: microcksEnsemble,
		kafkaContainer:   kafkaContainer,
		appServices:      &appServices,
	}, nil
}

func TeardownTestContext(tc *TestContext) {
	fmt.Println("TeardownTestContext called")
}
