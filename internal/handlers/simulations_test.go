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
	"github.com/garnizeh/luckyfive/internal/store/configs"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
)

// Mock interfaces for testing
type MockConfigService struct {
	GetPresetFunc func(ctx context.Context, name string) (configs.ConfigPreset, error)
	CreateFunc    func(ctx context.Context, req services.CreateConfigRequest) (configs.Config, error)
}

func (m *MockConfigService) GetPreset(ctx context.Context, name string) (configs.ConfigPreset, error) {
	if m.GetPresetFunc != nil {
		return m.GetPresetFunc(ctx, name)
	}
	return configs.ConfigPreset{}, nil
}

func (m *MockConfigService) Create(ctx context.Context, req services.CreateConfigRequest) (configs.Config, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return configs.Config{}, nil
}

func (m *MockConfigService) Get(ctx context.Context, id int64) (configs.Config, error) {
	return configs.Config{}, nil
}

func (m *MockConfigService) GetByName(ctx context.Context, name string) (configs.Config, error) {
	return configs.Config{}, nil
}

func (m *MockConfigService) List(ctx context.Context, limit, offset int64) ([]configs.Config, error) {
	return nil, nil
}

func (m *MockConfigService) ListByMode(ctx context.Context, mode string, limit, offset int64) ([]configs.Config, error) {
	return nil, nil
}

func (m *MockConfigService) Update(ctx context.Context, id int64, req services.CreateConfigRequest) error {
	return nil
}

func (m *MockConfigService) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockConfigService) SetDefault(ctx context.Context, id int64) error {
	return nil
}

func (m *MockConfigService) GetDefault(ctx context.Context, mode string) (configs.Config, error) {
	return configs.Config{}, nil
}

func (m *MockConfigService) IncrementUsage(ctx context.Context, id int64) error {
	return nil
}

func (m *MockConfigService) ListPresets(ctx context.Context) ([]configs.ConfigPreset, error) {
	return nil, nil
}

type MockSimulationService struct {
	CreateSimulationFunc  func(ctx context.Context, req services.CreateSimulationRequest) (*simulations.Simulation, error)
	GetSimulationFunc     func(ctx context.Context, id int64) (*simulations.Simulation, error)
	ListSimulationsFunc   func(ctx context.Context, limit, offset int) ([]simulations.Simulation, error)
	CancelSimulationFunc  func(ctx context.Context, id int64) error
	GetContestResultsFunc func(ctx context.Context, simulationID int64, limit, offset int) ([]simulations.SimulationContestResult, error)
	ExecuteSimulationFunc func(ctx context.Context, simID int64) error
}

