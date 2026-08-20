// Package response provides standardized HTTP response helpers.
package response

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrForbidden is a sentinel error returned by services when access is denied.
// Handlers should map this to a 403 Forbidden response.
var ErrForbidden = errors.New("access forbidden")

// Pagination holds pagination metadata for list responses.
type Pagination struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// successResponse wraps a single data payload.
type successResponse struct {
	Data any `json:"data"`
}

// listResponse wraps a slice payload with pagination.
type listResponse struct {
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// errorDetail is the inner error object.
type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// errorResponse wraps an error detail.
type errorResponse struct {
	Error errorDetail `json:"error"`
}

// JSON writes a 200 OK response with the given data payload.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(successResponse{Data: data})
}

// List writes a 200 OK response with data and pagination metadata.
func List(w http.ResponseWriter, data any, pagination Pagination) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(listResponse{Data: data, Pagination: pagination})
}

// Error writes an error response with the given HTTP status code.
func Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Error: errorDetail{Code: code, Message: message},
	})
}

// ValidationError writes a 422 Unprocessable Entity response.
func ValidationError(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", message)
}

// Unauthorized writes a 401 Unauthorized response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden writes a 403 Forbidden response.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound writes a 404 Not Found response.
func NotFound(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusNotFound, code, message)
}

// Conflict writes a 409 Conflict response.
func Conflict(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusConflict, code, message)
}

// InternalError writes a 500 Internal Server Error response.
// The raw error is never exposed to the client.
func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
}

// BadRequest writes a 400 Bad Request response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}
