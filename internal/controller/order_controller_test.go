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
package controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	server "github.com/microcks/microcks-testcontainers-go-demo/cmd/run"
	app "github.com/microcks/microcks-testcontainers-go-demo/internal"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	kafkaTC "github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/network"
	client "microcks.io/go-client"
	ensemble "microcks.io/testcontainers-go/ensemble"
)

/*
func setup(ctx context.Context, t *testing.T) *microcks.MicrocksContainer {
	microcksContainer, err := microcks.Run(ctx,
		"quay.io/microcks/microcks-uber:1.10.0-native",
		microcks.WithMainArtifact("../../testdata/order-service-openapi.yaml"),
		microcks.WithMainArtifact("../../testdata/apipastries-openapi.yaml"),
		microcks.WithSecondaryArtifact("../../testdata/apipastries-postman-collection.json"),
		microcks.WithHostAccessPorts([]int{server.DefaultApplicationPort}),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := microcksContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
	return microcksContainer
}

func TestOpenAPIContractBasic(t *testing.T) {
	ctx := context.Background()
	microcksContainer := setup(ctx, t)

	baseApiUrl, err := microcksContainer.RestMockEndpoint(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)

	applicationProperties := &app.ApplicationProperties{
		PastriesBaseUrl: baseApiUrl,
	}

	go server.Run(*applicationProperties)
	defer server.Close()

	testRequest := client.TestRequest{
		ServiceId:    "Order Service API:0.1.0",
		RunnerType:   client.TestRunnerTypeOPENAPISCHEMA,
		TestEndpoint: fmt.Sprintf("http://host.testcontainers.internal:%d/api", server.DefaultApplicationPort),
		Timeout:      2000,
	}
	testResult, err := microcksContainer.TestEndpoint(ctx, &testRequest)
	require.NoError(t, err)

	t.Logf("Test Result success is %t", testResult.Success)

	require.True(t, testResult.Success)
	require.Equal(t, 1, len(*testResult.TestCaseResults))
}
*/

func setupEnsemble(ctx context.Context, t *testing.T, net *testcontainers.DockerNetwork) *ensemble.MicrocksContainersEnsemble {
	microcksEnsemble, err := ensemble.RunContainers(ctx,
		ensemble.WithMainArtifact("../../testdata/order-service-openapi.yaml"),
		ensemble.WithMainArtifact("../../testdata/apipastries-openapi.yaml"),
		ensemble.WithSecondaryArtifact("../../testdata/apipastries-postman-collection.json"),
		ensemble.WithPostman(true),
		//ensemble.WithNetwork(net),
		//ensemble.WithDefaultNetwork(),
		ensemble.WithHostAccessPorts([]int{server.DefaultApplicationPort}),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := microcksEnsemble.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
	return microcksEnsemble
}

func TestOpenAPIContractAdvanced(t *testing.T) {
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

	microcksEnsemble := setupEnsemble(ctx, t, net)

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

	go server.Run(*applicationProperties)
	defer server.Close()

	testRequest := client.TestRequest{
		ServiceId:    "Order Service API:0.1.0",
		RunnerType:   client.TestRunnerTypeOPENAPISCHEMA,
		TestEndpoint: fmt.Sprintf("http://host.testcontainers.internal:%d/api", server.DefaultApplicationPort),
		Timeout:      2000,
	}
	testResult, err := microcksEnsemble.GetMicrocksContainer().TestEndpoint(ctx, &testRequest)
	require.NoError(t, err)

	t.Logf("Test Result success is %t", testResult.Success)

	// Log TestResult raw structure.
	j, err := json.Marshal(testResult)
	t.Logf(string(j))

	require.True(t, testResult.Success)
	require.Equal(t, 1, len(*testResult.TestCaseResults))
}

func TestPostmanCollectionContract(t *testing.T) {
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
	)
	t.Cleanup(func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate Kafka container: %s", err)
		}
	})

	microcksEnsemble := setupEnsemble(ctx, t, net)

	baseApiUrl, err := microcksEnsemble.GetMicrocksContainer().RestMockEndpoint(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)

	brokerURL, err := kafkaContainer.Brokers(ctx)
	if err != nil || len(brokerURL) == 0 {
		require.NoError(t, err)
		return
	}

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

	go server.Run(*applicationProperties)
	defer server.Close()

	testRequest := client.TestRequest{
		ServiceId:    "Order Service API:0.1.0",
		RunnerType:   client.TestRunnerTypeOPENAPISCHEMA,
		TestEndpoint: fmt.Sprintf("http://host.testcontainers.internal:%d/api", server.DefaultApplicationPort),
		Timeout:      2000,
	}
	testResult, err := microcksEnsemble.GetMicrocksContainer().TestEndpoint(ctx, &testRequest)
	require.NoError(t, err)

	t.Logf("Test Result success is %t", testResult.Success)

	/*
		// Log TestResult raw structure.
		j, err := json.Marshal(testResult)
		t.Logf(string(j))
	*/

	require.True(t, testResult.Success)
	require.Equal(t, 1, len(*testResult.TestCaseResults))
}
