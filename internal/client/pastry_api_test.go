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

package client_test

import (
	"context"
	"testing"

	"github.com/microcks/microcks-testcontainers-go-demo/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	microcks "microcks.io/testcontainers-go"
)

func setup(ctx context.Context, t *testing.T) *microcks.MicrocksContainer {
	t.Helper()

	microcksContainer, err := microcks.Run(ctx,
		"quay.io/microcks/microcks-uber:1.10.0-native",
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
	microcksContainer := setup(ctx, t)

	baseAPIURL, err := microcksContainer.RestMockEndpoint(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)
	pastryAPIClient := client.NewPastryAPIClient(baseAPIURL)

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

	// Check that the mock API has really been invoked.
	mockInvoked, err := microcksContainer.Verify(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)
	require.True(t, mockInvoked)
}

func TestListPastries(t *testing.T) {
	ctx := context.Background()
	microcksContainer := setup(ctx, t)

	baseAPIURL, err := microcksContainer.RestMockEndpoint(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)
	pastryAPIClient := client.NewPastryAPIClient(baseAPIURL)

	// Get the number of invocations before our test.
	beforeMockInvocations, err := microcksContainer.ServiceInvocationsCount(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)

	pastries, err := pastryAPIClient.ListPastries("S")
	require.NoError(t, err)
	assert.Len(t, pastries, 1)

	pastries, err = pastryAPIClient.ListPastries("M")
	require.NoError(t, err)
	assert.Len(t, pastries, 2)

	pastries, err = pastryAPIClient.ListPastries("L")
	require.NoError(t, err)
	assert.Len(t, pastries, 2)

	// Check our mock API has been invoked the correct number of times.
	afterMockInvocations, err := microcksContainer.ServiceInvocationsCount(ctx, "API Pastries", "0.0.1")
	require.NoError(t, err)
	require.Equal(t, 3, afterMockInvocations-beforeMockInvocations)
}
