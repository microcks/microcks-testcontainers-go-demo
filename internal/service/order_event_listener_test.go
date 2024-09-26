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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/microcks/microcks-testcontainers-go-demo/internal/model"
	helper "github.com/microcks/microcks-testcontainers-go-demo/internal/test"
)

func TestEventIsConsumedAndProcessedByService(t *testing.T) {
	t.Run("Test", testEventIsConsumedAndProcessedByService(func(t *testing.T, tc *helper.TestContext) {
		// Test code goes here which can leverage the context
		fmt.Println("My test code got executed")

		// Try checking for received and processed order.
		done := make(chan struct{})
		go func() {
			defer close(done)
			orderFound := false
			for i := 0; i < 10; i++ {
				time.Sleep(500 * time.Millisecond)
				order := tc.AppServices.OrderService.GetOrder("123-456-789")
				if order == nil {
					fmt.Println("Got no order '123-456-789' yet...")
				}
				if order != nil {
					orderFound = true
					require.Equal(t, "lbroudoux", order.CustomerId)
					require.Equal(t, model.VALIDATED, order.Status)
					require.Equal(t, 2, len(order.ProductQuantities))
				}
			}
			require.True(t, orderFound)
		}()

		for {
			select {
			// Wait at most 5 seconds
			case <-time.After(5 * time.Second):
				return
			case <-done:
				return
			}
		}
	}))
}

func testEventIsConsumedAndProcessedByService(test func(t *testing.T, tc *helper.TestContext)) func(*testing.T) {
	return func(t *testing.T) {
		tc, err := helper.SetupTestContext(t)
		require.NoError(t, err)
		defer helper.TeardownTestContext(tc)
		test(t, tc)
	}
}
