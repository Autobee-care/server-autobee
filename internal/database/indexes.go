// Package database provides index management for MongoDB collections.
package database

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EnsureIndexes creates all application indexes idempotently.
// It is safe to call on every startup; MongoDB will skip existing indexes.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	if err := ensureUserIndexes(ctx, db); err != nil {
		return fmt.Errorf("user indexes: %w", err)
	}
	if err := ensureVehicleIndexes(ctx, db); err != nil {
		return fmt.Errorf("vehicle indexes: %w", err)
	}
	if err := ensureBookingIndexes(ctx, db); err != nil {
		return fmt.Errorf("booking indexes: %w", err)
	}
	if err := ensureTenantIndexes(ctx, db); err != nil {
		return fmt.Errorf("tenant indexes: %w", err)
	}
	return nil
}

func ensureUserIndexes(ctx context.Context, db *mongo.Database) error {
	col := db.Collection("users")
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "phone", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "email", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.D{
				{Key: "email", Value: bson.D{{Key: "$type", Value: "string"}}},
			}),
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "status", Value: 1},
			},
		},
	}
	return createIndexes(ctx, col, indexes)
}

func ensureVehicleIndexes(ctx context.Context, db *mongo.Database) error {
	col := db.Collection("vehicles")
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "registrationNumber", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
			},
		},
	}
	return createIndexes(ctx, col, indexes)
}

func ensureBookingIndexes(ctx context.Context, db *mongo.Database) error {
	col := db.Collection("bookings")
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "status", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "serviceCenterId", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenantId", Value: 1},
				{Key: "appointmentDate", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
			},
		},
	}
	return createIndexes(ctx, col, indexes)
}

func ensureTenantIndexes(ctx context.Context, db *mongo.Database) error {
	col := db.Collection("tenants")
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
			},
		},
	}
	return createIndexes(ctx, col, indexes)
}

// createIndexes creates indexes on a collection, gracefully handling option updates and existing indexes.
func createIndexes(ctx context.Context, col *mongo.Collection, indexes []mongo.IndexModel) error {
	for _, idx := range indexes {
		_, err := col.Indexes().CreateOne(ctx, idx)
		if err != nil {
			// If index specs/options were changed during development, drop the conflicting index and recreate.
			if strings.Contains(err.Error(), "IndexKeySpecsConflict") || strings.Contains(err.Error(), "IndexOptionsConflict") {
				// Extract the conflicting index name if present in the error message
				if match := extractIndexName(err.Error()); match != "" {
					_ = col.Indexes().DropOne(ctx, match)
					_, err = col.Indexes().CreateOne(ctx, idx)
				}
			}
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("create index on %s: %w", col.Name(), err)
			}
		}
	}
	return nil
}

func extractIndexName(errMsg string) string {
	// Look for name: "index_name" pattern in error string
	const prefix = `name: "`
	idx := strings.Index(errMsg, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.Index(errMsg[start:], `"`)
	if end == -1 {
		return ""
	}
	return errMsg[start : start+end]
}
