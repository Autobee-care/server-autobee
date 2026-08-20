// Package service_centers provides DTOs for the service centers module.
package service_centers

// CreateServiceCenterRequest is the request body for creating a service center.
type CreateServiceCenterRequest struct {
	Name    string `json:"name"    validate:"required,min=2,max=200"`
	Address string `json:"address" validate:"required,max=500"`
	Phone   string `json:"phone"   validate:"required"`
}
