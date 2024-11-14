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
	"testing"
	"time"

	"github.com/microcks/microcks-testcontainers-go-demo/internal/model"
	helper "github.com/microcks/microcks-testcontainers-go-demo/internal/test"
	"github.com/stretchr/testify/require"

	"microcks.io/go-client"
)

func TestOrderEventIsPublishedWhenOrderIsCreated(t *testing.T) {
	ctx := context.Background()
	t.Run("Test", testOrderEventIsPublishedWhenOrderIsCreated(func(t *testing.T, tc *helper.TestContext) {
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

		testResultChan := make(chan *client.TestResult)
		go tc.MicrocksEnsemble.GetMicrocksContainer().TestEndpointAsync(ctx, &testRequest, testResultChan)

		time.Sleep(750 * time.Millisecond)

		// Invoke the application to create an order.
		createdOrder, err := tc.AppServices.OrderService.PlaceOrder(&info)
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
	}))
}

func testOrderEventIsPublishedWhenOrderIsCreated(test func(t *testing.T, tc *helper.TestContext)) func(*testing.T) {
	return func(t *testing.T) {
		tc, err := helper.SetupTestContext(t)
		require.NoError(t, err)
		defer helper.TeardownTestContext(tc)
		test(t, tc)
	}
}
