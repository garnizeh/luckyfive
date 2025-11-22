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
	"github.com/garnizeh/luckyfive/internal/store/configs"
)

// Mock ConfigService for testing
type MockConfigsService struct {
	ListFunc       func(ctx context.Context, limit, offset int64) ([]configs.Config, error)
	CreateFunc     func(ctx context.Context, req services.CreateConfigRequest) (configs.Config, error)
	GetFunc        func(ctx context.Context, id int64) (configs.Config, error)
	GetByNameFunc  func(ctx context.Context, name string) (configs.Config, error)
	UpdateFunc     func(ctx context.Context, id int64, req services.CreateConfigRequest) error
	DeleteFunc     func(ctx context.Context, id int64) error
	SetDefaultFunc func(ctx context.Context, id int64) error
}

func (m *MockConfigsService) List(ctx context.Context, limit, offset int64) ([]configs.Config, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return nil, nil
}

func (m *MockConfigsService) Create(ctx context.Context, req services.CreateConfigRequest) (configs.Config, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return configs.Config{}, nil
}

func (m *MockConfigsService) Get(ctx context.Context, id int64) (configs.Config, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return configs.Config{}, nil
}

func (m *MockConfigsService) GetByName(ctx context.Context, name string) (configs.Config, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, name)
	}
	return configs.Config{}, nil
}

func (m *MockConfigsService) ListByMode(ctx context.Context, mode string, limit, offset int64) ([]configs.Config, error) {
	return nil, nil
}

func (m *MockConfigsService) Update(ctx context.Context, id int64, req services.CreateConfigRequest) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, req)
	}
	return nil
}

func (m *MockConfigsService) Delete(ctx context.Context, id int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockConfigsService) SetDefault(ctx context.Context, id int64) error {
	if m.SetDefaultFunc != nil {
		return m.SetDefaultFunc(ctx, id)
	}
	return nil
}

func (m *MockConfigsService) GetDefault(ctx context.Context, mode string) (configs.Config, error) {
	return configs.Config{}, nil
}

func (m *MockConfigsService) IncrementUsage(ctx context.Context, id int64) error {
	return nil
}

func (m *MockConfigsService) GetPreset(ctx context.Context, name string) (configs.ConfigPreset, error) {
	return configs.ConfigPreset{}, nil
}

func (m *MockConfigsService) ListPresets(ctx context.Context) ([]configs.ConfigPreset, error) {
	return nil, nil
}

