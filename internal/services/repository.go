// Package services provides MongoDB persistence for vehicle service documents.
package services

import "go.mongodb.org/mongo-driver/v2/mongo"

// Repository handles MongoDB operations for the services collection.
type Repository struct {
	col *mongo.Collection
}

// NewRepository creates a services Repository.
func NewRepository(col *mongo.Collection) *Repository {
	return &Repository{col: col}
}
