// Package bookings provides business logic for the bookings module.
package bookings

import "go.uber.org/zap"

// Service contains bookings business logic.
// Implement following the vehicles.Service pattern including tenant isolation.
type Service struct {
	repo *Repository
	log  *zap.Logger
}

// NewService creates a new bookings Service.
func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}
