// Package repairs provides business logic for the repairs module.
package repairs

import "go.uber.org/zap"

// Service contains repairs business logic.
type Service struct {
	repo *Repository
	log  *zap.Logger
}

// NewService creates a new repairs Service.
func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}
