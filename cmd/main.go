package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Initialize your application
	fmt.Println("Starting Microcks TestContainers Go Demo application...")

	// Define your HTTP routes
	http.HandleFunc("/", handler)

	// Start your HTTP server
	http.ListenAndServe(":9000", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Your HTTP request handler logic goes here
	fmt.Fprintf(w, "Hello, World!")
}
