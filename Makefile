.PHONY: lint build test

lint:
	golangci-lint run

build:
	go build -v ./...

test:
	go test -v ./...
