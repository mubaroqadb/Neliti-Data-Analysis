package config

import (
	"context"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoString     string
	Mongoconn       *mongo.Database
	PrivateKey      string
	PublicKey       string
	GCSBucket       string
	GCPProjectID    string
	VertexAIRegion  string
)

// SetEnv mengambil environment variables
func SetEnv() {
	MongoString = os.Getenv("MONGOSTRING")
	PrivateKey = os.Getenv("PRIVATEKEY")
	PublicKey = os.Getenv("PUBLICKEY")
	GCSBucket = os.Getenv("GCS_BUCKET")
	GCPProjectID = os.Getenv("GCP_PROJECT_ID")
	VertexAIRegion = os.Getenv("VERTEXAI_REGION")
	
	if VertexAIRegion == "" {
		VertexAIRegion = "asia-southeast1"
	}

	// Initialize MongoDB connection
	if MongoString != "" && Mongoconn == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoString))
		if err == nil {
			Mongoconn = client.Database("research_data_analysis")
		}
	}
}

// SetAccessControlHeaders mengatur CORS headers
func SetAccessControlHeaders(w http.ResponseWriter, r *http.Request) bool {
	// Allowed origins
	allowedOrigins := map[string]bool{
		"https://research-data-analysis.github.io": true,
		"http://localhost:8080":                    true,
		"http://127.0.0.1:8080":                    true,
	}

	origin := r.Header.Get("Origin")
	if allowedOrigins[origin] {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Login, Secret")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "3600")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}