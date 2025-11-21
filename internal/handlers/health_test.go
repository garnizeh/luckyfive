package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
)

// mockSystemService is a mock implementation of SystemService for testing
type mockSystemService struct {
	checkHealthFunc func() (*services.HealthStatus, error)
}

func (m *mockSystemService) CheckHealth() (*services.HealthStatus, error) {
	if m.checkHealthFunc != nil {
		return m.checkHealthFunc()
	}
	return &services.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Uptime:    "1h0m0s",
		Services: map[string]string{
			"database": "healthy",
			"api":      "healthy",
		},
	}, nil
}

func TestHealthCheck_Healthy(t *testing.T) {
	// Create mock service returning healthy status
	mockSvc := &mockSystemService{
		checkHealthFunc: func() (*services.HealthStatus, error) {
			return &services.HealthStatus{
				Status:    "healthy",
				Timestamp: "2023-11-21T12:00:00Z",
				Version:   "1.0.0",
				Uptime:    "1h0m0s",
				Services: map[string]string{
					"database": "healthy",
					"api":      "healthy",
				},
			}, nil
		},
	}

	// Create handler
	handler := HealthCheck(mockSvc)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Check response body
	var response services.HealthStatus
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	// Verify response content
	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", response.Version)
	}

	if response.Services["database"] != "healthy" {
		t.Errorf("Expected database service 'healthy', got '%s'", response.Services["database"])
	}
}

func TestHealthCheck_Unhealthy(t *testing.T) {
	// Create mock service returning unhealthy status
	mockSvc := &mockSystemService{
		checkHealthFunc: func() (*services.HealthStatus, error) {
			return &services.HealthStatus{
				Status:    "unhealthy",
				Timestamp: "2023-11-21T12:00:00Z",
				Version:   "1.0.0",
				Uptime:    "1h0m0s",
				Services: map[string]string{
					"database": "unhealthy",
					"api":      "healthy",
				},
			}, nil
		},
	}

	// Create handler
	handler := HealthCheck(mockSvc)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Check response body
	var response services.HealthStatus
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	// Verify response content
	if response.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", response.Status)
	}

	if response.Services["database"] != "unhealthy" {
		t.Errorf("Expected database service 'unhealthy', got '%s'", response.Services["database"])
	}
}

func TestHealthCheck_Error(t *testing.T) {
	// Create mock service returning error
	mockSvc := &mockSystemService{
		checkHealthFunc: func() (*services.HealthStatus, error) {
			return nil, errors.New("database connection failed")
		},
	}

	// Create handler
	handler := HealthCheck(mockSvc)

	// Create request and apply RequestID middleware
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Wrap with RequestID middleware to test request ID inclusion
	middlewareHandler := middleware.RequestID(handler)

	w := httptest.NewRecorder()
	middlewareHandler.ServeHTTP(w, req)

	// Check status code (should be InternalServerError for health_check_failed)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Check response body
	var response models.APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	// Verify response content
	if response.Code != "health_check_failed" {
		t.Errorf("Expected code 'health_check_failed', got '%s'", response.Code)
	}

	if response.Message != "Health check failed" {
		t.Errorf("Expected message 'Health check failed', got '%s'", response.Message)
	}

	// Verify request ID is included
	if response.RequestID == "" {
		t.Error("Expected request ID to be present, but it was empty")
	}
}

func TestHealthCheck_ErrorWithoutRequestID(t *testing.T) {
	// Create mock service returning error
	mockSvc := &mockSystemService{
		checkHealthFunc: func() (*services.HealthStatus, error) {
			return nil, errors.New("service unavailable")
		},
	}

	// Create handler
	handler := HealthCheck(mockSvc)

	// Create request without middleware
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Check response body
	var response models.APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response JSON: %v", err)
	}

	// Verify request ID is empty
	if response.RequestID != "" {
		t.Errorf("Expected empty request ID, got '%s'", response.RequestID)
	}
}
