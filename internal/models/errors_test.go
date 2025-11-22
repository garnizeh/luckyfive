package models

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNewAPIError(t *testing.T) {
	code := "test_error"
	message := "This is a test error"

	err := NewAPIError(code, message)

	if err.Code != code {
		t.Errorf("Expected code '%s', got '%s'", code, err.Code)
	}

	if err.Message != message {
		t.Errorf("Expected message '%s', got '%s'", message, err.Message)
	}

	if err.RequestID != "" {
		t.Errorf("Expected empty request ID, got '%s'", err.RequestID)
	}

	if err.Details != nil {
		t.Errorf("Expected nil details, got %v", err.Details)
	}
}

func TestAPIError_WithRequestID(t *testing.T) {
	err := NewAPIError("test", "message")
	requestID := "req-123"

	result := err.WithRequestID(requestID)

	// Should return the same instance
	if result != err {
		t.Error("WithRequestID should return the same instance")
	}

	if err.RequestID != requestID {
		t.Errorf("Expected request ID '%s', got '%s'", requestID, err.RequestID)
	}
}

func TestAPIError_WithDetails(t *testing.T) {
	err := NewAPIError("test", "message")
	details := map[string]any{
		"field":  "username",
		"reason": "required",
	}

	result := err.WithDetails(details)

	// Should return the same instance
	if result != err {
		t.Error("WithDetails should return the same instance")
	}

	if !reflect.DeepEqual(err.Details, details) {
		t.Errorf("Expected details %v, got %v", details, err.Details)
	}
}

func TestNewValidationError(t *testing.T) {
	message := "Validation failed"
	fieldErrors := map[string]string{
		"username": "required",
		"email":    "invalid format",
	}

	err := NewValidationError(message, fieldErrors)

	if err.Code != "validation_error" {
		t.Errorf("Expected code 'validation_error', got '%s'", err.Code)
	}

	if err.Message != message {
		t.Errorf("Expected message '%s', got '%s'", message, err.Message)
	}

	if !reflect.DeepEqual(err.FieldErrors, fieldErrors) {
		t.Errorf("Expected field errors %v, got %v", fieldErrors, err.FieldErrors)
	}

	if err.RequestID != "" {
		t.Errorf("Expected empty request ID, got '%s'", err.RequestID)
	}

	if err.Details != nil {
		t.Errorf("Expected nil details, got %v", err.Details)
	}
}

func TestAPIError_HTTPStatusCode_MethodNotAllowed(t *testing.T) {
	err := NewAPIError("method_not_allowed", "Method not allowed")
	status := err.HTTPStatusCode()

	if status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, status)
	}
}

func TestAPIError_HTTPStatusCode_BadRequest(t *testing.T) {
	badRequestCodes := []string{
		"invalid_json",
		"invalid_form",
		"no_file",
		"missing_artifact_id",
		"missing_contest",
		"invalid_contest",
		"invalid_limit",
		"invalid_offset",
		"validation_error",
	}

	for _, code := range badRequestCodes {
		t.Run(code, func(t *testing.T) {
			err := NewAPIError(code, "Bad request")
			status := err.HTTPStatusCode()

			if status != http.StatusBadRequest {
				t.Errorf("Expected status %d for code '%s', got %d", http.StatusBadRequest, code, status)
			}
		})
	}
}

func TestAPIError_HTTPStatusCode_InternalServerError(t *testing.T) {
	internalErrorCodes := []string{
		"upload_failed",
		"import_failed",
		"get_draw_failed",
		"list_draws_failed",
		"health_check_failed",
		"unknown_error", // default case
	}

	for _, code := range internalErrorCodes {
		t.Run(code, func(t *testing.T) {
			err := NewAPIError(code, "Internal error")
			status := err.HTTPStatusCode()

			if status != http.StatusInternalServerError {
				t.Errorf("Expected status %d for code '%s', got %d", http.StatusInternalServerError, code, status)
			}
		})
	}
}

func TestAPIError_HTTPStatusCode_NotFound(t *testing.T) {
	notFoundCodes := []string{
		"draw_not_found",
		"not_found",
	}

	for _, code := range notFoundCodes {
		t.Run(code, func(t *testing.T) {
			err := NewAPIError(code, "Not found")
			status := err.HTTPStatusCode()

			if status != http.StatusNotFound {
				t.Errorf("Expected status %d for code '%s', got %d", http.StatusNotFound, code, status)
			}
		})
	}
}

func TestAPIError_HTTPStatusCode_Default(t *testing.T) {
	err := NewAPIError("some_unknown_code", "Unknown error")
	status := err.HTTPStatusCode()

	if status != http.StatusInternalServerError {
		t.Errorf("Expected default status %d, got %d", http.StatusInternalServerError, status)
	}
}
