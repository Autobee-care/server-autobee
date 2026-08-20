// Package vehicles provides the vehicles domain model.
// This is the reference module demonstrating the architectural pattern:
// Handler → Service → Repository → MongoDB.
package vehicles

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// FuelType defines the supported fuel types for a vehicle.
type FuelType string

const (
	FuelTypePetrol   FuelType = "petrol"
	FuelTypeDiesel   FuelType = "diesel"
	FuelTypeElectric FuelType = "electric"
	FuelTypeHybrid   FuelType = "hybrid"
	FuelTypeCNG      FuelType = "cng"
)

// Vehicle is the stored vehicle document in MongoDB.
type Vehicle struct {
	ID                 bson.ObjectID `bson:"_id,omitempty"`
	TenantID           bson.ObjectID `bson:"tenantId"`
	UserID             bson.ObjectID `bson:"userId"`
	RegistrationNumber string             `bson:"registrationNumber"`
	Make               string             `bson:"make"`
	Model              string             `bson:"model"`
	Year               int                `bson:"year"`
	FuelType           FuelType           `bson:"fuelType"`
	CreatedAt          time.Time          `bson:"createdAt"`
	UpdatedAt          time.Time          `bson:"updatedAt"`
}

// VehicleResponse is the public projection of a vehicle document.
type VehicleResponse struct {
	ID                 string    `json:"id"`
	TenantID           string    `json:"tenantId"`
	UserID             string    `json:"userId"`
	RegistrationNumber string    `json:"registrationNumber"`
	Make               string    `json:"make"`
	Model              string    `json:"model"`
	Year               int       `json:"year"`
	FuelType           FuelType  `json:"fuelType"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// ToResponse maps a Vehicle document to its public representation.
func ToResponse(v *Vehicle) *VehicleResponse {
	return &VehicleResponse{
		ID:                 v.ID.Hex(),
		TenantID:           v.TenantID.Hex(),
		UserID:             v.UserID.Hex(),
		RegistrationNumber: v.RegistrationNumber,
		Make:               v.Make,
		Model:              v.Model,
		Year:               v.Year,
		FuelType:           v.FuelType,
		CreatedAt:          v.CreatedAt,
		UpdatedAt:          v.UpdatedAt,
	}
}
