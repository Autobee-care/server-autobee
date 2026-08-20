// Package middleware provides HTTP middleware components.
package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Logger returns a middleware that logs each HTTP request using Zap.
// It never logs Authorization headers, passwords, JWTs, or tokens.
func Logger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			requestID := GetRequestID(r.Context())
			userID := GetUserID(r.Context())
			tenantID := GetTenantID(r.Context())

			log.Info("http request",
				zap.String("requestId", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.status),
				zap.Duration("duration", time.Since(start)),
				zap.String("userId", userID),
				zap.String("tenantId", tenantID),
				zap.String("remoteAddr", r.RemoteAddr),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func (rw *responseWriter) WriteHeader(status int) {
	if !rw.written {
		rw.status = status
		rw.written = true
		rw.ResponseWriter.WriteHeader(status)
	}
}
