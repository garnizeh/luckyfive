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
	"github.com/garnizeh/luckyfive/internal/store/sweep_execution"
)

// Mock implementation of SweepServicer for testing
type mockSweepService struct {
	createSweepFunc         func(ctx context.Context, req services.CreateSweepRequest) (*sweep_execution.SweepJob, error)
	getSweepStatusFunc      func(ctx context.Context, sweepID int64) (*services.SweepStatus, error)
	updateSweepProgressFunc func(ctx context.Context, sweepID int64) error
	findBestFunc            func(ctx context.Context, sweepID int64, metric string) (*services.BestConfiguration, error)
	getVisualizationFunc    func(ctx context.Context, sweepID int64, metrics []string) (*services.VisualizationData, error)
}

func (m *mockSweepService) CreateSweep(ctx context.Context, req services.CreateSweepRequest) (*sweep_execution.SweepJob, error) {
	if m.createSweepFunc != nil {
		return m.createSweepFunc(ctx, req)
	}
	return &sweep_execution.SweepJob{
		ID:                1,
		Name:              "Test Sweep",
		Description:       sql.NullString{String: "Test sweep description", Valid: true},
		SweepConfigJson:   "{}",
		BaseContestRange:  "1-100",
		Status:            "pending",
		TotalCombinations: 10,
	}, nil
}

func (m *mockSweepService) GetSweepStatus(ctx context.Context, sweepID int64) (*services.SweepStatus, error) {
	if m.getSweepStatusFunc != nil {
		return m.getSweepStatusFunc(ctx, sweepID)
	}
	return &services.SweepStatus{
		Sweep: sweep_execution.SweepJob{
			ID:                sweepID,
			Name:              "Test Sweep",
			Status:            "running",
			TotalCombinations: 10,
		},
		Total:       10,
		Completed:   5,
		Running:     3,
		Failed:      1,
		Pending:     1,
		Simulations: []sweep_execution.GetSweepSimulationDetailsRow{},
	}, nil
}

func (m *mockSweepService) UpdateSweepProgress(ctx context.Context, sweepID int64) error {
	if m.updateSweepProgressFunc != nil {
		return m.updateSweepProgressFunc(ctx, sweepID)
	}
	return nil
}

func (m *mockSweepService) FindBest(ctx context.Context, sweepID int64, metric string) (*services.BestConfiguration, error) {
	if m.findBestFunc != nil {
		return m.findBestFunc(ctx, sweepID, metric)
	}
	return &services.BestConfiguration{
		SweepID:      sweepID,
		SimulationID: 1,
		Metrics: map[string]float64{
			"quina_rate": 0.05,
			"avg_hits":   2.3,
		},
		Rank:       1,
		Percentile: 100.0,
		VariationParams: map[string]interface{}{
			"alpha": 0.1,
			"beta":  0.2,
		},
	}, nil
}

func (m *mockSweepService) GetVisualizationData(ctx context.Context, sweepID int64, metrics []string) (*services.VisualizationData, error) {
	if m.getVisualizationFunc != nil {
		return m.getVisualizationFunc(ctx, sweepID, metrics)
	}
	return &services.VisualizationData{
		SweepID:    sweepID,
		Parameters: []string{"alpha", "beta"},
		Metrics:    metrics,
		DataPoints: []services.VisualizationDataPoint{
			{
				Params: map[string]interface{}{
					"alpha": 0.1,
					"beta":  0.2,
				},
				Metrics: map[string]float64{
					"quina_rate": 0.05,
					"avg_hits":   2.3,
				},
			},
		},
	}, nil
}

