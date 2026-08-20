// Package vehicles provides HTTP handlers for vehicle endpoints.
// This is the reference handler implementation for the autobee-server boilerplate.
package vehicles

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/autobee/server/pkg/response"
	appvalidator "github.com/autobee/server/pkg/validator"
)

// Handler handles HTTP requests for vehicle endpoints.
type Handler struct {
	svc *Service
	log *zap.Logger
}

// NewHandler creates a new vehicles Handler.
func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Create godoc
//
//	@Summary		Create a vehicle
//	@Description	Creates a new vehicle for the authenticated user. tenantId and userId are derived from the JWT.
//	@Tags			vehicles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateVehicleRequest	true	"Vehicle payload"
//	@Success		201		{object}	response.successResponse{data=VehicleResponse}
//	@Failure		400		{object}	response.errorResponse
//	@Failure		401		{object}	response.errorResponse
//	@Failure		409		{object}	response.errorResponse
//	@Failure		422		{object}	response.errorResponse
//	@Router			/vehicles [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := appvalidator.Validate(req); err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	vehicle, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrDuplicateVehicle) {
			response.Conflict(w, "VEHICLE_ALREADY_EXISTS", "A vehicle with this registration number already exists in this tenant")
			return
		}
		h.log.Error("create vehicle error", zap.Error(err))
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusCreated, vehicle)
}

// List godoc
//
//	@Summary		List vehicles
//	@Description	Returns vehicles. Users see their own; tenant admins see their tenant's; super admins see all.
//	@Tags			vehicles
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int	false	"Page number (default: 1)"
//	@Param			limit	query		int	false	"Page size (default: 20, max: 100)"
//	@Success		200		{object}	response.listResponse{data=[]VehicleResponse}
//	@Failure		401		{object}	response.errorResponse
//	@Router			/vehicles [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)

	result, err := h.svc.List(r.Context(), page, limit)
	if err != nil {
		h.log.Error("list vehicles error", zap.Error(err))
		response.InternalError(w)
		return
	}

	response.List(w, result.Vehicles, response.Pagination{
		Page:  result.Page,
		Limit: result.Limit,
		Total: result.Total,
	})
}

// GetByID godoc
//
//	@Summary		Get a vehicle by ID
//	@Description	Returns a single vehicle. Access is role-checked and tenant-isolated.
//	@Tags			vehicles
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Vehicle ID"
//	@Success		200	{object}	response.successResponse{data=VehicleResponse}
//	@Failure		401	{object}	response.errorResponse
//	@Failure		403	{object}	response.errorResponse
//	@Failure		404	{object}	response.errorResponse
//	@Router			/vehicles/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	vehicle, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidVehicleID):
			response.BadRequest(w, "invalid vehicle ID: must be a 24-character hexadecimal ID")
		case errors.Is(err, ErrVehicleNotFound):
			response.NotFound(w, "VEHICLE_NOT_FOUND", "Vehicle not found")
		case errors.Is(err, response.ErrForbidden):
			response.Forbidden(w, "You do not have access to this vehicle")
		default:
			h.log.Error("get vehicle error", zap.Error(err))
			response.InternalError(w)
		}
		return
	}

	response.JSON(w, http.StatusOK, vehicle)
}

// parseIntQuery extracts an integer from a query parameter with a default fallback.
func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return defaultVal
	}
	return n
}
