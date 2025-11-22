package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/garnizeh/luckyfive/internal/store/comparisons"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
)

// Helper function to create a test logger
func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestComparisonService_calculateStats(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	values := map[int64]float64{
		1: 1.0,
		2: 2.0,
		3: 3.0,
		4: 4.0,
		5: 5.0,
	}

	stats := service.calculateStats(values)

	if stats.Count != 5 {
		t.Errorf("expected count 5, got %d", stats.Count)
	}
	if stats.Mean != 3.0 {
		t.Errorf("expected mean 3.0, got %f", stats.Mean)
	}
	if stats.Median != 3.0 {
		t.Errorf("expected median 3.0, got %f", stats.Median)
	}
	if stats.Min != 1.0 {
		t.Errorf("expected min 1.0, got %f", stats.Min)
	}
	if stats.Max != 5.0 {
		t.Errorf("expected max 5.0, got %f", stats.Max)
	}
	if stats.StdDev <= 0 {
		t.Errorf("expected stddev > 0, got %f", stats.StdDev)
	}
}

func TestComparisonService_calculateStats_Empty(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	values := map[int64]float64{}

	stats := service.calculateStats(values)

	if stats.Count != 0 {
		t.Errorf("expected count 0, got %d", stats.Count)
	}
}

func TestComparisonService_calculateStats_SingleValue(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	values := map[int64]float64{
		1: 42.0,
	}

	stats := service.calculateStats(values)

	if stats.Count != 1 {
		t.Errorf("expected count 1, got %d", stats.Count)
	}
	if stats.Mean != 42.0 {
		t.Errorf("expected mean 42.0, got %f", stats.Mean)
	}
	if stats.Median != 42.0 {
		t.Errorf("expected median 42.0, got %f", stats.Median)
	}
	if stats.Min != 42.0 {
		t.Errorf("expected min 42.0, got %f", stats.Min)
	}
	if stats.Max != 42.0 {
		t.Errorf("expected max 42.0, got %f", stats.Max)
	}
	if stats.StdDev != 0.0 {
		t.Errorf("expected stddev 0.0, got %f", stats.StdDev)
	}
}

func TestComparisonService_isValidMetric(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	validMetrics := []string{
		"quina_rate",
		"quadra_rate",
		"terno_rate",
		"avg_hits",
		"total_quinaz",
		"total_quadras",
		"total_ternos",
		"hit_efficiency",
	}

	for _, metric := range validMetrics {
		if !service.isValidMetric(metric) {
			t.Errorf("metric %s should be valid", metric)
		}
	}

	invalidMetrics := []string{
		"invalid",
		"wrong_metric",
		"",
		"some_other_metric",
	}

	for _, metric := range invalidMetrics {
		if service.isValidMetric(metric) {
			t.Errorf("metric %s should be invalid", metric)
		}
	}
}

func TestCompareRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CompareRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{1, 2},
				Metrics:       []string{"quina_rate"},
			},
			wantErr: false,
		},
		{
			name: "too few simulations",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{1},
				Metrics:       []string{"quina_rate"},
			},
			wantErr: true,
		},
		{
			name: "no simulations",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{},
				Metrics:       []string{"quina_rate"},
			},
			wantErr: true,
		},
		{
			name: "empty metrics defaults to valid",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{1, 2},
				Metrics:       []string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &sql.DB{}
			logger := createTestLogger()
			service := NewComparisonService(nil, nil, db, logger)

			err := service.validateCompareRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCompareRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Mock implementations for testing
type mockSimulationQueries struct {
	simulations        map[int64]*simulations.Simulation
	getSimulationError error
}

func (m *mockSimulationQueries) GetSimulation(ctx context.Context, id int64) (simulations.Simulation, error) {
	if m.getSimulationError != nil {
		return simulations.Simulation{}, m.getSimulationError
	}
	if sim, exists := m.simulations[id]; exists {
		return *sim, nil
	}
	return simulations.Simulation{}, errors.New("simulation not found")
}

func (m *mockSimulationQueries) CancelSimulation(ctx context.Context, arg simulations.CancelSimulationParams) error {
	return nil
}

func (m *mockSimulationQueries) ClaimPendingSimulation(ctx context.Context, arg simulations.ClaimPendingSimulationParams) (simulations.Simulation, error) {
	return simulations.Simulation{}, nil
}

func (m *mockSimulationQueries) CompleteSimulation(ctx context.Context, arg simulations.CompleteSimulationParams) error {
	return nil
}

func (m *mockSimulationQueries) CountSimulationsByStatus(ctx context.Context, status string) (int64, error) {
	return 0, nil
}

func (m *mockSimulationQueries) CreateSimulation(ctx context.Context, arg simulations.CreateSimulationParams) (simulations.Simulation, error) {
	return simulations.Simulation{}, nil
}

func (m *mockSimulationQueries) FailSimulation(ctx context.Context, arg simulations.FailSimulationParams) error {
	return nil
}

func (m *mockSimulationQueries) GetContestResults(ctx context.Context, arg simulations.GetContestResultsParams) ([]simulations.SimulationContestResult, error) {
	return nil, nil
}

func (m *mockSimulationQueries) GetContestResultsByMinHits(ctx context.Context, arg simulations.GetContestResultsByMinHitsParams) ([]simulations.SimulationContestResult, error) {
	return nil, nil
}

func (m *mockSimulationQueries) InsertContestResult(ctx context.Context, arg simulations.InsertContestResultParams) error {
	return nil
}

func (m *mockSimulationQueries) ListSimulations(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error) {
	return nil, nil
}

func (m *mockSimulationQueries) ListSimulationsByStatus(ctx context.Context, arg simulations.ListSimulationsByStatusParams) ([]simulations.Simulation, error) {
	return nil, nil
}

func (m *mockSimulationQueries) UpdateSimulationStatus(ctx context.Context, arg simulations.UpdateSimulationStatusParams) error {
	return nil
}

type mockComparisonQueries struct {
	comparisons           map[int64]*comparisons.Comparison
	nextID                int64
	createComparisonError error
	getComparisonError    error
	insertMetricError     error
	updateResultError     error
	listComparisonsResult []comparisons.Comparison
	listComparisonsError  error
}

func (m *mockComparisonQueries) CreateComparison(ctx context.Context, arg comparisons.CreateComparisonParams) (comparisons.Comparison, error) {
	if m.createComparisonError != nil {
		return comparisons.Comparison{}, m.createComparisonError
	}
	m.nextID++
	comp := comparisons.Comparison{
		ID:            m.nextID,
		Name:          arg.Name,
		Description:   arg.Description,
		SimulationIds: arg.SimulationIds,
		Metric:        arg.Metric,
		CreatedAt:     sql.NullString{String: "2023-01-01T00:00:00Z", Valid: true},
	}
	m.comparisons[comp.ID] = &comp
	return comp, nil
}

func (m *mockComparisonQueries) GetComparison(ctx context.Context, id int64) (comparisons.Comparison, error) {
	if m.getComparisonError != nil {
		return comparisons.Comparison{}, m.getComparisonError
	}
	if comp, exists := m.comparisons[id]; exists {
		return *comp, nil
	}
	return comparisons.Comparison{}, errors.New("comparison not found")
}

func (m *mockComparisonQueries) InsertComparisonMetric(ctx context.Context, arg comparisons.InsertComparisonMetricParams) error {
	return m.insertMetricError
}

func (m *mockComparisonQueries) UpdateComparisonResult(ctx context.Context, arg comparisons.UpdateComparisonResultParams) error {
	return m.updateResultError
}

func (m *mockComparisonQueries) ListComparisons(ctx context.Context, arg comparisons.ListComparisonsParams) ([]comparisons.Comparison, error) {
	if m.listComparisonsError != nil {
		return nil, m.listComparisonsError
	}
	return m.listComparisonsResult, nil
}

func (m *mockComparisonQueries) DeleteComparison(ctx context.Context, id int64) error {
	return nil
}

func (m *mockComparisonQueries) GetComparisonMetrics(ctx context.Context, comparisonID int64) ([]comparisons.ComparisonMetric, error) {
	return nil, nil
}

func (m *mockComparisonQueries) GetComparisonMetricsBySimulation(ctx context.Context, simulationID int64) ([]comparisons.ComparisonMetric, error) {
	return nil, nil
}

func createMockSimulations() map[int64]*simulations.Simulation {
	summary1 := Summary{
		TotalContests: 100,
		QuinaHits:     5,
		QuadraHits:    10,
		TernoHits:     20,
		AverageHits:   2.5,
		HitRateQuina:  0.05,
		HitRateQuadra: 0.10,
		HitRateTerno:  0.20,
	}
	summary1JSON, _ := json.Marshal(summary1)

	summary2 := Summary{
		TotalContests: 100,
		QuinaHits:     8,
		QuadraHits:    15,
		TernoHits:     25,
		AverageHits:   3.2,
		HitRateQuina:  0.08,
		HitRateQuadra: 0.15,
		HitRateTerno:  0.25,
	}
	summary2JSON, _ := json.Marshal(summary2)

	return map[int64]*simulations.Simulation{
		1: {
			ID:          1,
			Status:      "completed",
			RecipeName:  sql.NullString{String: "Recipe A", Valid: true},
			SummaryJson: sql.NullString{String: string(summary1JSON), Valid: true},
		},
		2: {
			ID:          2,
			Status:      "completed",
			RecipeName:  sql.NullString{String: "Recipe B", Valid: true},
			SummaryJson: sql.NullString{String: string(summary2JSON), Valid: true},
		},
		3: {
			ID:          3,
			Status:      "failed",
			RecipeName:  sql.NullString{String: "Recipe C", Valid: true},
			SummaryJson: sql.NullString{String: "", Valid: false},
		},
	}
}

func TestComparisonService_Compare_Success(t *testing.T) {
	mockSimQueries := &mockSimulationQueries{
		simulations: createMockSimulations(),
	}
	mockCompQueries := &mockComparisonQueries{
		comparisons: make(map[int64]*comparisons.Comparison),
	}

	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(mockCompQueries, mockSimQueries, db, logger)

	req := CompareRequest{
		Name:          "Test Comparison",
		Description:   "Test description",
		SimulationIDs: []int64{1, 2},
		Metrics:       []string{"quina_rate", "avg_hits"},
	}

	result, err := service.Compare(context.Background(), req)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	if result.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, result.Name)
	}
	if result.Description != req.Description {
		t.Errorf("expected description %s, got %s", req.Description, result.Description)
	}
	if len(result.SimulationIDs) != 2 {
		t.Errorf("expected 2 simulation IDs, got %d", len(result.SimulationIDs))
	}
	if len(result.Metrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(result.Metrics))
	}
	if len(result.Rankings) != 2 {
		t.Errorf("expected 2 ranking entries, got %d", len(result.Rankings))
	}
	if len(result.Statistics) != 2 {
		t.Errorf("expected 2 statistics entries, got %d", len(result.Statistics))
	}
	if len(result.WinnerByMetric) != 2 {
		t.Errorf("expected 2 winners, got %d", len(result.WinnerByMetric))
	}
}

