// Package service_centers provides the service centers domain model.
package service_centers

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ServiceCenter is the stored service center document in MongoDB.
type ServiceCenter struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	TenantID  bson.ObjectID `bson:"tenantId"`
	Name      string             `bson:"name"`
	Address   string             `bson:"address"`
	Phone     string             `bson:"phone"`
	IsActive  bool               `bson:"isActive"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}
