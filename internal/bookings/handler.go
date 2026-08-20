// Package bookings provides HTTP handlers for booking endpoints.
package bookings

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/autobee/server/pkg/response"
)

// Handler handles HTTP requests for booking endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

// NewHandler creates a new bookings Handler.
func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Placeholder handler — implement following the vehicles.Handler pattern.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, []any{})
}
