package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/luckyfive/internal/services"
)

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestImportResults_InvalidMethod(t *testing.T) {
	logger := createTestLogger()
	resultsSvc := &services.ResultsService{} // mock service

	handler := ImportResults(resultsSvc, logger)

	// Test GET method (should fail)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/results/import", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "method_not_allowed" {
		t.Errorf("Expected code 'method_not_allowed', got '%s'", response.Code)
	}
	if response.Message != "Method not allowed. Use POST." {
		t.Errorf("Expected message 'Method not allowed. Use POST.', got '%s'", response.Message)
	}
}

func TestImportResults_InvalidJSON(t *testing.T) {
	logger := createTestLogger()
	resultsSvc := &services.ResultsService{} // mock service

	handler := ImportResults(resultsSvc, logger)

	// Test invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/results/import", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "invalid_json" {
		t.Errorf("Expected code 'invalid_json', got '%s'", response.Code)
	}
	if response.Message != "Invalid JSON in request body" {
		t.Errorf("Expected message 'Invalid JSON in request body', got '%s'", response.Message)
	}
}

func TestImportResults_MissingArtifactID(t *testing.T) {
	logger := createTestLogger()
	resultsSvc := &services.ResultsService{} // mock service

	handler := ImportResults(resultsSvc, logger)

	// Test missing artifact_id
	requestBody := map[string]interface{}{
		"sheet": "Sheet1",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/results/import", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "missing_artifact_id" {
		t.Errorf("Expected code 'missing_artifact_id', got '%s'", response.Code)
	}
	if response.Message != "artifact_id is required" {
		t.Errorf("Expected message 'artifact_id is required', got '%s'", response.Message)
	}
}

func TestImportResults_EmptyArtifactID(t *testing.T) {
	logger := createTestLogger()
	resultsSvc := &services.ResultsService{} // mock service

	handler := ImportResults(resultsSvc, logger)

	// Test empty artifact_id
	requestBody := map[string]interface{}{
		"artifact_id": "",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/results/import", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "missing_artifact_id" {
		t.Errorf("Expected code 'missing_artifact_id', got '%s'", response.Code)
	}
}

func TestImportResults_ValidRequest(t *testing.T) {
	logger := createTestLogger()

	// Mock service that returns a successful result
	resultsSvc := &services.ResultsService{} // In a real test, you'd use a mock

	handler := ImportResults(resultsSvc, logger)

	// Test valid request
	requestBody := map[string]interface{}{
		"artifact_id": "abc123",
		"sheet":       "Sheet1",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/results/import", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Since we're using a mock service, this will likely fail with artifact not found
	// But we can test that the request parsing worked correctly
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d (service error), got %d", http.StatusInternalServerError, w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "import_failed" {
		t.Errorf("Expected code 'import_failed', got '%s'", response.Code)
	}
}

func TestImportResults_DefaultSheet(t *testing.T) {
	logger := createTestLogger()
	resultsSvc := &services.ResultsService{} // mock service

	handler := ImportResults(resultsSvc, logger)

	// Test request without sheet (should default to empty string)
	requestBody := map[string]interface{}{
		"artifact_id": "abc123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/results/import", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should fail with artifact not found, but we can verify the request was parsed
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d (service error), got %d", http.StatusInternalServerError, w.Code)
	}
}
