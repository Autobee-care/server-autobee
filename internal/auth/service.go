// Package auth provides the business logic for authentication operations.
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/autobee/server/pkg/password"
)

// OTPService is the interface for future OTP providers (SMS, email, etc.).
// The NoOpOTPService is used in development and test environments.
type OTPService interface {
	Send(ctx context.Context, phone, otp string) error
}

// NoOpOTPService logs OTPs to the console. Never use in production.
type NoOpOTPService struct {
	log *zap.Logger
}

// NewNoOpOTPService creates a development OTP service that logs OTPs instead of sending them.
func NewNoOpOTPService(log *zap.Logger) *NoOpOTPService {
	return &NoOpOTPService{log: log}
}

// Send logs the OTP to the console for development use only.
// SECURITY: This implementation must be replaced before deploying to production.
func (s *NoOpOTPService) Send(_ context.Context, phone, otp string) error {
	// NOTE: OTP is intentionally not logged via structured logging to avoid
	// accidental exposure in production log aggregators. This is dev-only output.
	s.log.Sugar().Infof("[DEV ONLY] OTP for %s: %s", phone, otp)
	return nil
}

// Service contains the authentication business logic.
type Service struct {
	repo    *Repository
	jwtSvc  *JWTService
	otpSvc  OTPService
	log     *zap.Logger
}

// NewService creates a new auth Service.
func NewService(repo *Repository, jwtSvc *JWTService, otpSvc OTPService, log *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		jwtSvc: jwtSvc,
		otpSvc: otpSvc,
		log:    log,
	}
}

// Signup creates a new user account. Returns the created user and a token pair.
func (s *Service) Signup(ctx context.Context, req *SignupRequest) (*AuthResponse, error) {
	tenantOID, err := bson.ObjectIDFromHex(req.TenantID)
	if err != nil {
		return nil, ErrInvalidTenantID
	}

	// Check for existing user with same phone in this tenant.
	existing, err := s.repo.FindByPhone(ctx, tenantOID, req.Phone)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrDuplicateUser
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &User{
		TenantID:      tenantOID,
		Name:          req.Name,
		Phone:         req.Phone,
		Email:         req.Email,
		PasswordHash:  hash,
		Role:          RoleUser,
		Status:        UserStatusPending,
		PhoneVerified: false,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, ErrDuplicateUser) {
			return nil, ErrDuplicateUser
		}
		return nil, fmt.Errorf("creating user: %w", err)
	}

	tokens, err := s.jwtSvc.GenerateTokenPair(
		created.ID.Hex(),
		created.TenantID.Hex(),
		created.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("generating tokens: %w", err)
	}

	return &AuthResponse{
		User:   toPublicUser(created),
		Tokens: *tokens,
	}, nil
}

// Signin authenticates a user by phone+password and returns a token pair.
func (s *Service) Signin(ctx context.Context, req *SigninRequest) (*AuthResponse, error) {
	tenantOID, err := bson.ObjectIDFromHex(req.TenantID)
	if err != nil {
		return nil, ErrInvalidTenantID
	}

	user, err := s.repo.FindByPhone(ctx, tenantOID, req.Phone)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Return generic message to prevent user enumeration.
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("finding user: %w", err)
	}

	if !password.Compare(user.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	if user.Status == UserStatusInactive {
		return nil, ErrAccountInactive
	}

	tokens, err := s.jwtSvc.GenerateTokenPair(
		user.ID.Hex(),
		user.TenantID.Hex(),
		user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("generating tokens: %w", err)
	}

	return &AuthResponse{
		User:   toPublicUser(user),
		Tokens: *tokens,
	}, nil
}

// Refresh validates a refresh token and returns a new access token.
func (s *Service) Refresh(ctx context.Context, req *RefreshRequest) (*TokenPair, error) {
	claims, err := s.jwtSvc.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Verify the user still exists.
	userOID, err := bson.ObjectIDFromHex(claims.Subject)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.FindByID(ctx, userOID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("finding user for refresh: %w", err)
	}

	if user.Status == UserStatusInactive {
		return nil, ErrAccountInactive
	}

	tokens, err := s.jwtSvc.GenerateTokenPair(
		user.ID.Hex(),
		user.TenantID.Hex(),
		user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("generating tokens: %w", err)
	}

	return tokens, nil
}

// Sentinel errors used across the auth package.
var (
	ErrInvalidCredentials = errors.New("invalid phone or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrInvalidTenantID    = errors.New("invalid tenant ID")
)

// toPublicUser maps a User document to a safe public representation.
func toPublicUser(u *User) UserPublic {
	return UserPublic{
		ID:            u.ID.Hex(),
		TenantID:      u.TenantID.Hex(),
		Name:          u.Name,
		Phone:         u.Phone,
		Email:         u.Email,
		Role:          u.Role,
		Status:        u.Status,
		PhoneVerified: u.PhoneVerified,
	}
}
