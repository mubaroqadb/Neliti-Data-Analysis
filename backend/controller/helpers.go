package controller

import (
	"fmt"
	"net/http"

	"github.com/research-data-analysis/config"
	"github.com/research-data-analysis/helper/watoken"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// getMongoDB returns MongoDB database instance
func getMongoDB() *mongo.Database {
	db, err := config.GetConfig().GetMongoDatabase()
	if err != nil {
		return nil
	}
	return db
}

// getUserIDFromToken extracts user ID from PASETO token
func getUserIDFromToken(r *http.Request) (primitive.ObjectID, error) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		fmt.Printf("No authorization header found\n")
		return primitive.NilObjectID, fmt.Errorf("no authorization header")
	}

	// Remove "Bearer " prefix if present
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// Debug log for development
	fmt.Printf("Token string length: %d\n", len(tokenString))

	// Decode token using public key
	publicKey := config.GetConfig().Auth.PublicKey
	fmt.Printf("Public key length: %d\n", len(publicKey))

	payload, err := watoken.Decode(publicKey, tokenString)
	if err != nil {
		fmt.Printf("Token decode error: %v\n", err)
		return primitive.NilObjectID, fmt.Errorf("invalid token: %v", err)
	}

	// Convert user ID from string to ObjectID
	userID, err := primitive.ObjectIDFromHex(payload.Id)
	if err != nil {
		fmt.Printf("Invalid user ID in token: %v\n", err)
		return primitive.NilObjectID, fmt.Errorf("invalid user ID in token: %v", err)
	}

	fmt.Printf("Successfully extracted user ID: %s\n", userID.Hex())
	return userID, nil
}
