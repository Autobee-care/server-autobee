// Package service_centers provides business logic for the service centers module.
package service_centers

import "go.uber.org/zap"

// Service contains service centers business logic.
type Service struct {
	repo *Repository
	log  *zap.Logger
}

// NewService creates a new service centers Service.
func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}
