package models

import (
	"net/http"
)

// APIError represents a standardized API error response
type APIError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

// NewAPIError creates a new APIError with the given code and message
func NewAPIError(code, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// WithRequestID adds a request ID to the error
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// WithDetails adds additional details to the error
func (e *APIError) WithDetails(details map[string]any) *APIError {
	e.Details = details
	return e
}

// ValidationError represents a validation error with field-specific messages
type ValidationError struct {
	*APIError
	FieldErrors map[string]string `json:"field_errors"`
}

// NewValidationError creates a new validation error
func NewValidationError(message string, fieldErrors map[string]string) *ValidationError {
	return &ValidationError{
		APIError: &APIError{
			Code:    "validation_error",
			Message: message,
		},
		FieldErrors: fieldErrors,
	}
}

// HTTPStatusCode returns the appropriate HTTP status code for an APIError
func (e *APIError) HTTPStatusCode() int {
	switch e.Code {
	case "method_not_allowed":
		return http.StatusMethodNotAllowed
	case "invalid_json", "invalid_form", "no_file", "missing_artifact_id", "missing_contest", "invalid_contest", "invalid_limit", "invalid_offset", "invalid_request", "invalid_simulation_id", "invalid_config_id", "invalid_sweep_config_id":
		return http.StatusBadRequest
	case "upload_failed", "import_failed", "get_draw_failed", "list_draws_failed", "simulation_creation_failed", "simulation_cancel_failed", "config_creation_failed", "config_update_failed", "config_delete_failed", "simulation_not_found", "preset_not_found":
		return http.StatusInternalServerError
	case "not_found", "config_not_found", "sweep_config_not_found", "draw_not_found":
		return http.StatusNotFound
	case "validation_error":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
