// Package vehicles provides request/response DTOs for the vehicles module.
package vehicles

// CreateVehicleRequest is the request body for POST /api/v1/vehicles.
// tenantId and userId are derived from the authenticated user's JWT context;
// clients must NOT supply them.
type CreateVehicleRequest struct {
	RegistrationNumber string   `json:"registrationNumber" validate:"required,min=2,max=20"`
	Make               string   `json:"make"               validate:"required,min=1,max=100"`
	Model              string   `json:"model"              validate:"required,min=1,max=100"`
	Year               int      `json:"year"               validate:"required,min=1886,max=2100"`
	FuelType           FuelType `json:"fuelType"           validate:"required,oneof=petrol diesel electric hybrid cng"`
}

// ListVehiclesQuery holds pagination query parameters.
type ListVehiclesQuery struct {
	Page  int `schema:"page"`
	Limit int `schema:"limit"`
}
