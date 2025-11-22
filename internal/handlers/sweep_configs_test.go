package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/garnizeh/luckyfive/internal/store/sweeps"
	"github.com/garnizeh/luckyfive/pkg/sweep"
)

// Mock SweepConfigService for testing
type MockSweepConfigService struct {
	ListFunc           func(ctx context.Context, limit, offset int64) ([]sweeps.Sweep, error)
	CreateFunc         func(ctx context.Context, req services.CreateSweepConfigRequest) (sweeps.Sweep, error)
	GetFunc            func(ctx context.Context, id int64) (sweeps.Sweep, error)
	GetByNameFunc      func(ctx context.Context, name string) (sweeps.Sweep, error)
	UpdateFunc         func(ctx context.Context, id int64, req services.CreateSweepConfigRequest) error
	DeleteFunc         func(ctx context.Context, id int64) error
	IncrementUsageFunc func(ctx context.Context, id int64) error
}

func (m *MockSweepConfigService) List(ctx context.Context, limit, offset int64) ([]sweeps.Sweep, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return nil, nil
}

func (m *MockSweepConfigService) Create(ctx context.Context, req services.CreateSweepConfigRequest) (sweeps.Sweep, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return sweeps.Sweep{}, nil
}

func (m *MockSweepConfigService) Get(ctx context.Context, id int64) (sweeps.Sweep, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return sweeps.Sweep{}, nil
}

func (m *MockSweepConfigService) GetByName(ctx context.Context, name string) (sweeps.Sweep, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, name)
	}
	return sweeps.Sweep{}, nil
}

func (m *MockSweepConfigService) Update(ctx context.Context, id int64, req services.CreateSweepConfigRequest) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil
}

func (m *MockSweepConfigService) Delete(ctx context.Context, id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockSweepConfigService) IncrementUsage(ctx context.Context, id int64) error {
	if m.IncrementUsageFunc != nil {
		return m.IncrementUsageFunc(ctx, id)
	}
	return nil
}