func (m *MockSimulationService) CreateSimulation(ctx context.Context, req services.CreateSimulationRequest) (*simulations.Simulation, error) {
	if m.CreateSimulationFunc != nil {
		return m.CreateSimulationFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockSimulationService) GetSimulation(ctx context.Context, id int64) (*simulations.Simulation, error) {
	if m.GetSimulationFunc != nil {
		return m.GetSimulationFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSimulationService) ListSimulations(ctx context.Context, limit, offset int) ([]simulations.Simulation, error) {
	if m.ListSimulationsFunc != nil {
		return m.ListSimulationsFunc(ctx, limit, offset)
	}
	return nil, nil
}

func (m *MockSimulationService) CancelSimulation(ctx context.Context, id int64) error {
	if m.CancelSimulationFunc != nil {
		return m.CancelSimulationFunc(ctx, id)
	}
	return nil
}

func (m *MockSimulationService) GetContestResults(ctx context.Context, simulationID int64, limit, offset int) ([]simulations.SimulationContestResult, error) {
	if m.GetContestResultsFunc != nil {
		return m.GetContestResultsFunc(ctx, simulationID, limit, offset)
	}
	return nil, nil
}

func (m *MockSimulationService) ExecuteSimulation(ctx context.Context, simID int64) error {
	if m.ExecuteSimulationFunc != nil {
		return m.ExecuteSimulationFunc(ctx, simID)
	}
	return nil
}

func TestSimpleSimulation_ValidRequest(t *testing.T) {
	// Mock services
	mockConfigSvc := &MockConfigService{
		GetPresetFunc: func(ctx context.Context, name string) (configs.ConfigPreset, error) {
			return configs.ConfigPreset{
				Name:       "test-preset",
				RecipeJson: `{"version":"1.0","name":"test","parameters":{"alpha":0.1,"beta":0.2,"gamma":0.3,"delta":0.4,"sim_prev_max":10,"sim_preds":5}}`,
			}, nil
		},
	}

	mockSimSvc := &MockSimulationService{
		CreateSimulationFunc: func(ctx context.Context, req services.CreateSimulationRequest) (*simulations.Simulation, error) {
			return &simulations.Simulation{
				ID:     1,
				Status: "completed",
			}, nil
		},
	}

	// Create handler
	handler := SimpleSimulation(mockConfigSvc, mockSimSvc)

	// Create request
	reqBody := map[string]any{
		"preset":        "test-preset",
		"start_contest": 1000,
		"end_contest":   1010,
		"async":         false,
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/simulations/simple", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check response
	var response simulations.Simulation
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("Expected simulation ID 1, got %d", response.ID)
	}
}

func TestSimpleSimulation_InvalidJSON(t *testing.T) {
	mockConfigSvc := &MockConfigService{}
	mockSimSvc := &MockSimulationService{}

	handler := SimpleSimulation(mockConfigSvc, mockSimSvc)

	req := httptest.NewRequest("POST", "/api/v1/simulations/simple", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSimpleSimulation_MissingPreset(t *testing.T) {
	mockConfigSvc := &MockConfigService{}
	mockSimSvc := &MockSimulationService{}

	handler := SimpleSimulation(mockConfigSvc, mockSimSvc)

	reqBody := map[string]any{
		"start_contest": 1000,
		"end_contest":   1010,
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/simulations/simple", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetSimulation_ValidID(t *testing.T) {
	mockSimSvc := &MockSimulationService{
		GetSimulationFunc: func(ctx context.Context, id int64) (*simulations.Simulation, error) {
			return &simulations.Simulation{
				ID:     1,
				Status: "completed",
			}, nil
		},
	}

	handler := GetSimulation(mockSimSvc)

	req := httptest.NewRequest("GET", "/api/v1/simulations/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response simulations.Simulation
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("Expected simulation ID 1, got %d", response.ID)
	}
}

func TestGetSimulation_InvalidID(t *testing.T) {
	mockSimSvc := &MockSimulationService{}

	handler := GetSimulation(mockSimSvc)

	req := httptest.NewRequest("GET", "/api/v1/simulations/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListSimulations_ValidRequest(t *testing.T) {
	mockSimSvc := &MockSimulationService{
		ListSimulationsFunc: func(ctx context.Context, limit, offset int) ([]simulations.Simulation, error) {
			return []simulations.Simulation{
				{ID: 1, Status: "completed"},
				{ID: 2, Status: "running"},
			}, nil
		},
	}

	handler := ListSimulations(mockSimSvc)

	req := httptest.NewRequest("GET", "/api/v1/simulations?limit=10&offset=0", nil)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	sims, ok := response["simulations"].([]any)
	if !ok {
		t.Fatal("Expected simulations array in response")
	}

	if len(sims) != 2 {
		t.Errorf("Expected 2 simulations, got %d", len(sims))
	}
}

func TestCancelSimulation_ValidID(t *testing.T) {
	mockSimSvc := &MockSimulationService{
		CancelSimulationFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	handler := CancelSimulation(mockSimSvc)

	req := httptest.NewRequest("POST", "/api/v1/simulations/1/cancel", nil)
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

	if response["message"] != "Simulation cancelled" {
		t.Errorf("Expected 'Simulation cancelled' message, got '%s'", response["message"])
	}
}

func TestAdvancedSimulation_ValidRequest(t *testing.T) {
	// Mock services
	mockConfigSvc := &MockConfigService{}
	mockSimSvc := &MockSimulationService{
		CreateSimulationFunc: func(ctx context.Context, req services.CreateSimulationRequest) (*simulations.Simulation, error) {
			return &simulations.Simulation{
				ID:     1,
				Status: "completed",
			}, nil
		},
	}

	// Create handler
	handler := AdvancedSimulation(mockConfigSvc, mockSimSvc)

	// Create request with full recipe
	reqBody := map[string]any{
		"recipe": map[string]any{
			"version": "1.0",
			"name":    "advanced-test",
			"parameters": map[string]any{
				"alpha":        0.1,
				"beta":         0.2,
				"gamma":        0.3,
				"delta":        0.4,
				"sim_prev_max": 10,
				"sim_preds":    5,
			},
		},
		"start_contest": 1000,
		"end_contest":   1010,
		"async":         false,
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/simulations/advanced", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check response
	var response map[string]any
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["simulation_id"].(float64) != 1 {
		t.Errorf("Expected simulation_id 1, got %v", response["simulation_id"])
	}

	sim, ok := response["simulation"].(map[string]any)
	if !ok {
		t.Fatal("Expected simulation object in response")
	}

	if sim["id"].(float64) != 1 {
		t.Errorf("Expected simulation ID 1, got %v", sim["id"])
	}
}

func TestAdvancedSimulation_InvalidJSON(t *testing.T) {
	mockConfigSvc := &MockConfigService{}
	mockSimSvc := &MockSimulationService{}

	handler := AdvancedSimulation(mockConfigSvc, mockSimSvc)

	req := httptest.NewRequest("POST", "/api/v1/simulations/advanced", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAdvancedSimulation_InvalidRecipe(t *testing.T) {
	mockConfigSvc := &MockConfigService{}
	mockSimSvc := &MockSimulationService{}

	handler := AdvancedSimulation(mockConfigSvc, mockSimSvc)

	// Invalid recipe - missing required parameters
	reqBody := map[string]any{
		"recipe": map[string]any{
			"version": "1.0",
			"name":    "invalid-test",
			"parameters": map[string]any{
				"alpha": 0.1,
				// missing beta, gamma, delta, sim_prev_max, sim_preds
			},
		},
		"start_contest": 1000,
		"end_contest":   1010,
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/simulations/advanced", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAdvancedSimulation_SaveAsConfig(t *testing.T) {
	configCreated := false
	mockConfigSvc := &MockConfigService{
		CreateFunc: func(ctx context.Context, req services.CreateConfigRequest) (configs.Config, error) {
			configCreated = true
			return configs.Config{
				ID:   1,
				Name: req.Name,
			}, nil
		},
	}
	mockSimSvc := &MockSimulationService{
		CreateSimulationFunc: func(ctx context.Context, req services.CreateSimulationRequest) (*simulations.Simulation, error) {
			return &simulations.Simulation{
				ID:     1,
				Status: "completed",
			}, nil
		},
	}

	handler := AdvancedSimulation(mockConfigSvc, mockSimSvc)

	reqBody := map[string]any{
		"recipe": map[string]any{
			"version": "1.0",
			"name":    "config-test",
			"parameters": map[string]any{
				"alpha":        0.1,
				"beta":         0.2,
				"gamma":        0.3,
				"delta":        0.4,
				"sim_prev_max": 10,
				"sim_preds":    5,
			},
		},
		"start_contest":      1000,
		"end_contest":        1010,
		"save_as_config":     true,
		"config_name":        "my-config",
		"config_description": "Test config",
		"async":              false,
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/simulations/advanced", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	if !configCreated {
		t.Error("Expected config to be created, but it wasn't")
	}
}

func TestGetContestResults_ValidRequest(t *testing.T) {
	mockSimSvc := &MockSimulationService{
		GetContestResultsFunc: func(ctx context.Context, simulationID int64, limit, offset int) ([]simulations.SimulationContestResult, error) {
			return []simulations.SimulationContestResult{
				{
					SimulationID:  simulationID,
					Contest:       1000,
					ActualNumbers: "1,2,3,4,5",
					BestHits:      3,
				},
				{
					SimulationID:  simulationID,
					Contest:       1001,
					ActualNumbers: "6,7,8,9,10",
					BestHits:      2,
				},
			}, nil
		},
	}

	handler := GetContestResults(mockSimSvc)

	req := httptest.NewRequest("GET", "/api/v1/simulations/1/results?limit=50&offset=0", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["simulation_id"].(float64) != 1 {
		t.Errorf("Expected simulation_id 1, got %v", response["simulation_id"])
	}

	results, ok := response["results"].([]any)
	if !ok {
		t.Fatal("Expected results array in response")
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestGetContestResults_InvalidID(t *testing.T) {
	mockSimSvc := &MockSimulationService{}

	handler := GetContestResults(mockSimSvc)

	req := httptest.NewRequest("GET", "/api/v1/simulations/abc/results", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}
