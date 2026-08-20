// Package tenants provides MongoDB persistence for tenant documents.
package tenants

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ErrTenantNotFound is returned when a tenant lookup finds no documents.
var ErrTenantNotFound = errors.New("tenant not found")

// Repository handles MongoDB operations for the tenants collection.
type Repository struct {
	col *mongo.Collection
}

// NewRepository creates a tenants Repository.
func NewRepository(col *mongo.Collection) *Repository {
	return &Repository{col: col}
}

// Create inserts a new tenant document.
func (r *Repository) Create(ctx context.Context, t *Tenant) (*Tenant, error) {
	t.ID = bson.NewObjectID()
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now
	t.Status = TenantStatusActive

	_, err := r.col.InsertOne(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("insert tenant: %w", err)
	}
	return t, nil
}

// FindByID retrieves a tenant by ObjectID.
func (r *Repository) FindByID(ctx context.Context, id bson.ObjectID) (*Tenant, error) {
	var t Tenant
	err := r.col.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&t)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("find tenant: %w", err)
	}
	return &t, nil
}

// FindAll retrieves all tenants. Reserved for super_admin use.
func (r *Repository) FindAll(ctx context.Context) ([]*Tenant, error) {
	cursor, err := r.col.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find all tenants: %w", err)
	}
	defer cursor.Close(ctx)

	var tenants []*Tenant
	if err := cursor.All(ctx, &tenants); err != nil {
		return nil, fmt.Errorf("decode tenants: %w", err)
	}
	return tenants, nil
}
