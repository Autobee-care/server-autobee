// Package middleware provides HTTP middleware components.
package middleware

import (
	"net/http"

	"github.com/go-chi/cors"

	"github.com/autobee/server/internal/config"
)

// CORS returns a Chi-compatible CORS middleware configured from application settings.
// When credentials are enabled, wildcard origins are not permitted per the CORS spec.
func CORS(cfg *config.CORSConfig) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           300, // maximum value for preflight cache (5 minutes)
	})
}
