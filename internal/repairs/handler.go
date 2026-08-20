// Package repairs provides HTTP handlers for repair endpoints.
package repairs

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/autobee/server/pkg/response"
)

// Handler handles HTTP requests for repair endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

// NewHandler creates a new repairs Handler.
func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Placeholder handler.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, []any{})
}
