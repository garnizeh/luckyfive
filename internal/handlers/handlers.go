package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/garnizeh/luckyfive/internal/models"
)

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response with request ID
func WriteError(w http.ResponseWriter, r *http.Request, err models.APIError) {
	// Add request ID if available
	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		err.RequestID = reqID
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatusCode())
	json.NewEncoder(w).Encode(err)
}

// WriteValidationError writes a validation error response
func WriteValidationError(w http.ResponseWriter, r *http.Request, err models.ValidationError) {
	// Add request ID if available
	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		err.RequestID = reqID
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatusCode())
	json.NewEncoder(w).Encode(err)
}

// APIError represents an API error response (deprecated: use models.APIError)
type APIError = models.APIError
