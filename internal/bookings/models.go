// Package bookings provides the bookings domain model.
// Implement booking business logic here following the vehicles module pattern.
package bookings

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// BookingStatus defines the lifecycle state of a booking.
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCompleted BookingStatus = "completed"
	BookingStatusCancelled BookingStatus = "cancelled"
)

// Booking is the stored booking document in MongoDB.
type Booking struct {
	ID              bson.ObjectID `bson:"_id,omitempty"`
	TenantID        bson.ObjectID `bson:"tenantId"`
	UserID          bson.ObjectID `bson:"userId"`
	VehicleID       bson.ObjectID `bson:"vehicleId"`
	ServiceCenterID bson.ObjectID `bson:"serviceCenterId"`
	ServiceID       bson.ObjectID `bson:"serviceId"`
	Status          BookingStatus      `bson:"status"`
	AppointmentDate time.Time          `bson:"appointmentDate"`
	Notes           string             `bson:"notes,omitempty"`
	CreatedAt       time.Time          `bson:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt"`
}
