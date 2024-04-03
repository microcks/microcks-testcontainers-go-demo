package main

import (
	"os"

	run "github.com/microcks/microcks-testcontainers-go-demo/cmd/run"
)

func main() {
	pastryAPIBaseURL, present := os.LookupEnv("PASTRY_API_URL")
	if !present {
		pastryAPIBaseURL = "http://localhost:9090/rest/API+Pastries/0.0.1"
	}
	run.Run(pastryAPIBaseURL)
}
