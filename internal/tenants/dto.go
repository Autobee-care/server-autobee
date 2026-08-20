// Package tenants provides request/response DTOs for the tenants module.
package tenants

// CreateTenantRequest is the request body for creating a new tenant.
type CreateTenantRequest struct {
	Name string `json:"name" validate:"required,min=2,max=200"`
}