func TestListSweepConfigs_ValidRequest(t *testing.T) {
	mockSvc := &MockSweepConfigService{
		ListFunc: func(ctx context.Context, limit, offset int64) ([]sweeps.Sweep, error) {
			return []sweeps.Sweep{
				{
					ID:          1,
					Name:        "test-sweep-1",
					Description: sql.NullString{String: "Test sweep 1", Valid: true},
					ConfigJson:  `{"name":"test-sweep-1"}`,
					CreatedAt:   "2025-01-01 00:00:00",
					UpdatedAt:   "2025-01-01 00:00:00",
					CreatedBy:   sql.NullString{String: "test_user", Valid: true},
					TimesUsed:   sql.NullInt64{Int64: 0, Valid: true},
					LastUsedAt:  sql.NullString{Valid: false},
				},
				{
					ID:          2,
					Name:        "test-sweep-2",
					Description: sql.NullString{String: "Test sweep 2", Valid: true},
					ConfigJson:  `{"name":"test-sweep-2"}`,
					CreatedAt:   "2025-01-02 00:00:00",
					UpdatedAt:   "2025-01-02 00:00:00",
					CreatedBy:   sql.NullString{String: "test_user", Valid: true},
					TimesUsed:   sql.NullInt64{Int64: 0, Valid: true},
					LastUsedAt:  sql.NullString{Valid: false},
				},
			}, nil
		},
	}

	handler := ListSweepConfigs(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/sweep-configs?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	sweepConfigs, ok := response["sweep_configs"].([]interface{})
	if !ok {
		t.Fatal("Expected sweep_configs array in response")
	}

	if len(sweepConfigs) != 2 {
		t.Errorf("Expected 2 sweep configs, got %d", len(sweepConfigs))
	}
}

func TestCreateSweepConfig_ValidRequest(t *testing.T) {
	mockSvc := &MockSweepConfigService{
		CreateFunc: func(ctx context.Context, req services.CreateSweepConfigRequest) (sweeps.Sweep, error) {
			return sweeps.Sweep{
				ID:          1,
				Name:        req.Name,
				Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
				ConfigJson:  `{"name":"test-sweep"}`,
				CreatedAt:   "2025-01-01 00:00:00",
				UpdatedAt:   "2025-01-01 00:00:00",
				CreatedBy:   sql.NullString{String: req.CreatedBy, Valid: true},
				TimesUsed:   sql.NullInt64{Int64: 0, Valid: true},
				LastUsedAt:  sql.NullString{Valid: false},
			}, nil
		},
	}

	handler := CreateSweepConfig(mockSvc)

	reqBody := CreateSweepConfigRequest{
		Name:        "test-sweep",
		Description: "Test sweep configuration",
		Config: sweep.SweepConfig{
			Name:        "test-sweep",
			Description: "Test sweep configuration",
			BaseRecipe: sweep.Recipe{
				Version: "1.0",
				Name:    "advanced",
				Parameters: map[string]interface{}{
					"alpha": 0.1,
				},
			},
			Parameters: []sweep.ParameterSweep{
				{
					Name: "alpha",
					Type: "range",
					Values: sweep.RangeValues{
						Min:  0.0,
						Max:  1.0,
						Step: 0.1,
					},
				},
			},
		},
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/sweep-configs", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response sweeps.Sweep
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("Expected sweep config ID 1, got %d", response.ID)
	}

	if response.Name != "test-sweep" {
		t.Errorf("Expected sweep config name 'test-sweep', got '%s'", response.Name)
	}
}

func TestCreateSweepConfig_InvalidJSON(t *testing.T) {
	mockSvc := &MockSweepConfigService{}

	handler := CreateSweepConfig(mockSvc)

	req := httptest.NewRequest("POST", "/api/v1/sweep-configs", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateSweepConfig_MissingName(t *testing.T) {
	mockSvc := &MockSweepConfigService{}

	handler := CreateSweepConfig(mockSvc)

	reqBody := CreateSweepConfigRequest{
		Description: "Test sweep configuration",
		Config: sweep.SweepConfig{
			Description: "Test sweep configuration",
			BaseRecipe: sweep.Recipe{
				Version: "1.0",
				Name:    "advanced",
			},
		},
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/sweep-configs", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetSweepConfig_ValidID(t *testing.T) {
	mockSvc := &MockSweepConfigService{
		GetFunc: func(ctx context.Context, id int64) (sweeps.Sweep, error) {
			return sweeps.Sweep{
				ID:          id,
				Name:        "test-sweep",
				Description: sql.NullString{String: "Test sweep configuration", Valid: true},
				ConfigJson:  `{"name":"test-sweep"}`,
				CreatedAt:   "2025-01-01 00:00:00",
				UpdatedAt:   "2025-01-01 00:00:00",
				CreatedBy:   sql.NullString{String: "test_user", Valid: true},
				TimesUsed:   sql.NullInt64{Int64: 0, Valid: true},
				LastUsedAt:  sql.NullString{Valid: false},
			}, nil
		},
	}

	handler := GetSweepConfig(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/sweep-configs/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response sweeps.Sweep
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("Expected sweep config ID 1, got %d", response.ID)
	}
}

func TestGetSweepConfig_InvalidID(t *testing.T) {
	mockSvc := &MockSweepConfigService{}

	handler := GetSweepConfig(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/sweep-configs/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d. Response body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestGetSweepConfig_NotFound(t *testing.T) {
	mockSvc := &MockSweepConfigService{
		GetFunc: func(ctx context.Context, id int64) (sweeps.Sweep, error) {
			return sweeps.Sweep{}, fmt.Errorf("sweep config not found") // Simulate not found error
		},
	}

	handler := GetSweepConfig(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/sweep-configs/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateSweepConfig_ValidRequest(t *testing.T) {
	mockSvc := &MockSweepConfigService{
		UpdateFunc: func(ctx context.Context, id int64, req services.CreateSweepConfigRequest) error {
			return nil
		},
	}

	handler := UpdateSweepConfig(mockSvc)

	reqBody := UpdateSweepConfigRequest{
		Name:        "updated-sweep",
		Description: "Updated sweep configuration",
		Config: sweep.SweepConfig{
			Name:        "updated-sweep",
			Description: "Updated sweep configuration",
			BaseRecipe: sweep.Recipe{
				Version: "1.0",
				Name:    "advanced",
				Parameters: map[string]interface{}{
					"alpha": 0.2,
				},
			},
			Parameters: []sweep.ParameterSweep{
				{
					Name: "alpha",
					Type: "range",
					Values: sweep.RangeValues{
						Min:  0.0,
						Max:  1.0,
						Step: 0.1,
					},
				},
			},
		},
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/sweep-configs/1", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "Sweep config updated successfully" {
		t.Errorf("Expected 'Sweep config updated successfully' message, got '%s'", response["message"])
	}
}

func TestUpdateSweepConfig_InvalidID(t *testing.T) {
	mockSvc := &MockSweepConfigService{}

	handler := UpdateSweepConfig(mockSvc)

	reqBody := UpdateSweepConfigRequest{
		Name: "updated-sweep",
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/sweep-configs/abc", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteSweepConfig_ValidID(t *testing.T) {
	mockSvc := &MockSweepConfigService{
		DeleteFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	handler := DeleteSweepConfig(mockSvc)

	req := httptest.NewRequest("DELETE", "/api/v1/sweep-configs/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "Sweep config deleted successfully" {
		t.Errorf("Expected 'Sweep config deleted successfully' message, got '%s'", response["message"])
	}
}

func TestDeleteSweepConfig_InvalidID(t *testing.T) {
	mockSvc := &MockSweepConfigService{}

	handler := DeleteSweepConfig(mockSvc)

	req := httptest.NewRequest("DELETE", "/api/v1/sweep-configs/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}
