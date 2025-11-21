package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/go-chi/chi/v5"
)

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// ResultsServiceInterface defines the interface for results services
type ResultsServiceInterface interface {
	GetDraw(ctx context.Context, contest int) (*models.Draw, error)
	ListDraws(ctx context.Context, limit, offset int) ([]models.Draw, error)
	ImportArtifact(ctx context.Context, artifactID, sheet string) (*services.ImportResult, error)
}

// GetDrawTest is a test version that accepts the interface
func GetDrawTest(resultsSvc ResultsServiceInterface, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contestStr := chi.URLParam(r, "contest")
		if contestStr == "" {
			WriteError(w, http.StatusBadRequest, APIError{
				Code:    "missing_contest",
				Message: "contest parameter is required",
			})
			return
		}

		contest, err := strconv.Atoi(contestStr)
		if err != nil {
			WriteError(w, http.StatusBadRequest, APIError{
				Code:    "invalid_contest",
				Message: "contest must be a valid number",
			})
			return
		}

		logger.Info("Getting draw", "contest", contest)

		draw, err := resultsSvc.GetDraw(r.Context(), contest)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				WriteError(w, http.StatusNotFound, APIError{
					Code:    "draw_not_found",
					Message: "draw not found",
				})
				return
			}
			logger.Error("Failed to get draw", "error", err, "contest", contest)
			WriteError(w, http.StatusInternalServerError, APIError{
				Code:    "get_draw_failed",
				Message: "failed to get draw",
			})
			return
		}

		WriteJSON(w, http.StatusOK, draw)
	}
}

// ListDrawsTest is a test version that accepts the interface
func ListDrawsTest(resultsSvc ResultsServiceInterface, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		limit := 50 // default limit
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
				limit = l
			} else {
				WriteError(w, http.StatusBadRequest, APIError{
					Code:    "invalid_limit",
					Message: "limit must be a number between 1 and 1000",
				})
				return
			}
		}

		offset := 0 // default offset
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			} else {
				WriteError(w, http.StatusBadRequest, APIError{
					Code:    "invalid_offset",
					Message: "offset must be a non-negative number",
				})
				return
			}
		}

		logger.Info("Listing draws", "limit", limit, "offset", offset)

		draws, err := resultsSvc.ListDraws(r.Context(), limit, offset)
		if err != nil {
			logger.Error("Failed to list draws", "error", err)
			WriteError(w, http.StatusInternalServerError, APIError{
				Code:    "list_draws_failed",
				Message: "failed to list draws",
			})
			return
		}

		response := map[string]interface{}{
			"draws":  draws,
			"limit":  limit,
			"offset": offset,
			"count":  len(draws),
		}

		WriteJSON(w, http.StatusOK, response)
	}
}

// createTestRouter creates a test router with the results handlers
func createTestRouter(resultsSvc ResultsServiceInterface, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// For testing, we'll create wrapper handlers that work with the interface
	r.Get("/api/v1/results/{contest}", GetDrawTest(resultsSvc, logger))
	r.Get("/api/v1/results", ListDrawsTest(resultsSvc, logger))
	return r
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

func TestGetDraw_MissingContest(t *testing.T) {
	logger := createTestLogger()
	resultsSvc := &services.ResultsService{} // mock service

	router := createTestRouter(resultsSvc, logger)

	// Test missing contest parameter - use a URL that doesn't match the route pattern
	req := httptest.NewRequest(http.MethodGet, "/api/v1/results/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Chi returns 404 for routes that don't match the pattern
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
	// Since it's a 404 from Chi, not our handler, we don't check the response format
}

func TestGetDraw_InvalidContest(t *testing.T) {
	logger := createTestLogger()
	resultsSvc := &services.ResultsService{} // mock service

	router := createTestRouter(resultsSvc, logger)

	// Test invalid contest parameter (not a number)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/results/abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "invalid_contest" {
		t.Errorf("Expected code 'invalid_contest', got '%s'", response.Code)
	}
	if response.Message != "contest must be a valid number" {
		t.Errorf("Expected message 'contest must be a valid number', got '%s'", response.Message)
	}
}

func TestGetDraw_ValidRequest(t *testing.T) {
	logger := createTestLogger()

	// Create a mock service that doesn't panic
	resultsSvc := &mockResultsService{}

	router := createTestRouter(resultsSvc, logger)

	// Test valid request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/results/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// The mock service returns an error, so we expect internal server error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d (service error), got %d", http.StatusInternalServerError, w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "get_draw_failed" {
		t.Errorf("Expected code 'get_draw_failed', got '%s'", response.Code)
	}
}

// mockResultsService is a simple mock that implements the ResultsService interface
type mockResultsService struct{}

func (m *mockResultsService) GetDraw(ctx context.Context, contest int) (*models.Draw, error) {
	return nil, fmt.Errorf("mock error")
}

func (m *mockResultsService) ListDraws(ctx context.Context, limit, offset int) ([]models.Draw, error) {
	return nil, fmt.Errorf("mock error")
}

func (m *mockResultsService) ImportArtifact(ctx context.Context, artifactID, sheet string) (*services.ImportResult, error) {
	return nil, fmt.Errorf("mock error")
}

func TestListDraws_InvalidLimit(t *testing.T) {
	logger := createTestLogger()
	resultsSvc := &services.ResultsService{} // mock service

	router := createTestRouter(resultsSvc, logger)

	// Test invalid limit parameter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/results?limit=invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "invalid_limit" {
		t.Errorf("Expected code 'invalid_limit', got '%s'", response.Code)
	}
	if response.Message != "limit must be a number between 1 and 1000" {
		t.Errorf("Expected message 'limit must be a number between 1 and 1000', got '%s'", response.Message)
	}
}

func TestListDraws_InvalidOffset(t *testing.T) {
	logger := createTestLogger()
	resultsSvc := &services.ResultsService{} // mock service

	router := createTestRouter(resultsSvc, logger)

	// Test invalid offset parameter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/results?offset=-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "invalid_offset" {
		t.Errorf("Expected code 'invalid_offset', got '%s'", response.Code)
	}
	if response.Message != "offset must be a non-negative number" {
		t.Errorf("Expected message 'offset must be a non-negative number', got '%s'", response.Message)
	}
}

func TestListDraws_ValidRequest(t *testing.T) {
	logger := createTestLogger()

	// Create a mock service that doesn't panic
	resultsSvc := &mockResultsService{}

	router := createTestRouter(resultsSvc, logger)

	// Test valid request with parameters
	req := httptest.NewRequest(http.MethodGet, "/api/v1/results?limit=10&offset=5", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// The mock service returns an error, so we expect internal server error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d (service error), got %d", http.StatusInternalServerError, w.Code)
	}

	var response APIError
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Code != "list_draws_failed" {
		t.Errorf("Expected code 'list_draws_failed', got '%s'", response.Code)
	}
}

func TestListDraws_DefaultParameters(t *testing.T) {
	logger := createTestLogger()

	// Create a mock service that doesn't panic
	resultsSvc := &mockResultsService{}

	router := createTestRouter(resultsSvc, logger)

	// Test request without parameters (should use defaults)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/results", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should fail with service error, but we can verify the request was parsed
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d (service error), got %d", http.StatusInternalServerError, w.Code)
	}
}
