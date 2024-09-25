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

	"github.com/stretchr/testify/require"

	helper "github.com/microcks/microcks-testcontainers-go-demo/internal/test"
)

func TestEventIsConsumedAndProcessedByService(t *testing.T) {
	t.Run("Test", testEventIsConsumedAndProcessedByService(func(t *testing.T, ts *helper.TestContext) {
		// Test code goes here which can leverage the context
		fmt.Println("My test code got executed")
	}))
}

func testEventIsConsumedAndProcessedByService(test func(t *testing.T, ts *helper.TestContext)) func(*testing.T) {
	return func(t *testing.T) {
		tc, err := helper.SetupTestContext(t)
		require.NoError(t, err)
		defer helper.TeardownTestContext(tc)
		test(t, tc)
	}
}
