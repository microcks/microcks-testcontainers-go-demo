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

	server "github.com/microcks/microcks-testcontainers-go-demo/cmd/run"
	helper "github.com/microcks/microcks-testcontainers-go-demo/internal/test"
	"github.com/stretchr/testify/require"
	client "microcks.io/go-client"
)

func TestOpenAPIContractAdvanced(t *testing.T) {
	ctx := context.Background()
	t.Run("Test", testOpenAPIContractAdvanced(func(t *testing.T, tc *helper.TestContext) {
		// Test code goes here which can leverage the context
		// Prepare a Microcks Test.
		testRequest := client.TestRequest{
			ServiceId:    "Order Service API:0.1.0",
			RunnerType:   client.TestRunnerTypeOPENAPISCHEMA,
			TestEndpoint: fmt.Sprintf("http://host.testcontainers.internal:%d/api", server.DefaultApplicationPort),
			Timeout:      2000,
		}
		testResult, err := tc.MicrocksEnsemble.GetMicrocksContainer().TestEndpoint(ctx, &testRequest)
		require.NoError(t, err)

		t.Logf("Test Result success is %t", testResult.Success)

		// Log TestResult raw structure.
		j, err := json.Marshal(testResult)
		t.Logf(string(j))

		require.True(t, testResult.Success)
		require.Equal(t, 1, len(*testResult.TestCaseResults))
	}))
}

func testOpenAPIContractAdvanced(test func(t *testing.T, tc *helper.TestContext)) func(*testing.T) {
	return func(t *testing.T) {
		tc, err := helper.SetupTestContext(t)
		require.NoError(t, err)
		defer helper.TeardownTestContext(tc)
		test(t, tc)
	}
}

func TestPostmanCollectionContract(t *testing.T) {
	ctx := context.Background()
	t.Run("Test", testPostmanCollectionContract(func(t *testing.T, tc *helper.TestContext) {
		// Test code goes here which can leverage the context
		// Prepare a Microcks Test.
		testRequest := client.TestRequest{
			ServiceId:    "Order Service API:0.1.0",
			RunnerType:   client.TestRunnerTypePOSTMAN,
			TestEndpoint: fmt.Sprintf("http://host.testcontainers.internal:%d/api", server.DefaultApplicationPort),
			Timeout:      2000,
		}
		testResult, err := tc.MicrocksEnsemble.GetMicrocksContainer().TestEndpoint(ctx, &testRequest)
		require.NoError(t, err)

		t.Logf("Test Result success is %t", testResult.Success)

		// Log TestResult raw structure.
		j, err := json.Marshal(testResult)
		t.Logf(string(j))

		require.True(t, testResult.Success)
		require.Equal(t, 1, len(*testResult.TestCaseResults))
	}))
}

func testPostmanCollectionContract(test func(t *testing.T, tc *helper.TestContext)) func(*testing.T) {
	return func(t *testing.T) {
		tc, err := helper.SetupTestContext(t)
		require.NoError(t, err)
		defer helper.TeardownTestContext(tc)
		test(t, tc)
	}
}
