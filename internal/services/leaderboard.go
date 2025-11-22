package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/garnizeh/luckyfive/internal/store/simulations"
)

type LeaderboardService struct {
	simulationQueries simulations.Querier
	logger            *slog.Logger
}

type LeaderboardServicer interface {
	GetLeaderboard(ctx context.Context, req LeaderboardRequest) ([]LeaderboardEntry, error)
}

func NewLeaderboardService(
	simulationQueries simulations.Querier,
	logger *slog.Logger,
) *LeaderboardService {
	return &LeaderboardService{
		simulationQueries: simulationQueries,
		logger:            logger,
	}
}

type LeaderboardEntry struct {
	Rank           int     `json:"rank"`
	SimulationID   int64   `json:"simulation_id"`
	SimulationName string  `json:"simulation_name"`
	RecipeName     string  `json:"recipe_name"`
	MetricValue    float64 `json:"metric_value"`
	CreatedAt      string  `json:"created_at"`
	CreatedBy      string  `json:"created_by"`
}

type LeaderboardRequest struct {
	Metric   string `json:"metric"`
	Mode     string `json:"mode"`      // "simple", "advanced", "sweep", "all"
	DateFrom string `json:"date_from"` // RFC3339 format
	DateTo   string `json:"date_to"`   // RFC3339 format
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

func (s *LeaderboardService) GetLeaderboard(
	ctx context.Context,
	req LeaderboardRequest,
) ([]LeaderboardEntry, error) {
	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}
	if req.Metric == "" {
		req.Metric = "quina_rate"
	}
	if req.Mode == "" {
		req.Mode = "all"
	}

	// Validate metric
	if !s.isValidMetric(req.Metric) {
		return nil, fmt.Errorf("invalid metric: %s", req.Metric)
	}

	// Fetch simulations with filters
	simulations, err := s.fetchSimulations(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fetch simulations: %w", err)
	}

	// Calculate metrics for each simulation
	entries := make([]LeaderboardEntry, 0, len(simulations))
	for _, sim := range simulations {
		if !sim.SummaryJson.Valid {
			continue // Skip simulations without summary
		}

		var summary Summary
		if err := json.Unmarshal([]byte(sim.SummaryJson.String), &summary); err != nil {
			if s.logger != nil {
				s.logger.Warn("failed to unmarshal summary", "simulation_id", sim.ID, "error", err)
			}
			continue
		}

		metricValue := s.calculateMetricValue(req.Metric, summary)

		entry := LeaderboardEntry{
			SimulationID:   sim.ID,
			SimulationName: fmt.Sprintf("Simulation %d", sim.ID),
			RecipeName:     sim.RecipeName.String,
			MetricValue:    metricValue,
			CreatedAt:      sim.CreatedAt,
			CreatedBy:      sim.CreatedBy.String,
		}
		entries = append(entries, entry)
	}

	// Sort by metric value (descending - higher is better)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].MetricValue > entries[j].MetricValue
	})

	// Apply pagination
	totalEntries := len(entries)
	start := req.Offset
	if start >= totalEntries {
		return []LeaderboardEntry{}, nil
	}
	end := start + req.Limit
	if end > totalEntries {
		end = totalEntries
	}

	// Assign ranks
	result := entries[start:end]
	for i := range result {
		result[i].Rank = start + i + 1
	}

	if s.logger != nil {
		s.logger.Info("leaderboard generated",
			"metric", req.Metric,
			"mode", req.Mode,
			"total_simulations", len(simulations),
			"ranked_entries", len(result))
	}

	return result, nil
}

func (s *LeaderboardService) fetchSimulations(
	ctx context.Context,
	req LeaderboardRequest,
) ([]simulations.Simulation, error) {
	// For leaderboard, we need all completed simulations
	// Use a large limit to get all simulations, then filter in memory
	allSims, err := s.simulationQueries.ListSimulations(ctx, simulations.ListSimulationsParams{
		Limit:  10000, // Large limit to get most simulations
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	var filtered []simulations.Simulation
	for _, sim := range allSims {
		// Filter by status
		if sim.Status != "completed" {
			continue
		}

		// Filter by mode
		if req.Mode != "all" && sim.Mode != req.Mode {
			continue
		}

		// Filter by date range
		if req.DateFrom != "" {
			fromTime, err := time.Parse(time.RFC3339, req.DateFrom)
			if err != nil {
				return nil, fmt.Errorf("invalid date_from format: %w", err)
			}
			// Parse sim.CreatedAt - SQLite stores as ISO 8601 format like "2025-11-22 14:30:45"
			simTime, err := time.Parse("2006-01-02 15:04:05", sim.CreatedAt)
			if err != nil {
				// Try RFC3339 format as fallback
				simTime, err = time.Parse(time.RFC3339, sim.CreatedAt)
				if err != nil {
					continue // Skip if can't parse
				}
			}
			if simTime.Before(fromTime) {
				continue
			}
		}

		if req.DateTo != "" {
			toTime, err := time.Parse(time.RFC3339, req.DateTo)
			if err != nil {
				return nil, fmt.Errorf("invalid date_to format: %w", err)
			}
			// Parse sim.CreatedAt - SQLite stores as ISO 8601 format like "2025-11-22 14:30:45"
			simTime, err := time.Parse("2006-01-02 15:04:05", sim.CreatedAt)
			if err != nil {
				// Try RFC3339 format as fallback
				simTime, err = time.Parse(time.RFC3339, sim.CreatedAt)
				if err != nil {
					continue // Skip if can't parse
				}
			}
			if simTime.After(toTime) {
				continue
			}
		}

		filtered = append(filtered, sim)
	}

	return filtered, nil
}

func (s *LeaderboardService) calculateMetricValue(metric string, summary Summary) float64 {
	switch metric {
	case "quina_rate":
		return summary.HitRateQuina
	case "quadra_rate":
		return summary.HitRateQuadra
	case "terno_rate":
		return summary.HitRateTerno
	case "avg_hits":
		return summary.AverageHits
	case "total_quinaz":
		return float64(summary.QuinaHits)
	case "total_quadras":
		return float64(summary.QuadraHits)
	case "total_ternos":
		return float64(summary.TernoHits)
	case "hit_efficiency":
		if summary.TotalContests > 0 {
			return summary.AverageHits
		}
		return 0
	default:
		return 0
	}
}

func (s *LeaderboardService) isValidMetric(metric string) bool {
	validMetrics := map[string]bool{
		"quina_rate":     true,
		"quadra_rate":    true,
		"terno_rate":     true,
		"avg_hits":       true,
		"total_quinaz":   true,
		"total_quadras":  true,
		"total_ternos":   true,
		"hit_efficiency": true,
	}
	return validMetrics[metric]
}
