// Package vehicles provides MongoDB persistence for vehicle documents.
package vehicles

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ErrVehicleNotFound is returned when a vehicle lookup produces no result.
var ErrVehicleNotFound = errors.New("vehicle not found")

// ErrDuplicateVehicle is returned when a registration number already exists in the tenant.
var ErrDuplicateVehicle = errors.New("vehicle with this registration number already exists")

// Repository handles all MongoDB operations for vehicle documents.
type Repository struct {
	col *mongo.Collection
}

// NewRepository creates a new vehicles Repository.
func NewRepository(col *mongo.Collection) *Repository {
	return &Repository{col: col}
}

// Create inserts a new vehicle document.
func (r *Repository) Create(ctx context.Context, v *Vehicle) (*Vehicle, error) {
	v.ID = bson.NewObjectID()
	now := time.Now().UTC()
	v.CreatedAt = now
	v.UpdatedAt = now

	_, err := r.col.InsertOne(ctx, v)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrDuplicateVehicle
		}
		return nil, fmt.Errorf("insert vehicle: %w", err)
	}
	return v, nil
}

// FindByID retrieves a vehicle by its ObjectID.
func (r *Repository) FindByID(ctx context.Context, id bson.ObjectID) (*Vehicle, error) {
	var v Vehicle
	err := r.col.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&v)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("find vehicle by id: %w", err)
	}
	return &v, nil
}

// FindByUserID returns all vehicles belonging to a specific user within a tenant.
func (r *Repository) FindByUserID(ctx context.Context, tenantID, userID bson.ObjectID, page, limit int) ([]*Vehicle, int64, error) {
	filter := bson.D{
		{Key: "tenantId", Value: tenantID},
		{Key: "userId", Value: userID},
	}
	return r.findMany(ctx, filter, page, limit)
}

// FindByTenantID returns all vehicles within a tenant (for tenant admin).
func (r *Repository) FindByTenantID(ctx context.Context, tenantID bson.ObjectID, page, limit int) ([]*Vehicle, int64, error) {
	filter := bson.D{
		{Key: "tenantId", Value: tenantID},
	}
	return r.findMany(ctx, filter, page, limit)
}

// FindAll returns all vehicles across all tenants (for super admin).
func (r *Repository) FindAll(ctx context.Context, page, limit int) ([]*Vehicle, int64, error) {
	return r.findMany(ctx, bson.D{}, page, limit)
}

func (r *Repository) findMany(ctx context.Context, filter bson.D, page, limit int) ([]*Vehicle, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count vehicles: %w", err)
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find vehicles: %w", err)
	}
	defer cursor.Close(ctx)

	var vehicles []*Vehicle
	if err := cursor.All(ctx, &vehicles); err != nil {
		return nil, 0, fmt.Errorf("decode vehicles: %w", err)
	}

	return vehicles, total, nil
}
