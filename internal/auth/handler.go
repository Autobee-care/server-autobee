// Package auth provides HTTP handlers for authentication endpoints.
package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	appvalidator "github.com/autobee/server/pkg/validator"
	"github.com/autobee/server/pkg/response"
)

// Handler handles HTTP requests for auth endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

// NewHandler creates a new auth Handler.
func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Signup godoc
//
//	@Summary		Register a new user
//	@Description	Creates a new user account within a tenant
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		SignupRequest	true	"Signup payload"
//	@Success		201		{object}	response.successResponse{data=AuthResponse}
//	@Failure		400		{object}	response.errorResponse
//	@Failure		409		{object}	response.errorResponse
//	@Failure		422		{object}	response.errorResponse
//	@Router			/auth/signup [post]
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := appvalidator.Validate(req); err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	res, err := h.svc.Signup(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, res)
}

// Signin godoc
//
//	@Summary		Authenticate a user
//	@Description	Authenticates credentials and returns a JWT token pair
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		SigninRequest	true	"Signin payload"
//	@Success		200		{object}	response.successResponse{data=AuthResponse}
//	@Failure		400		{object}	response.errorResponse
//	@Failure		401		{object}	response.errorResponse
//	@Failure		422		{object}	response.errorResponse
//	@Router			/auth/signin [post]
func (h *Handler) Signin(w http.ResponseWriter, r *http.Request) {
	var req SigninRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := appvalidator.Validate(req); err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	res, err := h.svc.Signin(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// Refresh godoc
//
//	@Summary		Refresh access token
//	@Description	Issues a new access token using a valid refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RefreshRequest	true	"Refresh token payload"
//	@Success		200		{object}	response.successResponse{data=TokenPair}
//	@Failure		400		{object}	response.errorResponse
//	@Failure		401		{object}	response.errorResponse
//	@Router			/auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := appvalidator.Validate(req); err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	tokens, err := h.svc.Refresh(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

// handleServiceError maps auth service errors to appropriate HTTP responses.
func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidTenantID):
		response.BadRequest(w, "invalid tenantId format: must be a 24-character hexadecimal ID")
	case errors.Is(err, ErrDuplicateUser):
		response.Conflict(w, "USER_ALREADY_EXISTS", "A user with this phone number already exists in this tenant")
	case errors.Is(err, ErrInvalidCredentials):
		response.Unauthorized(w, "Invalid phone number or password")
	case errors.Is(err, ErrInvalidToken):
		response.Unauthorized(w, "Invalid or expired token")
	case errors.Is(err, ErrAccountInactive):
		response.Unauthorized(w, "Account is inactive")
	default:
		h.log.Error("auth handler error", zap.Error(err))
		response.InternalError(w)
	}
}
