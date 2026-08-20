// Package tenants provides HTTP handlers for tenant endpoints.
package tenants

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	appvalidator "github.com/autobee/server/pkg/validator"
	"github.com/autobee/server/pkg/response"
)

// Handler handles HTTP requests for tenant endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

// NewHandler creates a new tenants Handler.
func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Create godoc
//
//	@Summary		Create a tenant
//	@Description	Creates a new tenant. Super admin only.
//	@Tags			tenants
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateTenantRequest	true	"Tenant payload"
//	@Success		201		{object}	response.successResponse{data=TenantResponse}
//	@Failure		403		{object}	response.errorResponse
//	@Router			/tenants [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := appvalidator.Validate(req); err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	tenant, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		h.log.Error("create tenant error", zap.Error(err))
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusCreated, tenant)
}

// ListAll godoc
//
//	@Summary		List all tenants
//	@Description	Returns all tenants. Super admin only.
//	@Tags			tenants
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.successResponse{data=[]TenantResponse}
//	@Failure		403	{object}	response.errorResponse
//	@Router			/tenants [get]
func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.svc.ListAll(r.Context())
	if err != nil {
		h.log.Error("list tenants error", zap.Error(err))
		response.InternalError(w)
		return
	}
	response.JSON(w, http.StatusOK, tenants)
}

// GetByID godoc
//
//	@Summary		Get a tenant by ID
//	@Description	Returns a single tenant by its ID. Super admin only.
//	@Tags			tenants
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Tenant ID"
//	@Success		200	{object}	response.successResponse{data=TenantResponse}
//	@Failure		404	{object}	response.errorResponse
//	@Router			/tenants/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	tenant, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInvalidTenantID) {
			response.BadRequest(w, "invalid tenant ID")
			return
		}
		if errors.Is(err, ErrTenantNotFound) {
			response.NotFound(w, "TENANT_NOT_FOUND", "Tenant not found")
			return
		}
		h.log.Error("get tenant error", zap.Error(err))
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, tenant)
}
