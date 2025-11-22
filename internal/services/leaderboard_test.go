package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/garnizeh/luckyfive/internal/store/simulations"
)

// Mock implementation of simulations.Querier for testing
type mockSimulationQuerier struct {
	listSimulationsFunc func(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error)
}

func (m *mockSimulationQuerier) CancelSimulation(ctx context.Context, arg simulations.CancelSimulationParams) error {
	return nil
}

func (m *mockSimulationQuerier) ClaimPendingSimulation(ctx context.Context, arg simulations.ClaimPendingSimulationParams) (simulations.Simulation, error) {
	return simulations.Simulation{}, nil
}

func (m *mockSimulationQuerier) CompleteSimulation(ctx context.Context, arg simulations.CompleteSimulationParams) error {
	return nil
}

func (m *mockSimulationQuerier) CountSimulationsByStatus(ctx context.Context, status string) (int64, error) {
	return 0, nil
}

func (m *mockSimulationQuerier) CreateSimulation(ctx context.Context, arg simulations.CreateSimulationParams) (simulations.Simulation, error) {
	return simulations.Simulation{}, nil
}

func (m *mockSimulationQuerier) FailSimulation(ctx context.Context, arg simulations.FailSimulationParams) error {
	return nil
}

func (m *mockSimulationQuerier) GetContestResults(ctx context.Context, arg simulations.GetContestResultsParams) ([]simulations.SimulationContestResult, error) {
	return nil, nil
}

func (m *mockSimulationQuerier) GetContestResultsByMinHits(ctx context.Context, arg simulations.GetContestResultsByMinHitsParams) ([]simulations.SimulationContestResult, error) {
	return nil, nil
}

func (m *mockSimulationQuerier) GetSimulation(ctx context.Context, id int64) (simulations.Simulation, error) {
	return simulations.Simulation{}, nil
}

func (m *mockSimulationQuerier) InsertContestResult(ctx context.Context, arg simulations.InsertContestResultParams) error {
	return nil
}

func (m *mockSimulationQuerier) ListSimulations(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error) {
	if m.listSimulationsFunc != nil {
		return m.listSimulationsFunc(ctx, arg)
	}
	return []simulations.Simulation{}, nil
}

func (m *mockSimulationQuerier) ListSimulationsByStatus(ctx context.Context, arg simulations.ListSimulationsByStatusParams) ([]simulations.Simulation, error) {
	return []simulations.Simulation{}, nil
}

func (m *mockSimulationQuerier) UpdateSimulationStatus(ctx context.Context, arg simulations.UpdateSimulationStatusParams) error {
	return nil
}

