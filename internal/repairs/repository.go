// Package repairs provides MongoDB persistence for repair documents.
package repairs

import "go.mongodb.org/mongo-driver/v2/mongo"

// Repository handles MongoDB operations for the repairs collection.
type Repository struct {
	col *mongo.Collection
}

// NewRepository creates a repairs Repository.
func NewRepository(col *mongo.Collection) *Repository {
	return &Repository{col: col}
}
