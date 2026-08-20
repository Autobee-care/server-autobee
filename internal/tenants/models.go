// Package tenants provides the tenants domain model.
package tenants

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// TenantStatus defines the lifecycle state of a tenant.
type TenantStatus string

const (
	TenantStatusActive   TenantStatus = "active"
	TenantStatusInactive TenantStatus = "inactive"
)

// Tenant is the stored tenant document in MongoDB.
type Tenant struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"-"`
	Name      string             `bson:"name"          json:"name"`
	Status    TenantStatus       `bson:"status"        json:"status"`
	CreatedAt time.Time          `bson:"createdAt"     json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"     json:"updatedAt"`
}

// TenantResponse is the public representation of a tenant.
type TenantResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Status    TenantStatus `json:"status"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// ToResponse converts a Tenant document to its public representation.
func ToResponse(t *Tenant) *TenantResponse {
	return &TenantResponse{
		ID:        t.ID.Hex(),
		Name:      t.Name,
		Status:    t.Status,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
