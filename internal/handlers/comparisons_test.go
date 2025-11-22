package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/garnizeh/luckyfive/internal/store/comparisons"
)

// Mock implementation of ComparisonServicer for testing
type mockComparisonService struct {
	compareFunc         func(ctx context.Context, req services.CompareRequest) (*services.ComparisonResult, error)
	getComparisonFunc   func(ctx context.Context, id int64) (*services.ComparisonResult, error)
	listComparisonsFunc func(ctx context.Context, limit, offset int) ([]comparisons.Comparison, error)
}

func (m *mockComparisonService) Compare(ctx context.Context, req services.CompareRequest) (*services.ComparisonResult, error) {
	if m.compareFunc != nil {
		return m.compareFunc(ctx, req)
	}
	return &services.ComparisonResult{ID: 1, Name: "Test"}, nil
}

func (m *mockComparisonService) GetComparison(ctx context.Context, id int64) (*services.ComparisonResult, error) {
	if m.getComparisonFunc != nil {
		return m.getComparisonFunc(ctx, id)
	}
	// Temporarily always return success
	return &services.ComparisonResult{ID: id, Name: "Test"}, nil
}

func (m *mockComparisonService) ListComparisons(ctx context.Context, limit, offset int) ([]comparisons.Comparison, error) {
	if m.listComparisonsFunc != nil {
		return m.listComparisonsFunc(ctx, limit, offset)
	}
	return []comparisons.Comparison{
		{ID: 1, Name: "Comp1"},
		{ID: 2, Name: "Comp2"},
	}, nil
}

func TestCreateComparison_Success(t *testing.T) {
	mockSvc := &mockComparisonService{
		compareFunc: func(ctx context.Context, req services.CompareRequest) (*services.ComparisonResult, error) {
			return &services.ComparisonResult{
				ID:   1,
				Name: req.Name,
			}, nil
		},
	}

	handler := CreateComparison(mockSvc)

	reqBody := services.CompareRequest{
		Name:          "Test Comparison",
		SimulationIDs: []int64{1, 2},
		Metrics:       []string{"quina_rate"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/comparisons", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response services.ComparisonResult
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("expected ID 1, got %d", response.ID)
	}
	if response.Name != "Test Comparison" {
		t.Errorf("expected name 'Test Comparison', got %s", response.Name)
	}
}

func TestCreateComparison_InvalidJSON(t *testing.T) {
	mockSvc := &mockComparisonService{}
	handler := CreateComparison(mockSvc)

	req := httptest.NewRequest("POST", "/api/v1/comparisons", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetComparison_Success(t *testing.T) {
	mockSvc := &mockComparisonService{
		getComparisonFunc: func(ctx context.Context, id int64) (*services.ComparisonResult, error) {
			return &services.ComparisonResult{
				ID:   id,
				Name: "Test Comparison",
			}, nil
		},
	}

	handler := GetComparison(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/comparisons/123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response services.ComparisonResult
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.ID != 123 {
		t.Errorf("expected ID 123, got %d", response.ID)
	}
}

func TestGetComparison_InvalidID(t *testing.T) {
	mockSvc := &mockComparisonService{}
	handler := GetComparison(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/comparisons/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetComparison_ZeroID(t *testing.T) {
	mockSvc := &mockComparisonService{
		getComparisonFunc: func(ctx context.Context, id int64) (*services.ComparisonResult, error) {
			if id != 0 {
				t.Errorf("expected ID 0, got %d", id)
			}
			return &services.ComparisonResult{
				ID:   0,
				Name: "Zero ID Comparison",
			}, nil
		},
	}

	handler := GetComparison(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/comparisons/0", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "0")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response services.ComparisonResult
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.ID != 0 {
		t.Errorf("expected ID 0, got %d", response.ID)
	}
	if response.Name != "Zero ID Comparison" {
		t.Errorf("expected name 'Zero ID Comparison', got %s", response.Name)
	}
}

func TestListComparisons_Success(t *testing.T) {
	mockSvc := &mockComparisonService{}
	handler := ListComparisons(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/comparisons?limit=10&offset=5", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["limit"] != 10.0 {
		t.Errorf("expected limit 10, got %v", response["limit"])
	}
	if response["offset"] != 5.0 {
		t.Errorf("expected offset 5, got %v", response["offset"])
	}

	comparisons, ok := response["comparisons"].([]any)
	if !ok {
		t.Fatal("expected comparisons array")
	}
	if len(comparisons) != 2 {
		t.Errorf("expected 2 comparisons, got %d", len(comparisons))
	}
}

func TestListComparisons_DefaultParams(t *testing.T) {
	mockSvc := &mockComparisonService{}
	handler := ListComparisons(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/comparisons", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["limit"] != 50.0 {
		t.Errorf("expected default limit 50, got %v", response["limit"])
	}
	if response["offset"] != 0.0 {
		t.Errorf("expected default offset 0, got %v", response["offset"])
	}
}
