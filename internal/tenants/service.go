// Package tenants provides business logic for the tenants module.
package tenants

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

// Service contains tenants business logic.
type Service struct {
	repo *Repository
	log  *zap.Logger
}

// NewService creates a new tenants Service.
func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Create creates a new tenant. Only super_admin may call this — enforced at the route level.
func (s *Service) Create(ctx context.Context, req *CreateTenantRequest) (*TenantResponse, error) {
	tenant := &Tenant{Name: req.Name}
	created, err := s.repo.Create(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("creating tenant: %w", err)
	}
	return ToResponse(created), nil
}

// ErrInvalidTenantID is returned when an invalid hex string is provided as a tenant ID.
var ErrInvalidTenantID = errors.New("invalid tenant ID")

// GetByID retrieves a tenant by ID.
func (s *Service) GetByID(ctx context.Context, id string) (*TenantResponse, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidTenantID
	}

	tenant, err := s.repo.FindByID(ctx, oid)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("finding tenant: %w", err)
	}
	return ToResponse(tenant), nil
}

// ListAll lists all tenants. Reserved for super_admin only.
func (s *Service) ListAll(ctx context.Context) ([]*TenantResponse, error) {
	tenants, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing tenants: %w", err)
	}

	result := make([]*TenantResponse, 0, len(tenants))
	for _, t := range tenants {
		result = append(result, ToResponse(t))
	}
	return result, nil
}