func TestComparisonService_Compare_ValidationError(t *testing.T) {
	mockSimQueries := &mockSimulationQueries{}
	mockCompQueries := &mockComparisonQueries{}

	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(mockCompQueries, mockSimQueries, db, logger)

	tests := []struct {
		name    string
		req     CompareRequest
		wantErr string
	}{
		{
			name: "too few simulations",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{1},
				Metrics:       []string{"quina_rate"},
			},
			wantErr: "need at least 2 simulations",
		},
		{
			name: "invalid metric",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{1, 2},
				Metrics:       []string{"invalid_metric"},
			},
			wantErr: "invalid metric: invalid_metric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Compare(context.Background(), tt.req)
			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErr)
			} else if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestComparisonService_Compare_SimulationNotFound(t *testing.T) {
	mockSimQueries := &mockSimulationQueries{
		simulations: make(map[int64]*simulations.Simulation),
	}
	mockCompQueries := &mockComparisonQueries{}

	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(mockCompQueries, mockSimQueries, db, logger)

	req := CompareRequest{
		Name:          "Test",
		SimulationIDs: []int64{1, 2},
		Metrics:       []string{"quina_rate"},
	}

	_, err := service.Compare(context.Background(), req)
	if err == nil {
		t.Error("expected error for simulation not found, got nil")
	}
	if !strings.Contains(err.Error(), "get simulation") {
		t.Errorf("expected error about getting simulation, got %q", err.Error())
	}
}

