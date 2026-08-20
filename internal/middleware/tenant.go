// Package middleware provides HTTP middleware components.
package middleware

import (
	"net/http"
)

// EnforceTenantContext is a no-op pass-through middleware that documents
// the tenant isolation strategy: tenant context is derived from the JWT
// (set by Authenticate), not from client-supplied query/body parameters.
//
// Individual handlers and services are responsible for using GetTenantID(ctx)
// rather than trusting any client-provided tenantId field.
// This middleware exists to make the architectural intent explicit in routes.go.
func EnforceTenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Tenant context is already set in the context by the Authenticate middleware.
		// Services MUST call middleware.GetTenantID(ctx) to obtain the tenant ID.
		next.ServeHTTP(w, r)
	})
}
