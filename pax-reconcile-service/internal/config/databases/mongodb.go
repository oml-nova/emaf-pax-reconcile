package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	Client  *mongo.Client
	MongoDB *mongo.Database
)

func InitMongoDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("DB_URI")
	if uri == "" {
		return fmt.Errorf("DB_URI is not set")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return fmt.Errorf("DB_NAME is not set")
	}

	opts := options.Client().
		ApplyURI(uri).
		SetReadPreference(readpref.SecondaryPreferred())

	var err error
	Client, err = mongo.Connect(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	if err = Client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	MongoDB = Client.Database(dbName)
	return nil
}

func GetMongoDB() *mongo.Database { return MongoDB }

func GetMongoDBByName(dbName string) *mongo.Database {
	if Client == nil {
		return nil
	}
	return Client.Database(dbName)
}
