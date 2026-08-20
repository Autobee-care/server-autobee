// Package services provides DTOs for the vehicle services module.
package services

// CreateServiceRequest is the request body for creating a service offering.
type CreateServiceRequest struct {
	ServiceCenterID string  `json:"serviceCenterId" validate:"required"`
	Name            string  `json:"name"            validate:"required,min=2,max=200"`
	Description     string  `json:"description"     validate:"omitempty,max=1000"`
	DurationMinutes int     `json:"durationMinutes" validate:"required,min=5"`
	Price           float64 `json:"price"           validate:"required,min=0"`
}
