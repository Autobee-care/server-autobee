// Package auth provides authentication types and constants.
package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Role defines the available user roles in the system.
type Role string

const (
	// RoleSuperAdmin has cross-tenant access to all resources.
	RoleSuperAdmin Role = "super_admin"
	// RoleTenantAdmin has access to all resources within their tenant.
	RoleTenantAdmin Role = "tenant_admin"
	// RoleUser has access only to their own resources.
	RoleUser Role = "user"
)

// UserStatus defines the lifecycle state of a user account.
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusPending  UserStatus = "pending"
)

// TokenType distinguishes access tokens from refresh tokens.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// User is the stored user document in MongoDB.
type User struct {
	ID            bson.ObjectID `bson:"_id,omitempty"`
	TenantID      bson.ObjectID `bson:"tenantId"`
	Name          string             `bson:"name"`
	Phone         string             `bson:"phone"`
	Email         string             `bson:"email,omitempty"`
	PasswordHash  string             `bson:"passwordHash"`
	Role          Role               `bson:"role"`
	Status        UserStatus         `bson:"status"`
	PhoneVerified bool               `bson:"phoneVerified"`
	CreatedAt     time.Time          `bson:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"`
}

// TokenPair holds an access token and a refresh token.
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Claims holds the parsed JWT payload for use in application context.
type Claims struct {
	Subject  string // user ID
	TenantID string
	Role     Role
	Type     TokenType
	JTI      string
}