func TestLeaderboardService_GetLeaderboard_Success(t *testing.T) {
	// Create mock summary
	summary := Summary{
		TotalContests: 100,
		QuinaHits:     5,
		QuadraHits:    20,
		TernoHits:     50,
		AverageHits:   2.5,
		HitRateQuina:  0.05,
		HitRateQuadra: 0.20,
		HitRateTerno:  0.50,
	}
	summaryJSON, _ := json.Marshal(summary)

	mockQuerier := &mockSimulationQuerier{
		listSimulationsFunc: func(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error) {
			return []simulations.Simulation{
				{
					ID:          1,
					Status:      "completed",
					Mode:        "simple",
					RecipeName:  sql.NullString{String: "Test Recipe 1", Valid: true},
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   time.Now().Format(time.RFC3339),
					CreatedBy:   sql.NullString{String: "user1", Valid: true},
				},
				{
					ID:          2,
					Status:      "completed",
					Mode:        "advanced",
					RecipeName:  sql.NullString{String: "Test Recipe 2", Valid: true},
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   time.Now().Format(time.RFC3339),
					CreatedBy:   sql.NullString{String: "user2", Valid: true},
				},
			}, nil
		},
	}

	service := NewLeaderboardService(mockQuerier, nil)

	req := LeaderboardRequest{
		Metric: "quina_rate",
		Mode:   "all",
		Limit:  10,
		Offset: 0,
	}

	entries, err := service.GetLeaderboard(context.Background(), req)
	if err != nil {
		t.Fatalf("GetLeaderboard failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	// Check ranking (should be sorted by metric value descending)
	if entries[0].Rank != 1 || entries[1].Rank != 2 {
		t.Errorf("incorrect ranking: entry 1 rank=%d, entry 2 rank=%d", entries[0].Rank, entries[1].Rank)
	}

	if entries[0].MetricValue != 0.05 {
		t.Errorf("expected metric value 0.05, got %f", entries[0].MetricValue)
	}
}

func TestLeaderboardService_GetLeaderboard_FilterByMode(t *testing.T) {
	summaryJSON, _ := json.Marshal(Summary{HitRateQuina: 0.05})

	mockQuerier := &mockSimulationQuerier{
		listSimulationsFunc: func(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error) {
			return []simulations.Simulation{
				{
					ID:          1,
					Status:      "completed",
					Mode:        "simple",
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   time.Now().Format(time.RFC3339),
				},
				{
					ID:          2,
					Status:      "completed",
					Mode:        "advanced",
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   time.Now().Format(time.RFC3339),
				},
			}, nil
		},
	}

	service := NewLeaderboardService(mockQuerier, nil)

	req := LeaderboardRequest{
		Metric: "quina_rate",
		Mode:   "simple",
		Limit:  10,
		Offset: 0,
	}

	entries, err := service.GetLeaderboard(context.Background(), req)
	if err != nil {
		t.Fatalf("GetLeaderboard failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry for mode filter, got %d", len(entries))
	}

	if entries[0].SimulationID != 1 {
		t.Errorf("expected simulation ID 1, got %d", entries[0].SimulationID)
	}
}

func TestLeaderboardService_GetLeaderboard_InvalidMetric(t *testing.T) {
	mockQuerier := &mockSimulationQuerier{}
	service := NewLeaderboardService(mockQuerier, nil)

	req := LeaderboardRequest{
		Metric: "invalid_metric",
		Mode:   "all",
		Limit:  10,
		Offset: 0,
	}

	_, err := service.GetLeaderboard(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid metric")
	}

	if err.Error() != "invalid metric: invalid_metric" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLeaderboardService_GetLeaderboard_Pagination(t *testing.T) {
	summaryJSON, _ := json.Marshal(Summary{HitRateQuina: 0.05})

	mockQuerier := &mockSimulationQuerier{
		listSimulationsFunc: func(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error) {
			return []simulations.Simulation{
				{
					ID:          1,
					Status:      "completed",
					Mode:        "simple",
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   time.Now().Format(time.RFC3339),
				},
				{
					ID:          2,
					Status:      "completed",
					Mode:        "simple",
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   time.Now().Format(time.RFC3339),
				},
				{
					ID:          3,
					Status:      "completed",
					Mode:        "simple",
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   time.Now().Format(time.RFC3339),
				},
			}, nil
		},
	}

	service := NewLeaderboardService(mockQuerier, nil)

	// Test limit
	req := LeaderboardRequest{
		Metric: "quina_rate",
		Mode:   "all",
		Limit:  2,
		Offset: 0,
	}

	entries, err := service.GetLeaderboard(context.Background(), req)
	if err != nil {
		t.Fatalf("GetLeaderboard failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries with limit, got %d", len(entries))
	}

	// Test offset
	req.Offset = 1
	entries, err = service.GetLeaderboard(context.Background(), req)
	if err != nil {
		t.Fatalf("GetLeaderboard failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries with offset, got %d", len(entries))
	}

	if entries[0].Rank != 2 {
		t.Errorf("expected first entry rank 2 with offset 1, got %d", entries[0].Rank)
	}
}

func TestLeaderboardService_GetLeaderboard_Defaults(t *testing.T) {
	summaryJSON, _ := json.Marshal(Summary{HitRateQuina: 0.05})

	mockQuerier := &mockSimulationQuerier{
		listSimulationsFunc: func(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error) {
			return []simulations.Simulation{
				{
					ID:          1,
					Status:      "completed",
					Mode:        "simple",
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   time.Now().Format(time.RFC3339),
				},
			}, nil
		},
	}

	service := NewLeaderboardService(mockQuerier, nil)

	// Test with empty request (should use defaults)
	req := LeaderboardRequest{}

	entries, err := service.GetLeaderboard(context.Background(), req)
	if err != nil {
		t.Fatalf("GetLeaderboard failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry with defaults, got %d", len(entries))
	}

	// Should use default metric "quina_rate"
	if entries[0].MetricValue != 0.05 {
		t.Errorf("expected default metric quina_rate value 0.05, got %f", entries[0].MetricValue)
	}
}

func TestLeaderboardService_calculateMetricValue(t *testing.T) {
	service := NewLeaderboardService(nil, nil)

	summary := Summary{
		TotalContests: 100,
		QuinaHits:     5,
		QuadraHits:    20,
		TernoHits:     50,
		AverageHits:   2.5,
		HitRateQuina:  0.05,
		HitRateQuadra: 0.20,
		HitRateTerno:  0.50,
	}

	tests := []struct {
		metric   string
		expected float64
	}{
		{"quina_rate", 0.05},
		{"quadra_rate", 0.20},
		{"terno_rate", 0.50},
		{"avg_hits", 2.5},
		{"total_quinaz", 5.0},
		{"total_quadras", 20.0},
		{"total_ternos", 50.0},
		{"hit_efficiency", 2.5},
		{"invalid", 0.0},
	}

	for _, test := range tests {
		value := service.calculateMetricValue(test.metric, summary)
		if value != test.expected {
			t.Errorf("metric %s: expected %f, got %f", test.metric, test.expected, value)
		}
	}
}

func TestLeaderboardService_isValidMetric(t *testing.T) {
	service := NewLeaderboardService(nil, nil)

	validMetrics := []string{
		"quina_rate", "quadra_rate", "terno_rate", "avg_hits",
		"total_quinaz", "total_quadras", "total_ternos", "hit_efficiency",
	}

	for _, metric := range validMetrics {
		if !service.isValidMetric(metric) {
			t.Errorf("expected metric %s to be valid", metric)
		}
	}

	if service.isValidMetric("invalid_metric") {
		t.Error("expected invalid_metric to be invalid")
	}
}

func TestLeaderboardService_fetchSimulations_ListSimulationsError(t *testing.T) {
	mockQuerier := &mockSimulationQuerier{
		listSimulationsFunc: func(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error) {
			return nil, fmt.Errorf("database connection failed")
		},
	}

	service := NewLeaderboardService(mockQuerier, nil)

	req := LeaderboardRequest{
		Metric: "quina_rate",
		Mode:   "all",
	}

	_, err := service.fetchSimulations(context.Background(), req)
	if err == nil {
		t.Error("expected error from ListSimulations, got nil")
	}
	if !strings.Contains(err.Error(), "database connection failed") {
		t.Errorf("expected database error, got %q", err.Error())
	}
}

func TestLeaderboardService_fetchSimulations_InvalidDateFrom(t *testing.T) {
	mockQuerier := &mockSimulationQuerier{
		listSimulationsFunc: func(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error) {
			return []simulations.Simulation{
				{
					ID:        1,
					Status:    "completed",
					Mode:      "simple",
					CreatedAt: "2023-01-01T00:00:00Z",
				},
			}, nil
		},
	}

	service := NewLeaderboardService(mockQuerier, nil)

	req := LeaderboardRequest{
		Metric:   "quina_rate",
		Mode:     "all",
		DateFrom: "invalid-date",
	}

	_, err := service.fetchSimulations(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid date_from, got nil")
	}
	if !strings.Contains(err.Error(), "invalid date_from format") {
		t.Errorf("expected date_from error, got %q", err.Error())
	}
}

func TestLeaderboardService_fetchSimulations_InvalidDateTo(t *testing.T) {
	mockQuerier := &mockSimulationQuerier{
		listSimulationsFunc: func(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error) {
			return []simulations.Simulation{
				{
					ID:        1,
					Status:    "completed",
					Mode:      "simple",
					CreatedAt: "2023-01-01T00:00:00Z",
				},
			}, nil
		},
	}

	service := NewLeaderboardService(mockQuerier, nil)

	req := LeaderboardRequest{
		Metric: "quina_rate",
		Mode:   "all",
		DateTo: "invalid-date",
	}

	_, err := service.fetchSimulations(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid date_to, got nil")
	}
	if !strings.Contains(err.Error(), "invalid date_to format") {
		t.Errorf("expected date_to error, got %q", err.Error())
	}
}

func TestLeaderboardService_fetchSimulations_DateFiltering(t *testing.T) {
	summaryJSON, _ := json.Marshal(Summary{HitRateQuina: 0.05})

	mockQuerier := &mockSimulationQuerier{
		listSimulationsFunc: func(ctx context.Context, arg simulations.ListSimulationsParams) ([]simulations.Simulation, error) {
			return []simulations.Simulation{
				{
					ID:          1,
					Status:      "completed",
					Mode:        "simple",
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   "2023-01-01T00:00:00Z",
				},
				{
					ID:          2,
					Status:      "completed",
					Mode:        "simple",
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   "2023-01-15T00:00:00Z",
				},
				{
					ID:          3,
					Status:      "completed",
					Mode:        "simple",
					SummaryJson: sql.NullString{String: string(summaryJSON), Valid: true},
					CreatedAt:   "2023-01-31T00:00:00Z",
				},
			}, nil
		},
	}

	service := NewLeaderboardService(mockQuerier, nil)

	// Test date range filtering
	req := LeaderboardRequest{
		Metric:   "quina_rate",
		Mode:     "all",
		DateFrom: "2023-01-10T00:00:00Z",
		DateTo:   "2023-01-20T00:00:00Z",
	}

	sims, err := service.fetchSimulations(context.Background(), req)
	if err != nil {
		t.Fatalf("fetchSimulations failed: %v", err)
	}

	// Should only return simulation 2 (ID=2, created 2023-01-15)
	if len(sims) != 1 {
		t.Errorf("expected 1 simulation in date range, got %d", len(sims))
	}
	if len(sims) > 0 && sims[0].ID != 2 {
		t.Errorf("expected simulation ID 2, got %d", sims[0].ID)
	}
}
