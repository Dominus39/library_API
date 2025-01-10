package config

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectionDatabaseUsers(ctx context.Context) (*mongo.Collection, error) {
	// Define the MongoDB connection string
	var mongoURI string

	if os.Getenv("RUNNING_IN_DOCKER") == "true" {
		mongoURI = "mongodb://host.docker.internal:27017"
	} else {
		mongoURI = "mongodb://localhost:27017" // For local development
	}

	// Create client options with the URI
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Create a MongoDB client
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Test the connection with a timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	// Return the collection
	collection := client.Database("GC2").Collection("users")
	return collection, nil
}
