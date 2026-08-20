// Package auth provides the MongoDB persistence layer for authentication.
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ErrUserNotFound is returned when a user lookup produces no result.
var ErrUserNotFound = errors.New("user not found")

// ErrDuplicateUser is returned when a unique constraint is violated.
var ErrDuplicateUser = errors.New("user already exists")

// Repository handles all MongoDB operations for user documents.
type Repository struct {
	col *mongo.Collection
}

// NewRepository creates a new auth Repository backed by the given collection.
func NewRepository(col *mongo.Collection) *Repository {
	return &Repository{col: col}
}

// Create inserts a new user document and returns the created user.
func (r *Repository) Create(ctx context.Context, u *User) (*User, error) {
	u.ID = bson.NewObjectID()
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	_, err := r.col.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrDuplicateUser
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

// FindByPhone returns a user matching the given phone number within a tenant.
func (r *Repository) FindByPhone(ctx context.Context, tenantID bson.ObjectID, phone string) (*User, error) {
	filter := bson.D{
		{Key: "tenantId", Value: tenantID},
		{Key: "phone", Value: phone},
	}
	return r.findOne(ctx, filter)
}

// FindByEmail returns a user matching the given email within a tenant.
func (r *Repository) FindByEmail(ctx context.Context, tenantID bson.ObjectID, email string) (*User, error) {
	filter := bson.D{
		{Key: "tenantId", Value: tenantID},
		{Key: "email", Value: email},
	}
	return r.findOne(ctx, filter)
}

// FindByID returns a user by their ObjectID.
func (r *Repository) FindByID(ctx context.Context, id bson.ObjectID) (*User, error) {
	return r.findOne(ctx, bson.D{{Key: "_id", Value: id}})
}

func (r *Repository) findOne(ctx context.Context, filter bson.D) (*User, error) {
	var u User
	err := r.col.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &u, nil
}
