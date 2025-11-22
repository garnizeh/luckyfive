package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"

	"github.com/garnizeh/luckyfive/internal/store/comparisons"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
)

type ComparisonService struct {
	comparisonQueries comparisons.Querier
	simulationQueries simulations.Querier
	comparisonDB      *sql.DB
	logger            *slog.Logger
}

type ComparisonServicer interface {
	Compare(ctx context.Context, req CompareRequest) (*ComparisonResult, error)
	GetComparison(ctx context.Context, id int64) (*ComparisonResult, error)
	ListComparisons(ctx context.Context, limit, offset int) ([]comparisons.Comparison, error)
}

func NewComparisonService(
	comparisonQueries comparisons.Querier,
	simulationQueries simulations.Querier,
	comparisonDB *sql.DB,
	logger *slog.Logger,
) *ComparisonService {
	return &ComparisonService{
		comparisonQueries: comparisonQueries,
		simulationQueries: simulationQueries,
		comparisonDB:      comparisonDB,
		logger:            logger,
	}
}

type CompareRequest struct {
	Name          string
	Description   string
	SimulationIDs []int64
	Metrics       []string // ["quina_rate", "avg_hits", "roi"]
}

type ComparisonResult struct {
	ID             int64                       `json:"id"`
	Name           string                      `json:"name"`
	Description    string                      `json:"description,omitempty"`
	SimulationIDs  []int64                     `json:"simulation_ids"`
	Metrics        []string                    `json:"metrics"`
	Rankings       map[string][]SimulationRank `json:"rankings"`         // metric -> ranked list
	Statistics     map[string]MetricStats      `json:"statistics"`       // metric -> stats
	WinnerByMetric map[string]int64            `json:"winner_by_metric"` // metric -> simulation_id
	CreatedAt      string                      `json:"created_at"`
}

type SimulationRank struct {
	SimulationID   int64   `json:"simulation_id"`
	SimulationName string  `json:"simulation_name"`
	Value          float64 `json:"value"`
	Rank           int     `json:"rank"`
	Percentile     float64 `json:"percentile"`
}

