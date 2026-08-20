// Package bookings provides request/response DTOs for the bookings module.
package bookings

import "time"

// CreateBookingRequest is the request body for creating a booking.
type CreateBookingRequest struct {
	VehicleID       string    `json:"vehicleId"       validate:"required"`
	ServiceCenterID string    `json:"serviceCenterId" validate:"required"`
	ServiceID       string    `json:"serviceId"       validate:"required"`
	AppointmentDate time.Time `json:"appointmentDate" validate:"required"`
	Notes           string    `json:"notes"           validate:"omitempty,max=500"`
}
