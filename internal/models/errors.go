package models

import (
	"net/http"
)

// APIError represents a standardized API error response
type APIError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
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
func (e *APIError) WithDetails(details map[string]interface{}) *APIError {
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
	case "invalid_json", "invalid_form", "no_file", "missing_artifact_id", "missing_contest", "invalid_contest", "invalid_limit", "invalid_offset":
		return http.StatusBadRequest
	case "upload_failed", "import_failed", "get_draw_failed", "list_draws_failed", "health_check_failed":
		return http.StatusInternalServerError
	case "draw_not_found", "not_found":
		return http.StatusNotFound
	case "validation_error":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
