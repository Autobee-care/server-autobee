// Package vehicles provides business logic for the vehicles module.
// This is the reference implementation of tenant isolation logic.
package vehicles

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/autobee/server/internal/auth"
	"github.com/autobee/server/internal/middleware"
	"github.com/autobee/server/pkg/response"
)

// Service contains the vehicles business logic.
type Service struct {
	repo *Repository
	log  *zap.Logger
}

// NewService creates a new vehicles Service.
func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// Create creates a new vehicle for the authenticated user.
// tenantId and userId are derived from the JWT context — never trusted from the client.
func (s *Service) Create(ctx context.Context, req *CreateVehicleRequest) (*VehicleResponse, error) {
	userID := middleware.GetUserID(ctx)
	tenantID := middleware.GetTenantID(ctx)

	tenantOID, err := bson.ObjectIDFromHex(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenantId in context: %w", err)
	}

	userOID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid userId in context: %w", err)
	}

	vehicle := &Vehicle{
		TenantID:           tenantOID,
		UserID:             userOID,
		RegistrationNumber: req.RegistrationNumber,
		Make:               req.Make,
		Model:              req.Model,
		Year:               req.Year,
		FuelType:           req.FuelType,
	}

	created, err := s.repo.Create(ctx, vehicle)
	if err != nil {
		if errors.Is(err, ErrDuplicateVehicle) {
			return nil, ErrDuplicateVehicle
		}
		return nil, fmt.Errorf("creating vehicle: %w", err)
	}

	return ToResponse(created), nil
}

// ListResult holds a page of vehicles and total count.
type ListResult struct {
	Vehicles   []*VehicleResponse
	Total      int64
	Page       int
	Limit      int
}

// List returns vehicles appropriate for the calling user's role.
// Tenant isolation is enforced here — users see only their own vehicles,
// tenant admins see their tenant's vehicles, super admins see all.
func (s *Service) List(ctx context.Context, page, limit int) (*ListResult, error) {
	role := auth.Role(middleware.GetRole(ctx))
	userID := middleware.GetUserID(ctx)
	tenantID := middleware.GetTenantID(ctx)

	tenantOID, err := bson.ObjectIDFromHex(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenantId: %w", err)
	}

	var (
		vehicles []*Vehicle
		total    int64
	)

	switch role {
	case auth.RoleSuperAdmin:
		vehicles, total, err = s.repo.FindAll(ctx, page, limit)
	case auth.RoleTenantAdmin:
		vehicles, total, err = s.repo.FindByTenantID(ctx, tenantOID, page, limit)
	default:
		// RoleUser — show only their own vehicles.
		userOID, uerr := bson.ObjectIDFromHex(userID)
		if uerr != nil {
			return nil, fmt.Errorf("invalid userId: %w", uerr)
		}
		vehicles, total, err = s.repo.FindByUserID(ctx, tenantOID, userOID, page, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("listing vehicles: %w", err)
	}

	responses := make([]*VehicleResponse, 0, len(vehicles))
	for _, v := range vehicles {
		responses = append(responses, ToResponse(v))
	}

	return &ListResult{
		Vehicles: responses,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}

// ErrInvalidVehicleID is returned when an invalid hex string is provided as a vehicle ID.
var ErrInvalidVehicleID = errors.New("invalid vehicle ID")

// GetByID retrieves a single vehicle by ID, enforcing tenant isolation.
func (s *Service) GetByID(ctx context.Context, vehicleID string) (*VehicleResponse, error) {
	role := auth.Role(middleware.GetRole(ctx))
	userID := middleware.GetUserID(ctx)
	tenantID := middleware.GetTenantID(ctx)

	vehicleOID, err := bson.ObjectIDFromHex(vehicleID)
	if err != nil {
		return nil, ErrInvalidVehicleID
	}

	vehicle, err := s.repo.FindByID(ctx, vehicleOID)
	if err != nil {
		if errors.Is(err, ErrVehicleNotFound) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("finding vehicle: %w", err)
	}

	// Tenant isolation check.
	if err := s.enforceAccess(vehicle, role, tenantID, userID); err != nil {
		return nil, err
	}

	return ToResponse(vehicle), nil
}

// enforceAccess checks that the caller is allowed to view the given vehicle.
func (s *Service) enforceAccess(v *Vehicle, role auth.Role, tenantID, userID string) error {
	switch role {
	case auth.RoleSuperAdmin:
		return nil
	case auth.RoleTenantAdmin:
		if v.TenantID.Hex() != tenantID {
			return response.ErrForbidden
		}
		return nil
	default:
		// RoleUser — must own the vehicle and be in the same tenant.
		if v.TenantID.Hex() != tenantID || v.UserID.Hex() != userID {
			return response.ErrForbidden
		}
		return nil
	}
}
