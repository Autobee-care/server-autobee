// Package users provides request/response DTOs for the users module.
package users

// UpdateUserRequest is the request body for updating a user profile.
type UpdateUserRequest struct {
	Name  string `json:"name"  validate:"omitempty,min=2,max=100"`
	Email string `json:"email" validate:"omitempty,email"`
}
