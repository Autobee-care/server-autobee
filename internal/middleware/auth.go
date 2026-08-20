// Package middleware provides HTTP middleware components.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/autobee/server/internal/auth"
	"github.com/autobee/server/pkg/response"
)

type authContextKey string

const (
	userIDKey   authContextKey = "userID"
	tenantIDKey authContextKey = "tenantID"
	roleKey     authContextKey = "role"
)

// Authenticate validates the JWT Bearer token in the Authorization header.
// On success it stores userID, tenantID, and role in the request context.
// On failure it returns 401 Unauthorized.
func Authenticate(jwtSvc *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				response.Unauthorized(w, "missing or malformed Authorization header")
				return
			}

			claims, err := jwtSvc.ValidateAccessToken(token)
			if err != nil {
				response.Unauthorized(w, "invalid or expired token")
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, claims.Subject)
			ctx = context.WithValue(ctx, tenantIDKey, claims.TenantID)
			ctx = context.WithValue(ctx, roleKey, string(claims.Role))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns a middleware that permits only requests from users
// whose role is in the allowed list. Must be used after Authenticate.
func RequireRole(roles ...auth.Role) func(http.Handler) http.Handler {
	allowed := make(map[auth.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := auth.Role(GetRole(r.Context()))
			if _, ok := allowed[role]; !ok {
				response.Forbidden(w, "you do not have permission to access this resource")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID retrieves the authenticated user's ID from the context.
func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

// GetTenantID retrieves the authenticated user's tenant ID from the context.
func GetTenantID(ctx context.Context) string {
	if v, ok := ctx.Value(tenantIDKey).(string); ok {
		return v
	}
	return ""
}

// GetRole retrieves the authenticated user's role from the context.
func GetRole(ctx context.Context) string {
	if v, ok := ctx.Value(roleKey).(string); ok {
		return v
	}
	return ""
}

// extractBearerToken parses "Bearer <token>" from the Authorization header.
func extractBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
