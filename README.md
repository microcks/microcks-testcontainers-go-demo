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
$ go test -timeout 30s -run "^TestGetPastry$" ./internal/client -v

=== RUN   TestGetPastry
2024/09/25 22:05:32 github.com/testcontainers/testcontainers-go - Connected to docker: 
  Server Version: 24.0.2
  API Version: 1.43
  Operating System: Docker Desktop
  Total Memory: 11962 MB
  Testcontainers for Go Version: v0.34.0
  Resolved Docker Host: unix:///var/run/docker.sock
  Resolved Docker Socket Path: /var/run/docker.sock
  Test SessionID: 1b0461d7b7d13ee30ffb86fc72d836fcc9d8ae715fb0407d2af3f4e088d26657
  Test ProcessID: 7629b56f-618d-4c6d-a37d-37ea9b7d6e56
2024/09/25 22:05:32 🐳 Creating container for image testcontainers/ryuk:0.9.0
2024/09/25 22:05:32 ✅ Container created: c23103c472eb
2024/09/25 22:05:32 🐳 Starting container: c23103c472eb
2024/09/25 22:05:32 ✅ Container started: c23103c472eb
2024/09/25 22:05:32 ⏳ Waiting for container id c23103c472eb image: testcontainers/ryuk:0.9.0. Waiting for: &{Port:8080/tcp timeout:<nil> PollInterval:100ms skipInternalCheck:false}
2024/09/25 22:05:32 🔔 Container is ready: c23103c472eb
2024/09/25 22:05:32 🐳 Creating container for image quay.io/microcks/microcks-uber:1.9.0-native
2024/09/25 22:05:32 ✅ Container created: 77ddd018ca5f
2024/09/25 22:05:32 🐳 Starting container: 77ddd018ca5f
2024/09/25 22:05:32 ✅ Container started: 77ddd018ca5f
2024/09/25 22:05:32 ⏳ Waiting for container id 77ddd018ca5f image: quay.io/microcks/microcks-uber:1.9.0-native. Waiting for: &{timeout:<nil> Log:Started MicrocksApplication IsRegexp:false Occurrence:1 PollInterval:100ms}
2024/09/25 22:05:33 🔔 Container is ready: 77ddd018ca5f
2024/09/25 22:05:33 🐳 Terminating container: 77ddd018ca5f
2024/09/25 22:05:33 🚫 Container terminated: 77ddd018ca5f
--- PASS: TestGetPastry (1.61s)
PASS
ok      github.com/microcks/microcks-testcontainers-go-demo/internal/client     1.940s
```

```sh
$ go test -timeout 30s -run "^TestListPastries$" ./internal/client -v

=== RUN   TestListPastries
2024/09/25 22:05:00 github.com/testcontainers/testcontainers-go - Connected to docker: 
  Server Version: 24.0.2
  API Version: 1.43
  Operating System: Docker Desktop
  Total Memory: 11962 MB
  Testcontainers for Go Version: v0.34.0
  Resolved Docker Host: unix:///var/run/docker.sock
  Resolved Docker Socket Path: /var/run/docker.sock
  Test SessionID: 96332d7af971d08d478592eca13f0c15f30f89ee17251b870f732595e9f5f341
  Test ProcessID: bfbf6820-0924-4e6a-b6f9-9eeb5f677ac5
2024/09/25 22:05:00 🐳 Creating container for image testcontainers/ryuk:0.9.0
2024/09/25 22:05:00 ✅ Container created: eb0f86cd914d
2024/09/25 22:05:00 🐳 Starting container: eb0f86cd914d
2024/09/25 22:05:00 ✅ Container started: eb0f86cd914d
2024/09/25 22:05:00 ⏳ Waiting for container id eb0f86cd914d image: testcontainers/ryuk:0.9.0. Waiting for: &{Port:8080/tcp timeout:<nil> PollInterval:100ms skipInternalCheck:false}
2024/09/25 22:05:00 🔔 Container is ready: eb0f86cd914d
2024/09/25 22:05:00 🐳 Creating container for image quay.io/microcks/microcks-uber:1.9.0-native
2024/09/25 22:05:00 ✅ Container created: a7f410f91fb2
2024/09/25 22:05:00 🐳 Starting container: a7f410f91fb2
2024/09/25 22:05:01 ✅ Container started: a7f410f91fb2
2024/09/25 22:05:01 ⏳ Waiting for container id a7f410f91fb2 image: quay.io/microcks/microcks-uber:1.9.0-native. Waiting for: &{timeout:<nil> Log:Started MicrocksApplication IsRegexp:false Occurrence:1 PollInterval:100ms}
2024/09/25 22:05:01 🔔 Container is ready: a7f410f91fb2
2024/09/25 22:05:01 🐳 Terminating container: a7f410f91fb2
2024/09/25 22:05:01 🚫 Container terminated: a7f410f91fb2
--- PASS: TestListPastries (1.27s)
PASS
ok      github.com/microcks/microcks-testcontainers-go-demo/internal/client     1.586s
```

```sh
$ go test -timeout 30s -run "^TestOpenAPIContractAdvanced$" ./internal/controller -v

$ go test -timeout 30s -run "^TestPostmanCollectionContract$" ./internal/controller -v

$ go test -timeout 30s -run "^TestOrderEventIsPublishedWhenOrderIsCreated$" ./internal/service -v

$ go test -timeout 30s -run "^TestEventIsConsumedAndProcessedByService$" ./internal/service -v
```
