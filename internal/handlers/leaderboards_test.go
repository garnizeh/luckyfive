package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/garnizeh/luckyfive/internal/services"
)

// Mock implementation of LeaderboardServicer for testing
type mockLeaderboardService struct {
	getLeaderboardFunc func(ctx context.Context, req services.LeaderboardRequest) ([]services.LeaderboardEntry, error)
}

func (m *mockLeaderboardService) GetLeaderboard(ctx context.Context, req services.LeaderboardRequest) ([]services.LeaderboardEntry, error) {
	if m.getLeaderboardFunc != nil {
		return m.getLeaderboardFunc(ctx, req)
	}
	return []services.LeaderboardEntry{
		{
			Rank:           1,
			SimulationID:   1,
			SimulationName: "Simulation 1",
			RecipeName:     "test_recipe",
			MetricValue:    0.85,
			CreatedAt:      "2025-01-01T00:00:00Z",
			CreatedBy:      "test_user",
		},
		{
			Rank:           2,
			SimulationID:   2,
			SimulationName: "Simulation 2",
			RecipeName:     "test_recipe_2",
			MetricValue:    0.75,
			CreatedAt:      "2025-01-02T00:00:00Z",
			CreatedBy:      "test_user",
		},
	}, nil
}

func TestGetLeaderboard_Success(t *testing.T) {
	mockSvc := &mockLeaderboardService{
		getLeaderboardFunc: func(ctx context.Context, req services.LeaderboardRequest) ([]services.LeaderboardEntry, error) {
			// Verify request parameters
			if req.Metric != "quina_rate" {
				t.Errorf("expected metric 'quina_rate', got %s", req.Metric)
			}
			if req.Mode != "all" {
				t.Errorf("expected mode 'all', got %s", req.Mode)
			}
			if req.Limit != 50 {
				t.Errorf("expected limit 50, got %d", req.Limit)
			}
			if req.Offset != 0 {
				t.Errorf("expected offset 0, got %d", req.Offset)
			}
			return []services.LeaderboardEntry{
				{
					Rank:           1,
					SimulationID:   1,
					SimulationName: "Simulation 1",
					RecipeName:     "test_recipe",
					MetricValue:    0.85,
					CreatedAt:      "2025-01-01T00:00:00Z",
					CreatedBy:      "test_user",
				},
			}, nil
		},
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/leaderboards/quina_rate", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metric", "quina_rate")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetLeaderboard(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	leaderboard, ok := response["leaderboard"].([]any)
	if !ok {
		t.Fatal("expected leaderboard array in response")
	}

	if len(leaderboard) != 1 {
		t.Errorf("expected 1 leaderboard entry, got %d", len(leaderboard))
	}

	if response["total"].(float64) != 1 {
		t.Errorf("expected total 1, got %v", response["total"])
	}

	if response["limit"].(float64) != 50 {
		t.Errorf("expected limit 50, got %v", response["limit"])
	}

	if response["offset"].(float64) != 0 {
		t.Errorf("expected offset 0, got %v", response["offset"])
	}
}

func TestGetLeaderboard_WithFilters(t *testing.T) {
	mockSvc := &mockLeaderboardService{
		getLeaderboardFunc: func(ctx context.Context, req services.LeaderboardRequest) ([]services.LeaderboardEntry, error) {
			// Verify request parameters
			if req.Metric != "avg_hits" {
				t.Errorf("expected metric 'avg_hits', got %s", req.Metric)
			}
			if req.Mode != "simple" {
				t.Errorf("expected mode 'simple', got %s", req.Mode)
			}
			if req.DateFrom != "2025-01-01T00:00:00Z" {
				t.Errorf("expected date_from '2025-01-01T00:00:00Z', got %s", req.DateFrom)
			}
			if req.DateTo != "2025-12-31T23:59:59Z" {
				t.Errorf("expected date_to '2025-12-31T23:59:59Z', got %s", req.DateTo)
			}
			if req.Limit != 10 {
				t.Errorf("expected limit 10, got %d", req.Limit)
			}
			if req.Offset != 5 {
				t.Errorf("expected offset 5, got %d", req.Offset)
			}
			return []services.LeaderboardEntry{}, nil
		},
	}

	// Create request with query parameters
	req := httptest.NewRequest("GET", "/api/v1/leaderboards/avg_hits?mode=simple&date_from=2025-01-01T00:00:00Z&date_to=2025-12-31T23:59:59Z&limit=10&offset=5", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metric", "avg_hits")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetLeaderboard(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetLeaderboard_InvalidMetric(t *testing.T) {
	mockSvc := &mockLeaderboardService{
		getLeaderboardFunc: func(ctx context.Context, req services.LeaderboardRequest) ([]services.LeaderboardEntry, error) {
			return nil, fmt.Errorf("invalid metric: invalid_metric")
		},
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/leaderboards/invalid_metric", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metric", "invalid_metric")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetLeaderboard(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestGetLeaderboard_MissingMetric(t *testing.T) {
	mockSvc := &mockLeaderboardService{}

	// Create request without metric
	req := httptest.NewRequest("GET", "/api/v1/leaderboards/", nil)
	w := httptest.NewRecorder()

	// Add chi URL params (empty metric)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metric", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetLeaderboard(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetLeaderboard_InvalidLimit(t *testing.T) {
	mockSvc := &mockLeaderboardService{
		getLeaderboardFunc: func(ctx context.Context, req services.LeaderboardRequest) ([]services.LeaderboardEntry, error) {
			// Should use default limit of 50 for invalid input
			if req.Limit != 50 {
				t.Errorf("expected default limit 50 for invalid input, got %d", req.Limit)
			}
			return []services.LeaderboardEntry{}, nil
		},
	}

	// Create request with invalid limit
	req := httptest.NewRequest("GET", "/api/v1/leaderboards/quina_rate?limit=invalid", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metric", "quina_rate")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetLeaderboard(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetLeaderboard_LargeLimit(t *testing.T) {
	mockSvc := &mockLeaderboardService{
		getLeaderboardFunc: func(ctx context.Context, req services.LeaderboardRequest) ([]services.LeaderboardEntry, error) {
			// Should cap at 1000
			if req.Limit != 1000 {
				t.Errorf("expected limit capped at 1000, got %d", req.Limit)
			}
			return []services.LeaderboardEntry{}, nil
		},
	}

	// Create request with large limit
	req := httptest.NewRequest("GET", "/api/v1/leaderboards/quina_rate?limit=2000", nil)
	w := httptest.NewRecorder()

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metric", "quina_rate")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Call handler
	handler := GetLeaderboard(mockSvc)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
