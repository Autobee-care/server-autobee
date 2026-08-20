// Package services provides HTTP handlers for vehicle service endpoints.
package services

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/autobee/server/pkg/response"
)

// Handler handles HTTP requests for vehicle service endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

// NewHandler creates a new vehicle services Handler.
func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Placeholder handler.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, []any{})
}
