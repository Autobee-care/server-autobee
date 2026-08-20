// Package users provides MongoDB persistence for user documents.
package users

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/autobee/server/internal/auth"
)

// ErrUserNotFound is returned when no user document matches the query.
var ErrUserNotFound = errors.New("user not found")

// Repository handles MongoDB operations for the users collection.
type Repository struct {
	col *mongo.Collection
}

// NewRepository creates a users Repository.
func NewRepository(col *mongo.Collection) *Repository {
	return &Repository{col: col}
}

// FindByID retrieves a user by their ObjectID.
func (r *Repository) FindByID(ctx context.Context, id bson.ObjectID) (*auth.User, error) {
	var u auth.User
	err := r.col.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}
