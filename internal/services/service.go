// Package services provides business logic for the vehicle services module.
package services

import "go.uber.org/zap"

// Service contains vehicle services business logic.
type Service struct {
	repo *Repository
	log  *zap.Logger
}

// NewService creates a new vehicle services Service.
func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}