func TestCreateSweep_Success(t *testing.T) {
	mockSvc := &mockSweepService{
		createSweepFunc: func(ctx context.Context, req services.CreateSweepRequest) (*sweep_execution.SweepJob, error) {
			// Verify request parameters
			if req.Name != "Test Sweep" {
				t.Errorf("expected name 'Test Sweep', got %s", req.Name)
			}
			if req.StartContest != 1 {
				t.Errorf("expected start_contest 1, got %d", req.StartContest)
			}
			if req.EndContest != 100 {
				t.Errorf("expected end_contest 100, got %d", req.EndContest)
			}
			return &sweep_execution.SweepJob{
				ID:                1,
				Name:              req.Name,
				Status:            "pending",
				TotalCombinations: 10,
			}, nil
		},
	}

	reqBody := services.CreateSweepRequest{
		Name:         "Test Sweep",
		Description:  "Test sweep description",
		StartContest: 1,
		EndContest:   100,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/sweeps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler := CreateSweep(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	var response sweep_execution.SweepJob
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("expected ID 1, got %d", response.ID)
	}

	if response.Name != "Test Sweep" {
		t.Errorf("expected name 'Test Sweep', got %s", response.Name)
	}
}

func TestGetSweep_Success(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sweeps/1", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweep(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response sweep_execution.SweepJob
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != 1 {
		t.Errorf("expected ID 1, got %d", response.ID)
	}
}

func TestGetSweepStatus_Success(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sweeps/1/status", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepStatus(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response services.SweepStatus
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Sweep.ID != 1 {
		t.Errorf("expected sweep ID 1, got %d", response.Sweep.ID)
	}

	if response.Total != 10 {
		t.Errorf("expected total 10, got %d", response.Total)
	}

	if response.Completed != 5 {
		t.Errorf("expected completed 5, got %d", response.Completed)
	}
}

func TestGetSweepResults_Success(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sweeps/1/results", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepResults(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCancelSweep_Success(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/sweeps/1/cancel", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := CancelSweep(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["message"] != "Sweep job cancellation requested" {
		t.Errorf("unexpected message: %s", response["message"])
	}
}

func TestCreateSweep_InvalidJSON(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/sweeps", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler := CreateSweep(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetSweep_InvalidID(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request with invalid ID
	req := httptest.NewRequest("GET", "/api/v1/sweeps/invalid", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweep(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetSweep_MissingID(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request without ID
	req := httptest.NewRequest("GET", "/api/v1/sweeps/", nil)
	w := httptest.NewRecorder()

	// Add chi URL params (empty ID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweep(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetSweepBest_Success(t *testing.T) {
	mockSvc := &mockSweepService{
		findBestFunc: func(ctx context.Context, sweepID int64, metric string) (*services.BestConfiguration, error) {
			if metric != "quina_rate" {
				t.Errorf("expected metric 'quina_rate', got %s", metric)
			}
			return &services.BestConfiguration{
				SweepID:      sweepID,
				SimulationID: 1,
				Metrics: map[string]float64{
					"quina_rate": 0.05,
					"avg_hits":   2.3,
				},
				Rank:       1,
				Percentile: 100.0,
				VariationParams: map[string]interface{}{
					"alpha": 0.1,
					"beta":  0.2,
				},
			}, nil
		},
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sweeps/1/best?metric=quina_rate", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepBest(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response services.BestConfiguration
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.SweepID != 1 {
		t.Errorf("expected sweep ID 1, got %d", response.SweepID)
	}

	if response.SimulationID != 1 {
		t.Errorf("expected simulation ID 1, got %d", response.SimulationID)
	}

	if response.Rank != 1 {
		t.Errorf("expected rank 1, got %d", response.Rank)
	}

	if response.Percentile != 100.0 {
		t.Errorf("expected percentile 100.0, got %f", response.Percentile)
	}
}

func TestGetSweepBest_InvalidID(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request with invalid ID
	req := httptest.NewRequest("GET", "/api/v1/sweeps/invalid/best?metric=quina_rate", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepBest(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetSweepBest_MissingID(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request without ID
	req := httptest.NewRequest("GET", "/api/v1/sweeps//best?metric=quina_rate", nil)
	w := httptest.NewRecorder()

	// Add chi URL params (empty ID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepBest(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetSweepBest_MissingMetric(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request without metric
	req := httptest.NewRequest("GET", "/api/v1/sweeps/1/best", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepBest(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetSweepBest_ServiceError(t *testing.T) {
	mockSvc := &mockSweepService{
		findBestFunc: func(ctx context.Context, sweepID int64, metric string) (*services.BestConfiguration, error) {
			return nil, fmt.Errorf("sweep not completed")
		},
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sweeps/1/best?metric=quina_rate", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepBest(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetSweepVisualization_Success(t *testing.T) {
	mockSvc := &mockSweepService{
		getVisualizationFunc: func(ctx context.Context, sweepID int64, metrics []string) (*services.VisualizationData, error) {
			expectedMetrics := []string{"quina_rate", "avg_hits"}
			if len(metrics) != len(expectedMetrics) {
				t.Errorf("expected metrics %v, got %v", expectedMetrics, metrics)
			}
			return &services.VisualizationData{
				SweepID:    sweepID,
				Parameters: []string{"alpha", "beta"},
				Metrics:    metrics,
				DataPoints: []services.VisualizationDataPoint{
					{
						Params: map[string]interface{}{
							"alpha": 0.1,
							"beta":  0.2,
						},
						Metrics: map[string]float64{
							"quina_rate": 0.05,
							"avg_hits":   2.3,
						},
					},
				},
			}, nil
		},
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sweeps/1/visualization?metrics=quina_rate,avg_hits", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepVisualization(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response services.VisualizationData
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.SweepID != 1 {
		t.Errorf("expected sweep ID 1, got %d", response.SweepID)
	}

	if len(response.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(response.Parameters))
	}

	if len(response.DataPoints) != 1 {
		t.Errorf("expected 1 data point, got %d", len(response.DataPoints))
	}
}

func TestGetSweepVisualization_DefaultMetrics(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request without metrics
	req := httptest.NewRequest("GET", "/api/v1/sweeps/1/visualization", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepVisualization(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetSweepVisualization_InvalidID(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request with invalid ID
	req := httptest.NewRequest("GET", "/api/v1/sweeps/invalid/visualization", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepVisualization(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetSweepVisualization_MissingID(t *testing.T) {
	mockSvc := &mockSweepService{}

	// Create request without ID
	req := httptest.NewRequest("GET", "/api/v1/sweeps//visualization", nil)
	w := httptest.NewRecorder()

	// Add chi URL params (empty ID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepVisualization(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetSweepVisualization_ServiceError(t *testing.T) {
	mockSvc := &mockSweepService{
		getVisualizationFunc: func(ctx context.Context, sweepID int64, metrics []string) (*services.VisualizationData, error) {
			return nil, fmt.Errorf("sweep not found")
		},
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/sweeps/1/visualization", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetSweepVisualization(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
