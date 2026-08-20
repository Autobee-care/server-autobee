// Package users provides HTTP handlers for user endpoints.
package users

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/autobee/server/internal/middleware"
	"github.com/autobee/server/pkg/response"
)

// Handler handles HTTP requests for user endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

// NewHandler creates a new users Handler.
func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// GetMe godoc
//
//	@Summary		Get current user
//	@Description	Returns the profile of the currently authenticated user
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.successResponse{data=UserResponse}
//	@Failure		401	{object}	response.errorResponse
//	@Failure		404	{object}	response.errorResponse
//	@Router			/users/me [get]
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Unauthorized(w, "not authenticated")
		return
	}

	user, err := h.svc.GetMe(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(w, "USER_NOT_FOUND", "User not found")
			return
		}
		h.log.Error("get me error", zap.Error(err))
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, user)
}
