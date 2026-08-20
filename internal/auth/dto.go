// Package auth provides authentication DTOs for request/response binding.
package auth

// SignupRequest is the request body for POST /api/v1/auth/signup.
type SignupRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Phone    string `json:"phone"    validate:"required"`
	Email    string `json:"email"    validate:"omitempty,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	TenantID string `json:"tenantId" validate:"required,hexadecimal,len=24"`
}

// SigninRequest is the request body for POST /api/v1/auth/signin.
type SigninRequest struct {
	Phone    string `json:"phone"    validate:"required"`
	Password string `json:"password" validate:"required"`
	TenantID string `json:"tenantId" validate:"required,hexadecimal,len=24"`
}

// RefreshRequest is the request body for POST /api/v1/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// AuthResponse is returned on successful signup or signin.
type AuthResponse struct {
	User   UserPublic `json:"user"`
	Tokens TokenPair  `json:"tokens"`
}

// UserPublic is the safe public projection of a user — no password hash.
type UserPublic struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenantId"`
	Name          string     `json:"name"`
	Phone         string     `json:"phone"`
	Email         string     `json:"email,omitempty"`
	Role          Role       `json:"role"`
	Status        UserStatus `json:"status"`
	PhoneVerified bool       `json:"phoneVerified"`
}
