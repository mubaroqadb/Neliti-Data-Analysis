package main

import (
	"log"
	"net/http"
	"os"

	"github.com/research-data-analysis/config"
	"github.com/research-data-analysis/route"
)

func main() {
	// Auto-detect environment: production if running on Cloud Run, otherwise development
	if os.Getenv("ENVIRONMENT") == "" {
		if os.Getenv("K_SERVICE") != "" || os.Getenv("PORT") != "" {
			// Running on Cloud Run
			os.Setenv("ENVIRONMENT", "production")
		} else {
			// Running locally
			os.Setenv("ENVIRONMENT", "development")
		}
	}

	// Get port from environment variable (Cloud Run provides PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for local development
	}

	// Create HTTP server
	http.HandleFunc("/", route.URL)

	// Start server
	log.Printf("Server listening on port %s", port)
	log.Printf("Environment: %s", os.Getenv("ENVIRONMENT"))

	// Test MongoDB connection before starting server
	cfg := config.LoadConfig()
	if err := cfg.TestConnection(); err != nil {
		log.Printf("WARNING: MongoDB connection test failed: %v", err)
	} else {
		log.Printf("MongoDB connection test successful")
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
