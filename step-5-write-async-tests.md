# Step 5: Let's write tests for the Async Events

Now that we address the REST/Synchronous part, let's have a look on the part related to Asynchronous Kafka events.
Testing of asynchronous or event-driven system is usually a pain for developers 🥲

## First Test - Verify our OrderService is publishing events

In this section, we'll focus on testing the `Order Service` + `Event Publisher` components of our application:

![Event Publisher Test](./assets/test-order-event-publisher.png)

Even if it may be easy to check that the creation of an event object has been triggered with frameworks like [Gomock](https://github.com/uber-go/mock)
or others, it's far more complicated to check that this event is correctly serialized, sent to a broker and valid
regarding an Event definition...

Fortunately, Microcks and TestContainers make this thing easy!

Let's review the test function `TestOrderEventIsPublishedWhenOrderIsCreated` under `internal/test/suite_test.go`.

```go
func (s *BaseSuite) TestOrderEventIsPublishedWhenOrderIsCreated() {
	ctx := context.Background()
	// Prepare a Microcks Tests.
	testRequest := client.TestRequest{
		ServiceId:          "Order Events API:0.1.0",
		RunnerType:         client.TestRunnerTypeASYNCAPISCHEMA,
		TestEndpoint:       "kafka://kafka:19092/orders-created",
		Timeout:            2000,
		FilteredOperations: &[]string{"SUBSCRIBE orders-created"},
	}

	// Prepare an application Order.
	info := model.OrderInfo{
		CustomerID: "123-456-789",
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
	go func() {
		err := s.microcksEnsemble.GetMicrocksContainer().TestEndpointAsync(ctx, &testRequest, testResultChan)
		s.NoError(err)
	}()

	time.Sleep(500 * time.Millisecond)

	// Invoke the application to create an order.
	createdOrder, err := s.app.AppService.OrderService.PlaceOrder(&info)
	s.Require().NoError(err)

	// You may check additional stuff on createdOrder...
	s.NotNil(createdOrder)
	s.NotEmpty(createdOrder.ID)

	// Get the Microcks test result.
	testResult := <-testResultChan
	s.Require().NoError(err)

	s.T().Logf("Test Result success is %t", testResult.Success)

	// Log TestResult raw structure.
	j, err := json.Marshal(testResult)
	s.Require().NoError(err)
	s.T().Log(string(j))

	s.Require().True(testResult.Success)
    s.Equal(1, len(*testResult.TestCaseResults)) //nolint:testifylint
}
```

> You can execute this test from your IDE or from the terminal using the `go test ./internal/test -test.timeout=20m -failfast -v -test.run TestBaseSuite -testify.m ^TestOrderEventIsPublishedWhenOrderIsCreated` command

Things are a bit more complex here, but we'll walk through step-by-step:
* Similarly to previous section, we prepared a Microcks-provided `TestRequest` object
  * We ask for a `AsyncAPI Schema` conformance test that will use the definition found into the `order-events-asyncapi.yml` contract,
  * We ask Microcks to listen to the `kafka://kafka:19092/orders-created` endpoint that represents the `orders-created` topic on our Kafka broker managed by Testcontainers,
  * We ask to focus on a specific operation definition to mimic consumers that subscribe to the  `orders-created` channel,
  * We specified a timeout value that means that Microcks will only listen during 5 seconds for incoming messages. 
* We also prepared an `OrderInfo` object that will be used as the input of the `PlaceOrder()` function invocation on application's `OrderService`.
* Then, we launched the test on the Microcks side. This time, the launch is asynchronous, so we're using a go routine that will give us a `TestResult` later on
on a `testResultChan` channel
  * We wait a bit here to ensure, Microcks got some time to start the test and connect to Kafka broker.
* We can invoke our business service by creating an order with `PlaceOrder()` method. We could assert whatever we want on created order as well.
* Finally, we wait for the routine completion to retrieve the `TestResult` on channel and assert on the success and check we received 1 message as a result.

The sequence diagram below details the test sequence. You'll see 2 parallel blocks being executed:
* One that corresponds to Microcks test - where it connects and listen for Kafka messages,
* One that corresponds to the `OrderService` invokation that is expected to trigger a message on Kafka.

```mermaid
sequenceDiagram
    par Launch Microcks test
      OrderServiceTests->>Microcks: testEndpointAsync()
      participant Microcks
      Note right of Microcks: Initialized at test startup
      Microcks->>Kafka: poll()
      Kafka-->>Microcks: messages
      Microcks-->Microcks: validate messages
    and Invoke OrderService
      OrderServiceTests->>+OrderService: placeOrder(OrderInfo)
      OrderService->>+OrderEventPublisher: publishEvent(OrderEvent)
      OrderEventPublisher->>Kafka: send("orders-created")
      OrderEventPublisher-->-OrderService: done
      OrderService-->-OrderServiceTests: Order
    end
    OrderServiceTests->>+Microcks: get()
    Note over OrderServiceTests,Microcks: After at most 2 seconds
    Microcks-->OrderServiceTests: TestResult
```

## Second Test - Verify our OrderEventListener is processing events

## Wrap-up

Thanks a lot for being through this quite long demonstration. We hope you learned new techniques for integration tests with both REST and Async/Event-driven APIs. Cheers! 🍻