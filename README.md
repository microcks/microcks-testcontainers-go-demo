# microcks-testcontainers-go-demo

Go demonstration app on how to use Microcks Testcontainers in your dev/test workflow


## Basic commands

Start launching Microcks locally:

```sh
$ ./microcks.sh
==== OUTPUT ====
[...]
```

In another terminal, run the application:

```sh
$ go run cmd/main.go
==== OUTPUT ====
Starting Microcks TestContainers Go Demo application...
  Connecting to Kafka server: localhost:9092 
  Connecting to Microcks Pastries: http://localhost:9090/rest/API+Pastries/0.0.1 
%4|1725661943.865|CONFWARN|rdkafka#producer-2| [thrd:app]: Configuration property group.id is a consumer property and will be ignored by this producer instance
%4|1725661943.866|CONFWARN|rdkafka#producer-2| [thrd:app]: Configuration property auto.offset.reset is a consumer property and will be ignored by this producer instance
Microcks TestContainers Go Demo application is listening on localhost:9000

Consumed event from topic OrderEventsAPI-0.1.0-orders-reviewed: key = 1725661947312 value = {"timestamp":1706087114133,"order":{"id":"123-456-789","customerId":"lbroudoux","status":"VALIDATED","productQuantities":[{"productName":"Croissant","quantity":1},{"productName":"Pain Chocolat","quantity":1}],"totalPrice":4.2},"changeReason":"Validation"}
Order '123-456-789' has been updated after review
[...]
```

In a third terminal, call the API:

_Successful call_

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

_Error call_

```sh
$ curl -XPOST localhost:9000/api/orders -H 'Content-Type: application/json' \
    -d '{"customerId": "lbroudoux", "productQuantities": [{"productName": "Eclair Chocolat", "quantity": 1}], "totalPrice": 5.1}' -s | jq .
==== OUTPUT ====
{
  "productName": "Eclair Chocolat",
  "details": "Pastry Eclair Chocolat is not available"
}
```

## Running tests

```sh
$ go test -timeout 30s -run ^TestGetPastry$ github.com/microcks/microcks-testcontainers-go-demo/internal/client

=== RUN   TestGetPastry
2024/06/07 09:14:32 🐳 Creating container for image testcontainers/ryuk:0.7.0
2024/06/07 09:14:32 ✅ Container created: 409c6c59e370
2024/06/07 09:14:32 🐳 Starting container: 409c6c59e370
2024/06/07 09:14:33 ✅ Container started: 409c6c59e370
2024/06/07 09:14:33 🚧 Waiting for container id 409c6c59e370 image: testcontainers/ryuk:0.7.0. Waiting for: &{Port:8080/tcp timeout:<nil> PollInterval:100ms}
2024/06/07 09:14:33 🔔 Container is ready: 409c6c59e370
2024/06/07 09:14:33 🐳 Creating container for image quay.io/microcks/microcks-uber:1.9.0-native
2024/06/07 09:14:33 ✅ Container created: 5bc535cc9b57
2024/06/07 09:14:33 🐳 Starting container: 5bc535cc9b57
2024/06/07 09:14:33 ✅ Container started: 5bc535cc9b57
2024/06/07 09:14:33 🚧 Waiting for container id 5bc535cc9b57 image: quay.io/microcks/microcks-uber:1.9.0-native. Waiting for: &{timeout:<nil> Log:Started MicrocksApplication IsRegexp:false Occurrence:1 PollInterval:100ms}
2024/06/07 09:14:35 🔔 Container is ready: 5bc535cc9b57
2024/06/07 09:14:35 🐳 Terminating container: 5bc535cc9b57
2024/06/07 09:14:35 🚫 Container terminated: 5bc535cc9b57
--- PASS: TestGetPastry (3.46s)
PASS
ok      github.com/microcks/microcks-testcontainers-go-demo/internal/client     (cached)
```

```sh
$ go test -timeout 30s -run ^TestListPastries$ github.com/microcks/microcks-testcontainers-go-demo/internal/client

=== RUN   TestListPastries
2024/06/07 09:14:46 🐳 Creating container for image testcontainers/ryuk:0.7.0
2024/06/07 09:14:46 ✅ Container created: 7260f7bb8990
2024/06/07 09:14:46 🐳 Starting container: 7260f7bb8990
2024/06/07 09:14:46 ✅ Container started: 7260f7bb8990
2024/06/07 09:14:46 🚧 Waiting for container id 7260f7bb8990 image: testcontainers/ryuk:0.7.0. Waiting for: &{Port:8080/tcp timeout:<nil> PollInterval:100ms}
2024/06/07 09:14:47 🔔 Container is ready: 7260f7bb8990
2024/06/07 09:14:47 🐳 Creating container for image quay.io/microcks/microcks-uber:1.9.0-native
2024/06/07 09:14:47 ✅ Container created: 7ae147635552
2024/06/07 09:14:47 🐳 Starting container: 7ae147635552
2024/06/07 09:14:47 ✅ Container started: 7ae147635552
2024/06/07 09:14:47 🚧 Waiting for container id 7ae147635552 image: quay.io/microcks/microcks-uber:1.9.0-native. Waiting for: &{timeout:<nil> Log:Started MicrocksApplication IsRegexp:false Occurrence:1 PollInterval:100ms}
2024/06/07 09:14:47 🔔 Container is ready: 7ae147635552
2024/06/07 09:14:47 🐳 Terminating container: 7ae147635552
2024/06/07 09:14:47 🚫 Container terminated: 7ae147635552
--- PASS: TestListPastries (1.22s)
PASS
ok      github.com/microcks/microcks-testcontainers-go-demo/internal/client     (cached)
```

```sh
$ go test -timeout 30s -run ^TestOpenAPIContractBasic$ github.com/microcks/microcks-testcontainers-go-demo/internal/controller -v

$ go test -timeout 30s -run ^TestOpenAPIContractAdvanced$ github.com/microcks/microcks-testcontainers-go-demo/internal/controller -v
```