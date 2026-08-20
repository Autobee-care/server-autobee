// Package users provides business logic for the users module.
package users

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

// Service contains the users business logic.
type Service struct {
	repo *Repository
	log  *zap.Logger
}

// NewService creates a new users Service.
func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

// ErrInvalidUserID is returned when an invalid hex string is provided as a user ID.
var ErrInvalidUserID = errors.New("invalid user ID")

// GetMe retrieves the current authenticated user's profile.
func (s *Service) GetMe(ctx context.Context, userID string) (*UserResponse, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	user, err := s.repo.FindByID(ctx, oid)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("getting user: %w", err)
	}

	return ToResponse(user), nil
}
