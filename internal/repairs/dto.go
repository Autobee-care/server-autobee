// Package repairs provides request/response DTOs for the repairs module.
package repairs

// CreateRepairRequest is the request body for creating a repair job.
type CreateRepairRequest struct {
	BookingID      string  `json:"bookingId"      validate:"required"`
	VehicleID      string  `json:"vehicleId"      validate:"required"`
	DiagnosisNotes string  `json:"diagnosisNotes" validate:"omitempty,max=1000"`
	EstimatedCost  float64 `json:"estimatedCost"  validate:"omitempty,min=0"`
}
