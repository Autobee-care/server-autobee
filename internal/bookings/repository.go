// Package bookings provides MongoDB persistence for booking documents.
package bookings

import "go.mongodb.org/mongo-driver/v2/mongo"

// Repository handles MongoDB operations for the bookings collection.
// Implement following the vehicles.Repository pattern.
type Repository struct {
	col *mongo.Collection
}

// NewRepository creates a bookings Repository.
func NewRepository(col *mongo.Collection) *Repository {
	return &Repository{col: col}
}
