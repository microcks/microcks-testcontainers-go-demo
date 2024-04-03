package controller_test

import (
	"context"
	"testing"

	server "github.com/microcks/microcks-testcontainers-go-demo/cmd/run"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	client "microcks.io/go-client"
	microcks "microcks.io/testcontainers-go"
)

func setup(ctx context.Context, t *testing.T) *microcks.MicrocksContainer {
	microcksContainer, err := microcks.RunContainer(ctx,
		testcontainers.WithImage("quay.io/microcks/microcks-uber:1.9.0-native"),
		microcks.WithMainArtifact("../../testdata/order-service-openapi.yaml"),
		microcks.WithMainArtifact("../../testdata/apipastries-openapi.yaml"),
		microcks.WithSecondaryArtifact("../../testdata/apipastries-postman-collection.json"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := microcksContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
	return microcksContainer
}

func TestOpenAPIContract(t *testing.T) {
	ctx := context.Background()
	microcksContainer := setup(ctx, t)

	baseApiUrl, err := microcksContainer.RestMockEndpoint(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)

	go server.Run(baseApiUrl)
	defer server.Close()

	testRequest := client.TestRequest{
		ServiceId:    "Order Service API:0.1.0",
		RunnerType:   client.TestRunnerTypeOPENAPISCHEMA,
		TestEndpoint: "http://host.testcontainers.internal:9000/api",
		Timeout:      2000,
	}
	testResult, err := microcksContainer.TestEndpoint(ctx, &testRequest)
	require.NoError(t, err)

	t.Logf("Test Result success is %t", testResult.Success)
	// Test is failing at the moment because of
	// https://github.com/testcontainers/testcontainers-go/issues/2212
	require.False(t, testResult.Success)
}
