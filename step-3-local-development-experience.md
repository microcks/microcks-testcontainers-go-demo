# Step 3: Local development experience with Microcks

Our application uses Kafka and external dependencies.

Currently, if you run the application from your terminal, you will see the following error:

```shell
go run cmd/main.go

Starting Microcks TestContainers Go Demo application...
  Connecting to Kafka server: localhost:9092 
  Connecting to Microcks Pastries: http://localhost:9090/rest/API+Pastries/0.0.1 
%4|1732031480.769|CONFWARN|rdkafka#producer-2| [thrd:app]: Configuration property group.id is a consumer property and will be ignored by this producer instance
%4|1732031480.769|CONFWARN|rdkafka#producer-2| [thrd:app]: Configuration property auto.offset.reset is a consumer property and will be ignored by this producer instance
Microcks TestContainers Go Demo application is listening on localhost:9000

2024/11/19 16:51:20.772192 main.go:98: Starting application with Pastry API URL: http://localhost:9090/rest/API+Pastries/0.0.1, Kafka bootstrap: localhost:9092
2024/11/19 16:51:20.847780 order_event_listener.go:115: Error reading message: Subscribed topic not available: OrderEventsAPI-0.1.0-orders-reviewed: Broker: Unknown topic or partition
```

To run the application locally, we need to have a Kafka broker up and running + the other dependencies corresponding to our Pastry API provider and reviewing system.

Instead of installing these services on our local machine, or using Docker to run these services manually,
we will use a utility tool with this simple command `microcks.sh`. Microcks docker-compose file (`microcks-docker-compose.yml`)
has been configured to automatically import the `Order API` contract but also the `Pastry API` contracts. Both APIs are discovered on startup
and Microcks UI should be available on `http://localhost:9090` in your browser:

```shell
$ ./microcks.sh
==== OUTPUT ====
[+] Running 4/4
 ✔ Container microcks-testcontainers-node-nest-demo-microcks-1  Started                                                                                                                         0.2s 
 ✔ Container microcks-kafka                                     Started                                                                                                                         0.2s 
 ✔ Container microcks-async-minion                              Started                                                                                                                         0.4s 
 ✔ Container microcks-testcontainers-node-nest-demo-importer-1  Started                                                                                                                         0.4s
```

Because our `Order Service` application has been configured to talk to Microcks mocks (see the default settings in `cmd/main.go`),
you should be able to directly call the Order API and invoke the whole chain made of the 3 components:

```sh
$ curl -XPOST localhost:9000/api/orders -H 'Content-Type: application/json' \
    -d '{"customerId": "lbroudoux", "productQuantities": [{"productName": "Millefeuille", "quantity": 1}], "totalPrice": 5.1}' -s | jq .
==== OUTPUT ====
{
  "customerId": "lbroudoux",
  "productQuantities": [
    {
      "productName": "Millefeuille",
      "quantity": 1
    }
  ],
  "totalPrice": 5.1,
  "id": "dded1111-8e99-4ba7-8755-e718972480e3",
  "status": "CREATED"
}
```

## Review application configuration under cmd/main.go

In order to specify the dependant services we need, we use a `Config` structure that can be either populated from environment variables or defaults into `cmd/main.go`.

The defaults bind our application to the services provided by Microcks or loaded via the `microcks-docker-compose.yml` file.

```go
const (
    defaultPastryAPIURL   = "http://localhost:9090/rest/API+Pastries/0.0.1"
    defaultKafkaBootstrap = "localhost:9092"
    shutdownTimeout       = 15 * time.Second
    defaultOrdersTopic    = "orders-created"
    defaultReviewedTopic  = "OrderEventsAPI-0.1.0-orders-reviewed"
)

// loadConfig loads configuration from environment variables with defaults.
func loadConfig() *Config {
    return &Config{
	  PastryAPIURL:   getEnv("PASTRY_API_URL", defaultPastryAPIURL),
	  KafkaBootstrap: getEnv("KAFKA_BOOTSTRAP_URL", defaultKafkaBootstrap),
	  OrdersTopic:    getEnv("ORDERS_TOPIC", defaultOrdersTopic),
        ReviewedTopic:  getEnv("REVIEWED_TOPIC", defaultReviewedTopic),
    }
}
```

## Play with the API

### Create an order

```shell
curl -XPOST localhost:9000/api/orders -H 'Content-type: application/json' \
    -d '{"customerId": "lbroudoux", "productQuantities": [{"productName": "Millefeuille", "quantity": 1}], "totalPrice": 5.1}' -v
```

You should get a response similar to the following:

```shell
< HTTP/1.1 201 
< Content-Type: application/json
< Transfer-Encoding: chunked
< Date: Mon, 19 Nov 2024 17:15:42 GMT
< 
* Connection #0 to host localhost left intact
{"id":"2da3a517-9b3b-4788-81b5-b1a1aac71746","status":"CREATED","customerId":"lbroudoux","productQuantities":[{"productName":"Millefeuille","quantity":1}],"totalPrice":5.1}%
```

Now test with something else, requesting for another Pastry:

```shell
curl -XPOST localhost:9000/api/orders -H 'Content-type: application/json' \
    -d '{"customerId": "lbroudoux", "productQuantities": [{"productName": "Eclair Chocolat", "quantity": 1}], "totalPrice": 4.1}' -v
```

This time you get another "exception" response:

```shell
< HTTP/1.1 422 
< Content-Type: application/json
< Transfer-Encoding: chunked
< Date: Mon, 19 Nov 2024 17:19:08 GMT
< 
* Connection #0 to host localhost left intact
{"productName":"Eclair Chocolat","details":"Pastry Eclair Chocolat is not available"}%
```

and this is because Microcks has created different simulations for the Pastry API 3rd party API based on API artifacts we loaded.
Check the `testdata/apipastries-openapi.yaml` and `testdata/apipastries-postman-collection.json` files to get details.

### 
[Next](step-4-write-rest-tests.md)