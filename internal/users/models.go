// Package users provides the users domain models.
package users

import (
	"time"

	"github.com/autobee/server/internal/auth"
)

// User mirrors the auth.User document. The users package owns the public-facing
// representation; auth owns the authentication concerns.
type User = auth.User

// UserResponse is the safe public projection of a user document.
// passwordHash is never included.
type UserResponse struct {
	ID            string          `json:"id"`
	TenantID      string          `json:"tenantId"`
	Name          string          `json:"name"`
	Phone         string          `json:"phone"`
	Email         string          `json:"email,omitempty"`
	Role          auth.Role       `json:"role"`
	Status        auth.UserStatus `json:"status"`
	PhoneVerified bool            `json:"phoneVerified"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

// ToResponse converts a User document to its safe public representation.
func ToResponse(u *auth.User) *UserResponse {
	return &UserResponse{
		ID:            u.ID.Hex(),
		TenantID:      u.TenantID.Hex(),
		Name:          u.Name,
		Phone:         u.Phone,
		Email:         u.Email,
		Role:          u.Role,
		Status:        u.Status,
		PhoneVerified: u.PhoneVerified,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

