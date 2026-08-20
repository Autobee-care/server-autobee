// Package repairs provides the repairs domain model.
package repairs

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// RepairStatus defines the lifecycle state of a repair job.
type RepairStatus string

const (
	RepairStatusOpen       RepairStatus = "open"
	RepairStatusInProgress RepairStatus = "in_progress"
	RepairStatusCompleted  RepairStatus = "completed"
)

// Repair is the stored repair document in MongoDB.
type Repair struct {
	ID              bson.ObjectID `bson:"_id,omitempty"`
	TenantID        bson.ObjectID `bson:"tenantId"`
	BookingID       bson.ObjectID `bson:"bookingId"`
	VehicleID       bson.ObjectID `bson:"vehicleId"`
	TechnicianID    bson.ObjectID `bson:"technicianId,omitempty"`
	Status          RepairStatus       `bson:"status"`
	DiagnosisNotes  string             `bson:"diagnosisNotes,omitempty"`
	EstimatedCost   float64            `bson:"estimatedCost,omitempty"`
	FinalCost       float64            `bson:"finalCost,omitempty"`
	CompletedAt     *time.Time         `bson:"completedAt,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt"`
}