func TestListConfigs_ValidRequest(t *testing.T) {
	mockSvc := &MockConfigsService{
		ListFunc: func(ctx context.Context, limit, offset int64) ([]configs.Config, error) {
			return []configs.Config{
				{
					ID:          1,
					Name:        "test-config-1",
					Description: sql.NullString{String: "Test config 1", Valid: true},
					Mode:        "simple",
				},
				{
					ID:          2,
					Name:        "test-config-2",
					Description: sql.NullString{String: "Test config 2", Valid: true},
					Mode:        "advanced",
				},
			}, nil
		},
	}

	handler := ListConfigs(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/configs?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	configs, ok := response["configs"].([]interface{})
	if !ok {
		t.Fatal("Expected configs array in response")
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}
}

func TestCreateConfig_ValidRequest(t *testing.T) {
	mockSvc := &MockConfigsService{
		CreateFunc: func(ctx context.Context, req services.CreateConfigRequest) (configs.Config, error) {
			return configs.Config{
				ID:          1,
				Name:        req.Name,
				Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
				Mode:        req.Mode,
			}, nil
		},
	}

	handler := CreateConfig(mockSvc)

	reqBody := CreateConfigRequest{
		Name:        "test-config",
		Description: "Test configuration",
		Recipe: services.Recipe{
			Version: "1.0",
			Name:    "test-recipe",
			Parameters: services.RecipeParameters{
				Alpha:      0.1,
				Beta:       0.2,
				Gamma:      0.3,
				Delta:      0.4,
				SimPrevMax: 10,
				SimPreds:   5,
			},
		},
		Tags: "test,simulation",
		Mode: "simple",
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/configs", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response configs.Config
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("Expected config ID 1, got %d", response.ID)
	}

	if response.Name != "test-config" {
		t.Errorf("Expected config name 'test-config', got '%s'", response.Name)
	}
}

func TestCreateConfig_InvalidJSON(t *testing.T) {
	mockSvc := &MockConfigsService{}

	handler := CreateConfig(mockSvc)

	req := httptest.NewRequest("POST", "/api/v1/configs", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateConfig_MissingName(t *testing.T) {
	mockSvc := &MockConfigsService{}

	handler := CreateConfig(mockSvc)

	reqBody := CreateConfigRequest{
		Description: "Test configuration",
		Recipe: services.Recipe{
			Version: "1.0",
			Name:    "test-recipe",
		},
		Mode: "simple",
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/configs", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetConfig_ValidID(t *testing.T) {
	mockSvc := &MockConfigsService{
		GetFunc: func(ctx context.Context, id int64) (configs.Config, error) {
			return configs.Config{
				ID:          id,
				Name:        "test-config",
				Description: sql.NullString{String: "Test configuration", Valid: true},
				Mode:        "simple",
			}, nil
		},
	}

	handler := GetConfig(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/configs/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response configs.Config
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("Expected config ID 1, got %d", response.ID)
	}
}

func TestGetConfig_InvalidID(t *testing.T) {
	mockSvc := &MockConfigsService{}

	handler := GetConfig(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/configs/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d. Response body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestGetConfig_NotFound(t *testing.T) {
	mockSvc := &MockConfigsService{
		GetFunc: func(ctx context.Context, id int64) (configs.Config, error) {
			return configs.Config{}, fmt.Errorf("config not found") // Simulate not found error
		},
	}

	handler := GetConfig(mockSvc)

	req := httptest.NewRequest("GET", "/api/v1/configs/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateConfig_ValidRequest(t *testing.T) {
	mockSvc := &MockConfigsService{
		UpdateFunc: func(ctx context.Context, id int64, req services.CreateConfigRequest) error {
			return nil
		},
	}

	handler := UpdateConfig(mockSvc)

	reqBody := UpdateConfigRequest{
		Description: "Updated description",
		Recipe: services.Recipe{
			Version: "1.0",
			Name:    "updated-recipe",
			Parameters: services.RecipeParameters{
				Alpha:      0.2,
				Beta:       0.3,
				Gamma:      0.4,
				Delta:      0.5,
				SimPrevMax: 15,
				SimPreds:   8,
			},
		},
		Tags: "updated,test",
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/configs/1", bytes.NewReader(reqBytes))
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

	if response["message"] != "Config updated successfully" {
		t.Errorf("Expected 'Config updated successfully' message, got '%s'", response["message"])
	}
}

func TestUpdateConfig_InvalidID(t *testing.T) {
	mockSvc := &MockConfigsService{}

	handler := UpdateConfig(mockSvc)

	reqBody := UpdateConfigRequest{
		Description: "Updated description",
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/api/v1/configs/abc", bytes.NewReader(reqBytes))
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

func TestDeleteConfig_ValidID(t *testing.T) {
	mockSvc := &MockConfigsService{
		DeleteFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	handler := DeleteConfig(mockSvc)

	req := httptest.NewRequest("DELETE", "/api/v1/configs/1", nil)
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

	if response["message"] != "Config deleted successfully" {
		t.Errorf("Expected 'Config deleted successfully' message, got '%s'", response["message"])
	}
}

func TestDeleteConfig_InvalidID(t *testing.T) {
	mockSvc := &MockConfigsService{}

	handler := DeleteConfig(mockSvc)

	req := httptest.NewRequest("DELETE", "/api/v1/configs/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSetDefaultConfig_ValidID(t *testing.T) {
	mockSvc := &MockConfigsService{
		SetDefaultFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	handler := SetDefaultConfig(mockSvc)

	req := httptest.NewRequest("POST", "/api/v1/configs/1/set-default", nil)
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

	if response["message"] != "Config set as default successfully" {
		t.Errorf("Expected 'Config set as default successfully' message, got '%s'", response["message"])
	}
}

func TestSetDefaultConfig_InvalidID(t *testing.T) {
	mockSvc := &MockConfigsService{}

	handler := SetDefaultConfig(mockSvc)

	req := httptest.NewRequest("POST", "/api/v1/configs/abc/set-default", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}