func TestComparisonService_Compare_SimulationNotCompleted(t *testing.T) {
	mockSimQueries := &mockSimulationQueries{
		simulations: createMockSimulations(),
	}
	mockCompQueries := &mockComparisonQueries{}

	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(mockCompQueries, mockSimQueries, db, logger)

	req := CompareRequest{
		Name:          "Test",
		SimulationIDs: []int64{1, 3}, // 3 is failed
		Metrics:       []string{"quina_rate"},
	}

	_, err := service.Compare(context.Background(), req)
	if err == nil {
		t.Error("expected error for incomplete simulation, got nil")
	}
	if !strings.Contains(err.Error(), "not completed") {
		t.Errorf("expected error about simulation not completed, got %q", err.Error())
	}
}

func TestComparisonService_Compare_DefaultMetrics(t *testing.T) {
	mockSimQueries := &mockSimulationQueries{
		simulations: createMockSimulations(),
	}
	mockCompQueries := &mockComparisonQueries{
		comparisons: make(map[int64]*comparisons.Comparison),
	}

	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(mockCompQueries, mockSimQueries, db, logger)

	req := CompareRequest{
		Name:          "Test",
		SimulationIDs: []int64{1, 2},
		Metrics:       []string{}, // empty metrics should default
	}

	result, err := service.Compare(context.Background(), req)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	expectedMetrics := []string{"quina_rate", "avg_hits"}
	if len(result.Metrics) != len(expectedMetrics) {
		t.Errorf("expected %d default metrics, got %d", len(expectedMetrics), len(result.Metrics))
	}
	for i, expected := range expectedMetrics {
		if result.Metrics[i] != expected {
			t.Errorf("expected metric %s at position %d, got %s", expected, i, result.Metrics[i])
		}
	}
}

