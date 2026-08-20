// Package service_centers provides MongoDB persistence for service center documents.
package service_centers

import "go.mongodb.org/mongo-driver/v2/mongo"

// Repository handles MongoDB operations for the service_centers collection.
type Repository struct {
	col *mongo.Collection
}

// NewRepository creates a service centers Repository.
func NewRepository(col *mongo.Collection) *Repository {
	return &Repository{col: col}
}