type MetricStats struct {
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	StdDev float64 `json:"std_dev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  int     `json:"count"`
}

func (s *ComparisonService) validateCompareRequest(req CompareRequest) error {
	if len(req.SimulationIDs) < 2 {
		return fmt.Errorf("need at least 2 simulations to compare")
	}

	// Validate metrics
	for _, metric := range req.Metrics {
		if !s.isValidMetric(metric) {
			return fmt.Errorf("invalid metric: %s", metric)
		}
	}

	return nil
}

func (s *ComparisonService) Compare(
	ctx context.Context,
	req CompareRequest,
) (*ComparisonResult, error) {
	// Apply defaults
	if len(req.Metrics) == 0 {
		req.Metrics = []string{"quina_rate", "avg_hits"} // default metrics
	}

	if err := s.validateCompareRequest(req); err != nil {
		return nil, err
	}

	// Fetch all simulations
	simulations := make(map[int64]*simulations.Simulation)
	for _, id := range req.SimulationIDs {
		sim, err := s.simulationQueries.GetSimulation(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get simulation %d: %w", id, err)
		}
		if sim.Status != "completed" {
			return nil, fmt.Errorf("simulation %d is not completed (status: %s)", id, sim.Status)
		}
		simulations[id] = &sim
	}

	s.logger.Info("comparing simulations",
		"count", len(simulations),
		"metrics", req.Metrics)

	// Create comparison record
	simIDsJSON, _ := json.Marshal(req.SimulationIDs)
	metricsStr := strings.Join(req.Metrics, ",")

	comp, err := s.comparisonQueries.CreateComparison(ctx, comparisons.CreateComparisonParams{
		Name:          req.Name,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		SimulationIds: string(simIDsJSON),
		Metric:        metricsStr,
	})
	if err != nil {
		return nil, fmt.Errorf("create comparison: %w", err)
	}

	result := &ComparisonResult{
		ID:             comp.ID,
		Name:           req.Name,
		Description:    req.Description,
		SimulationIDs:  req.SimulationIDs,
		Metrics:        req.Metrics,
		Rankings:       make(map[string][]SimulationRank),
		Statistics:     make(map[string]MetricStats),
		WinnerByMetric: make(map[string]int64),
		CreatedAt:      comp.CreatedAt.String,
	}

	// Calculate each metric
	for _, metric := range req.Metrics {
		ranks, stats := s.calculateMetric(metric, simulations)
		result.Rankings[metric] = ranks
		result.Statistics[metric] = stats

		if len(ranks) > 0 {
			result.WinnerByMetric[metric] = ranks[0].SimulationID
		}

		// Store in database
		for _, rank := range ranks {
			err = s.comparisonQueries.InsertComparisonMetric(ctx, comparisons.InsertComparisonMetricParams{
				ComparisonID: comp.ID,
				SimulationID: rank.SimulationID,
				MetricName:   metric,
				MetricValue:  rank.Value,
				Rank:         sql.NullInt64{Int64: int64(rank.Rank), Valid: true},
				Percentile:   sql.NullFloat64{Float64: rank.Percentile, Valid: true},
			})
			if err != nil {
				return nil, fmt.Errorf("insert comparison metric %s for sim %d: %w", metric, rank.SimulationID, err)
			}
		}
	}

	// Save result JSON
	resultJSON, _ := json.Marshal(result)
	err = s.comparisonQueries.UpdateComparisonResult(ctx, comparisons.UpdateComparisonResultParams{
		ID:         comp.ID,
		ResultJson: sql.NullString{String: string(resultJSON), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("update comparison result: %w", err)
	}

	s.logger.Info("comparison completed",
		"id", comp.ID,
		"simulations", len(simulations),
		"metrics", len(req.Metrics))

	return result, nil
}

func (s *ComparisonService) GetComparison(ctx context.Context, id int64) (*ComparisonResult, error) {
	comp, err := s.comparisonQueries.GetComparison(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get comparison: %w", err)
	}

	if !comp.ResultJson.Valid {
		return nil, fmt.Errorf("comparison %d has no results yet", id)
	}

	var result ComparisonResult
	if err := json.Unmarshal([]byte(comp.ResultJson.String), &result); err != nil {
		return nil, fmt.Errorf("unmarshal comparison result: %w", err)
	}

	return &result, nil
}

func (s *ComparisonService) ListComparisons(ctx context.Context, limit, offset int) ([]comparisons.Comparison, error) {
	return s.comparisonQueries.ListComparisons(ctx, comparisons.ListComparisonsParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
}

func (s *ComparisonService) calculateMetric(
	metric string,
	simulations map[int64]*simulations.Simulation,
) ([]SimulationRank, MetricStats) {
	values := make(map[int64]float64)

	for id, sim := range simulations {
		if !sim.SummaryJson.Valid {
			s.logger.Warn("simulation has no summary", "id", id)
			continue
		}

		var summary Summary
		if err := json.Unmarshal([]byte(sim.SummaryJson.String), &summary); err != nil {
			s.logger.Error("failed to unmarshal summary", "id", id, "error", err)
			continue
		}

		switch metric {
		case "quina_rate":
			values[id] = summary.HitRateQuina
		case "quadra_rate":
			values[id] = summary.HitRateQuadra
		case "terno_rate":
			values[id] = summary.HitRateTerno
		case "avg_hits":
			values[id] = summary.AverageHits
		case "total_quinaz":
			values[id] = float64(summary.QuinaHits)
		case "total_quadras":
			values[id] = float64(summary.QuadraHits)
		case "total_ternos":
			values[id] = float64(summary.TernoHits)
		case "hit_efficiency":
			// Custom metric: average hits per contest
			if summary.TotalContests > 0 {
				values[id] = summary.AverageHits
			} else {
				values[id] = 0
			}
		default:
			values[id] = 0
		}
	}

	// Sort by value (descending - higher is better)
	type pair struct {
		id    int64
		value float64
	}

	pairs := make([]pair, 0, len(values))
	for id, val := range values {
		pairs = append(pairs, pair{id, val})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value > pairs[j].value
	})

	// Build rankings
	ranks := make([]SimulationRank, len(pairs))
	for i, p := range pairs {
		percentile := 100.0 * float64(len(pairs)-i) / float64(len(pairs))
		ranks[i] = SimulationRank{
			SimulationID:   p.id,
			SimulationName: simulations[p.id].RecipeName.String,
			Value:          p.value,
			Rank:           i + 1,
			Percentile:     percentile,
		}
	}

	// Calculate statistics
	stats := s.calculateStats(values)

	return ranks, stats
}

func (s *ComparisonService) calculateStats(values map[int64]float64) MetricStats {
	if len(values) == 0 {
		return MetricStats{}
	}

	vals := make([]float64, 0, len(values))
	for _, v := range values {
		vals = append(vals, v)
	}

	sort.Float64s(vals)

	// Mean
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	mean := sum / float64(len(vals))

	// StdDev
	variance := 0.0
	for _, v := range vals {
		variance += math.Pow(v-mean, 2)
	}
	stddev := math.Sqrt(variance / float64(len(vals)))

	// Median
	median := vals[len(vals)/2]
	if len(vals)%2 == 0 {
		median = (vals[len(vals)/2-1] + vals[len(vals)/2]) / 2
	}

	return MetricStats{
		Mean:   mean,
		Median: median,
		StdDev: stddev,
		Min:    vals[0],
		Max:    vals[len(vals)-1],
		Count:  len(vals),
	}
}

func (s *ComparisonService) isValidMetric(metric string) bool {
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
