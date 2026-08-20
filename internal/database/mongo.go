// Package database manages the MongoDB client lifecycle.
package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const (
	defaultConnectTimeout         = 10 * time.Second
	defaultServerSelectionTimeout = 10 * time.Second
)

// Client wraps the mongo.Client and exposes the target database.
type Client struct {
	client   *mongo.Client
	Database *mongo.Database
}

// Connect establishes a single MongoDB connection and pings the server.
// It must be called once at application startup.
func Connect(ctx context.Context, uri, dbName string) (*Client, error) {
	opts := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(defaultConnectTimeout).
		SetServerSelectionTimeout(defaultServerSelectionTimeout)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &Client{
		client:   client,
		Database: client.Database(dbName),
	}, nil
}

// Disconnect gracefully closes the MongoDB connection.
func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Collection returns a handle to the named collection.
func (c *Client) Collection(name string) *mongo.Collection {
	return c.Database.Collection(name)
}
