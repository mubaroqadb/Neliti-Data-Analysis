package main

import (
	"log"
	"net/http"
	"os"

	"github.com/research-data-analysis/route"
)

func main() {
	// Get port from environment variable (Cloud Run provides PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for local development
	}

	// Create HTTP server
	http.HandleFunc("/", route.URL)

	// Start server
	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}