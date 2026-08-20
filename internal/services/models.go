// Package services provides the vehicle services domain model.
// Note: "services" here refers to auto repair services (e.g. oil change, tyre rotation).
package services

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// VehicleService is the stored service offering document in MongoDB.
type VehicleService struct {
	ID              bson.ObjectID `bson:"_id,omitempty"`
	TenantID        bson.ObjectID `bson:"tenantId"`
	ServiceCenterID bson.ObjectID `bson:"serviceCenterId"`
	Name            string             `bson:"name"`
	Description     string             `bson:"description,omitempty"`
	DurationMinutes int                `bson:"durationMinutes"`
	Price           float64            `bson:"price"`
	IsActive        bool               `bson:"isActive"`
	CreatedAt       time.Time          `bson:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt"`
}