func TestComparisonService_GetComparison_Success(t *testing.T) {
	resultJSON := `{"id":1,"name":"Test","simulation_ids":[1,2],"metrics":["quina_rate"],"rankings":{"quina_rate":[]},"statistics":{"quina_rate":{"mean":0,"median":0,"std_dev":0,"min":0,"max":0,"count":0}},"winner_by_metric":{"quina_rate":1},"created_at":"2023-01-01T00:00:00Z"}`
	mockCompQueries := &mockComparisonQueries{
		comparisons: map[int64]*comparisons.Comparison{
			1: {
				ID:         1,
				Name:       "Test",
				ResultJson: sql.NullString{String: resultJSON, Valid: true},
			},
		},
	}

	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(mockCompQueries, nil, db, logger)

	result, err := service.GetComparison(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetComparison() error = %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.Name != "Test" {
		t.Errorf("expected name 'Test', got %s", result.Name)
	}
}

func TestComparisonService_GetComparison_NotFound(t *testing.T) {
	mockCompQueries := &mockComparisonQueries{
		comparisons: make(map[int64]*comparisons.Comparison),
	}

	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(mockCompQueries, nil, db, logger)

	_, err := service.GetComparison(context.Background(), 999)
	if err == nil {
		t.Error("expected error for comparison not found, got nil")
	}
}

func TestComparisonService_GetComparison_NoResults(t *testing.T) {
	mockCompQueries := &mockComparisonQueries{
		comparisons: map[int64]*comparisons.Comparison{
			1: {
				ID:         1,
				Name:       "Test",
				ResultJson: sql.NullString{String: "", Valid: false},
			},
		},
	}

	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(mockCompQueries, nil, db, logger)

	_, err := service.GetComparison(context.Background(), 1)
	if err == nil {
		t.Error("expected error for no results, got nil")
	}
	if !strings.Contains(err.Error(), "no results yet") {
		t.Errorf("expected error about no results, got %q", err.Error())
	}
}

func TestComparisonService_ListComparisons_Success(t *testing.T) {
	expectedComparisons := []comparisons.Comparison{
		{ID: 1, Name: "Comp1"},
		{ID: 2, Name: "Comp2"},
	}
	mockCompQueries := &mockComparisonQueries{
		listComparisonsResult: expectedComparisons,
	}

	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(mockCompQueries, nil, db, logger)

	result, err := service.ListComparisons(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("ListComparisons() error = %v", err)
	}

	if len(result) != len(expectedComparisons) {
		t.Errorf("expected %d comparisons, got %d", len(expectedComparisons), len(result))
	}
}

func TestComparisonService_calculateMetric(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	simulations := createMockSimulations()

	ranks, stats := service.calculateMetric("quina_rate", simulations)

	// Should have 2 simulations (IDs 1 and 2, since 3 is failed)
	if len(ranks) != 2 {
		t.Errorf("expected 2 ranks, got %d", len(ranks))
	}
	if stats.Count != 2 {
		t.Errorf("expected stats count 2, got %d", stats.Count)
	}

	// Check that ranks are ordered by quina_rate (simulation 2 should be first)
	if len(ranks) >= 2 {
		if ranks[0].SimulationID != 2 {
			t.Errorf("expected simulation 2 first (higher quina_rate), got %d", ranks[0].SimulationID)
		}
		if ranks[1].SimulationID != 1 {
			t.Errorf("expected simulation 1 second, got %d", ranks[1].SimulationID)
		}
	}
}

func TestComparisonService_calculateMetric_AllMetrics(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	simulations := createMockSimulations()

	tests := []struct {
		metric       string
		expectedSim1 float64 // expected value for simulation 1
		expectedSim2 float64 // expected value for simulation 2
		firstWinner  int64   // which simulation should be first (higher value wins)
	}{
		{
			metric:       "quina_rate",
			expectedSim1: 0.05,
			expectedSim2: 0.08,
			firstWinner:  2,
		},
		{
			metric:       "quadra_rate",
			expectedSim1: 0.10,
			expectedSim2: 0.15,
			firstWinner:  2,
		},
		{
			metric:       "terno_rate",
			expectedSim1: 0.20,
			expectedSim2: 0.25,
			firstWinner:  2,
		},
		{
			metric:       "avg_hits",
			expectedSim1: 2.5,
			expectedSim2: 3.2,
			firstWinner:  2,
		},
		{
			metric:       "total_quinaz",
			expectedSim1: 5.0,
			expectedSim2: 8.0,
			firstWinner:  2,
		},
		{
			metric:       "total_quadras",
			expectedSim1: 10.0,
			expectedSim2: 15.0,
			firstWinner:  2,
		},
		{
			metric:       "total_ternos",
			expectedSim1: 20.0,
			expectedSim2: 25.0,
			firstWinner:  2,
		},
		{
			metric:       "hit_efficiency",
			expectedSim1: 2.5, // same as avg_hits since TotalContests > 0
			expectedSim2: 3.2,
			firstWinner:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.metric, func(t *testing.T) {
			ranks, stats := service.calculateMetric(tt.metric, simulations)

			if len(ranks) != 2 {
				t.Errorf("expected 2 ranks, got %d", len(ranks))
			}
			if stats.Count != 2 {
				t.Errorf("expected stats count 2, got %d", stats.Count)
			}

			// Check values are correctly extracted
			var sim1Value, sim2Value float64
			for _, rank := range ranks {
				if rank.SimulationID == 1 {
					sim1Value = rank.Value
				} else if rank.SimulationID == 2 {
					sim2Value = rank.Value
				}
			}

			if sim1Value != tt.expectedSim1 {
				t.Errorf("expected simulation 1 value %f, got %f", tt.expectedSim1, sim1Value)
			}
			if sim2Value != tt.expectedSim2 {
				t.Errorf("expected simulation 2 value %f, got %f", tt.expectedSim2, sim2Value)
			}

			// Check ranking (first should be the winner)
			if len(ranks) > 0 && ranks[0].SimulationID != tt.firstWinner {
				t.Errorf("expected simulation %d first, got %d", tt.firstWinner, ranks[0].SimulationID)
			}
		})
	}
}

func TestComparisonService_calculateMetric_InvalidMetric(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	simulations := createMockSimulations()

	ranks, stats := service.calculateMetric("invalid_metric", simulations)

	// Should have 2 simulations with value 0 (default case)
	if len(ranks) != 2 {
		t.Errorf("expected 2 ranks, got %d", len(ranks))
	}
	if stats.Count != 2 {
		t.Errorf("expected stats count 2, got %d", stats.Count)
	}

	// All values should be 0
	for _, rank := range ranks {
		if rank.Value != 0.0 {
			t.Errorf("expected value 0 for invalid metric, got %f", rank.Value)
		}
	}
}

func TestComparisonService_calculateMetric_NoSummaryData(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	// Create simulation with invalid summary JSON
	simulations := map[int64]*simulations.Simulation{
		1: {
			ID:          1,
			Status:      "completed",
			RecipeName:  sql.NullString{String: "Recipe A", Valid: true},
			SummaryJson: sql.NullString{String: "invalid json", Valid: true}, // Invalid JSON
		},
	}

	ranks, stats := service.calculateMetric("quina_rate", simulations)

	// Should have 0 ranks (no valid data)
	if len(ranks) != 0 {
		t.Errorf("expected 0 ranks for invalid JSON, got %d", len(ranks))
	}
	if stats.Count != 0 {
		t.Errorf("expected stats count 0, got %d", stats.Count)
	}
}

func TestComparisonService_calculateMetric_EmptySimulations(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	simulations := make(map[int64]*simulations.Simulation)

	ranks, stats := service.calculateMetric("quina_rate", simulations)

	if len(ranks) != 0 {
		t.Errorf("expected 0 ranks for empty simulations, got %d", len(ranks))
	}
	if stats.Count != 0 {
		t.Errorf("expected stats count 0, got %d", stats.Count)
	}
}

func TestComparisonService_calculateStats_ZeroValues(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	values := map[int64]float64{
		1: 0.0,
		2: 0.0,
		3: 0.0,
	}

	stats := service.calculateStats(values)

	if stats.Count != 3 {
		t.Errorf("expected count 3, got %d", stats.Count)
	}
	if stats.Mean != 0.0 {
		t.Errorf("expected mean 0.0, got %f", stats.Mean)
	}
	if stats.StdDev != 0.0 {
		t.Errorf("expected stddev 0.0, got %f", stats.StdDev)
	}
}

func TestComparisonService_calculateStats_NegativeValues(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	values := map[int64]float64{
		1: -1.0,
		2: -2.0,
		3: -3.0,
	}

	stats := service.calculateStats(values)

	if stats.Count != 3 {
		t.Errorf("expected count 3, got %d", stats.Count)
	}
	if stats.Min != -3.0 {
		t.Errorf("expected min -3.0, got %f", stats.Min)
	}
	if stats.Max != -1.0 {
		t.Errorf("expected max -1.0, got %f", stats.Max)
	}
}

func TestComparisonService_calculateStats_LargeNumbers(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	values := map[int64]float64{
		1: 1000000.0,
		2: 2000000.0,
		3: 3000000.0,
	}

	stats := service.calculateStats(values)

	if stats.Count != 3 {
		t.Errorf("expected count 3, got %d", stats.Count)
	}
	if stats.Mean != 2000000.0 {
		t.Errorf("expected mean 2000000.0, got %f", stats.Mean)
	}
	if stats.Min != 1000000.0 {
		t.Errorf("expected min 1000000.0, got %f", stats.Min)
	}
	if stats.Max != 3000000.0 {
		t.Errorf("expected max 3000000.0, got %f", stats.Max)
	}
}

func TestComparisonService_validateCompareRequest_InvalidMetrics(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	tests := []struct {
		name    string
		req     CompareRequest
		wantErr string
	}{
		{
			name: "invalid metric",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{1, 2},
				Metrics:       []string{"invalid_metric"},
			},
			wantErr: "invalid metric: invalid_metric",
		},
		{
			name: "multiple invalid metrics",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{1, 2},
				Metrics:       []string{"quina_rate", "invalid1", "invalid2"},
			},
			wantErr: "invalid metric: invalid1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateCompareRequest(tt.req)
			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErr)
			} else if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestComparisonService_validateCompareRequest_EdgeCases(t *testing.T) {
	db := &sql.DB{}
	logger := createTestLogger()
	service := NewComparisonService(nil, nil, db, logger)

	tests := []struct {
		name    string
		req     CompareRequest
		wantErr bool
	}{
		{
			name: "exactly 2 simulations",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{1, 2},
				Metrics:       []string{"quina_rate"},
			},
			wantErr: false,
		},
		{
			name: "many simulations",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{1, 2, 3, 4, 5},
				Metrics:       []string{"quina_rate"},
			},
			wantErr: false,
		},
		{
			name: "all valid metrics",
			req: CompareRequest{
				Name:          "Test",
				SimulationIDs: []int64{1, 2},
				Metrics:       []string{"quina_rate", "quadra_rate", "terno_rate", "avg_hits", "total_quinaz", "total_quadras", "total_ternos", "hit_efficiency"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateCompareRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCompareRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
