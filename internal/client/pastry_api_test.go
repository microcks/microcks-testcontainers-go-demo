package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/microcks/microcks-testcontainers-go-demo/internal/client"
	"github.com/testcontainers/testcontainers-go"
	microcks "microcks.io/testcontainers-go"
)

func setup(t *testing.T) *microcks.MicrocksContainer {
	ctx := context.Background()

	microcksContainer, err := microcks.RunContainer(ctx,
		testcontainers.WithImage("quay.io/microcks/microcks-uber:1.9.0-native"),
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

func TestGetPastry(t *testing.T) {
	ctx := context.Background()

	microcksContainer := setup(t)

	baseApiUrl, err := microcksContainer.RestMockEndpoint(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)
	pastryAPIClient := client.NewPastryAPIClient(baseApiUrl)

	pastry, err := pastryAPIClient.GetPastry("Millefeuille")
	require.NoError(t, err)
	require.Equal(t, "Millefeuille", pastry.Name)
	require.Equal(t, "available", pastry.Status)

	pastry, err = pastryAPIClient.GetPastry("Eclair Cafe")
	require.NoError(t, err)
	require.Equal(t, "Eclair Cafe", pastry.Name)
	require.Equal(t, "available", pastry.Status)

	pastry, err = pastryAPIClient.GetPastry("Eclair Chocolat")
	require.NoError(t, err)
	require.Equal(t, "Eclair Chocolat", pastry.Name)
	require.Equal(t, "unknown", pastry.Status)
}

func TestListPastries(t *testing.T) {
	ctx := context.Background()

	microcksContainer := setup(t)

	baseApiUrl, err := microcksContainer.RestMockEndpoint(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)
	pastryAPIClient := client.NewPastryAPIClient(baseApiUrl)

	pastries, err := pastryAPIClient.ListPastries("S")
	require.NoError(t, err)
	require.Equal(t, 1, len(*pastries))

	pastries, err = pastryAPIClient.ListPastries("M")
	require.NoError(t, err)
	require.Equal(t, 2, len(*pastries))

	pastries, err = pastryAPIClient.ListPastries("L")
	require.NoError(t, err)
	require.Equal(t, 2, len(*pastries))
}
