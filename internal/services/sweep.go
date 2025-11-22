package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/garnizeh/luckyfive/internal/store/sweep_execution"
	"github.com/garnizeh/luckyfive/pkg/sweep"
)

type SweepServicer interface {
	CreateSweep(ctx context.Context, req CreateSweepRequest) (*sweep_execution.SweepJob, error)
	GetSweepStatus(ctx context.Context, sweepID int64) (*SweepStatus, error)
	UpdateSweepProgress(ctx context.Context, sweepID int64) error
	FindBest(ctx context.Context, sweepID int64, metric string) (*BestConfiguration, error)
	GetVisualizationData(ctx context.Context, sweepID int64, metrics []string) (*VisualizationData, error)
}

type SweepService struct {
	sweepExecutionQueries sweep_execution.Querier
	simulationsDB         *sql.DB
	simulationService     SimulationServicer
	generator             *sweep.Generator
	logger                *slog.Logger
}

func NewSweepService(
	sweepExecutionQueries sweep_execution.Querier,
	simulationsDB *sql.DB,
	simulationService SimulationServicer,
	logger *slog.Logger,
) *SweepService {
	return &SweepService{
		sweepExecutionQueries: sweepExecutionQueries,
		simulationsDB:         simulationsDB,
		simulationService:     simulationService,
		generator:             sweep.NewGenerator(),
		logger:                logger,
	}
}

type CreateSweepRequest struct {
	Name         string
	Description  string
	SweepConfig  sweep.SweepConfig
	StartContest int
	EndContest   int
	CreatedBy    string
}

type SweepStatus struct {
	Sweep       sweep_execution.SweepJob
	Total       int
	Completed   int
	Running     int
	Failed      int
	Pending     int
	Simulations []sweep_execution.GetSweepSimulationDetailsRow
}

type BestConfiguration struct {
	SweepID         int64              `json:"sweep_id"`
	SimulationID    int64              `json:"simulation_id"`
	Recipe          Recipe             `json:"recipe"`
	Metrics         map[string]float64 `json:"metrics"`
	Rank            int                `json:"rank"`
	Percentile      float64            `json:"percentile"`
	VariationParams map[string]any     `json:"variation_params"`
}

type VisualizationData struct {
	SweepID    int64                    `json:"sweep_id"`
	Parameters []string                 `json:"parameters"`
	Metrics    []string                 `json:"metrics"`
	DataPoints []VisualizationDataPoint `json:"data_points"`
}

type VisualizationDataPoint struct {
	Params  map[string]any     `json:"params"`
	Metrics map[string]float64 `json:"metrics"`
}

func (s *SweepService) CreateSweep(
	ctx context.Context,
	req CreateSweepRequest,
) (*sweep_execution.SweepJob, error) {
	// Validate sweep config
	if err := req.SweepConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid sweep config: %w", err)
	}

	// Generate all recipe combinations
	recipes, err := s.generator.Generate(req.SweepConfig)
	if err != nil {
		return nil, fmt.Errorf("generate recipes: %w", err)
	}

	if len(recipes) == 0 {
		return nil, fmt.Errorf("no valid combinations generated")
	}

	s.logger.Info("generated recipes", "count", len(recipes))

	// Create child simulations first (outside transaction to avoid deadlocks)
	var simulationIDs []int64
	for i, recipe := range recipes {
		// Convert sweep recipe to service recipe
		serviceRecipe := s.convertToServiceRecipe(recipe)

		// Create simulation (async mode)
		sim, err := s.simulationService.CreateSimulation(ctx, CreateSimulationRequest{
			Mode:         "simple",
			RecipeName:   fmt.Sprintf("%s_var_%d", req.Name, i),
			Recipe:       serviceRecipe,
			StartContest: req.StartContest,
			EndContest:   req.EndContest,
			Async:        true,
			CreatedBy:    req.CreatedBy,
		})
		if err != nil {
			return nil, fmt.Errorf("create simulation %d: %w", i, err)
		}

		simulationIDs = append(simulationIDs, sim.ID)
	}

	// Start transaction for sweep job creation and linking
	tx, err := s.simulationsDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txQueries := sweep_execution.New(tx)

	// Create sweep job record
	sweepConfigJSON, _ := json.Marshal(req.SweepConfig)
	contestRange := fmt.Sprintf("%d-%d", req.StartContest, req.EndContest)

	sweepJob, err := txQueries.CreateSweepJob(ctx, sweep_execution.CreateSweepJobParams{
		Name:              req.Name,
		Description:       sql.NullString{String: req.Description, Valid: req.Description != ""},
		SweepConfigJson:   string(sweepConfigJSON),
		BaseContestRange:  contestRange,
		TotalCombinations: int64(len(recipes)),
		CreatedBy:         sql.NullString{String: req.CreatedBy, Valid: req.CreatedBy != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("create sweep job: %w", err)
	}

	// Link simulations to sweep
	for i, recipe := range recipes {
		// Link to sweep
		paramsJSON, _ := json.Marshal(recipe.Parameters)
		err = txQueries.CreateSweepSimulation(ctx, sweep_execution.CreateSweepSimulationParams{
			SweepJobID:      sweepJob.ID,
			SimulationID:    simulationIDs[i],
			VariationIndex:  int64(i),
			VariationParams: string(paramsJSON),
		})
		if err != nil {
			return nil, fmt.Errorf("link simulation %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &sweepJob, nil
}

func (s *SweepService) convertToServiceRecipe(recipe sweep.GeneratedRecipe) Recipe {
	params := RecipeParameters{
		// Set defaults
		SimPrevMax: 10,
		SimPreds:   5,
	}

	// Extract parameters from the map
	if alpha, ok := recipe.Parameters["alpha"].(float64); ok {
		params.Alpha = alpha
	}
	if beta, ok := recipe.Parameters["beta"].(float64); ok {
		params.Beta = beta
	}
	if gamma, ok := recipe.Parameters["gamma"].(float64); ok {
		params.Gamma = gamma
	}
	if delta, ok := recipe.Parameters["delta"].(float64); ok {
		params.Delta = delta
	}
	if simPrevMax, ok := recipe.Parameters["sim_prev_max"].(float64); ok {
		params.SimPrevMax = int(simPrevMax)
	}
	if simPreds, ok := recipe.Parameters["sim_preds"].(float64); ok {
		params.SimPreds = int(simPreds)
	}
	if enableEvolutionary, ok := recipe.Parameters["enableEvolutionary"].(bool); ok {
		params.EnableEvolutionary = enableEvolutionary
	}
	if generations, ok := recipe.Parameters["generations"].(float64); ok {
		params.Generations = int(generations)
	}
	if mutationRate, ok := recipe.Parameters["mutationRate"].(float64); ok {
		params.MutationRate = mutationRate
	}

	return Recipe{
		Version:    "1.0",
		Name:       recipe.Name,
		Parameters: params,
	}
}

func (s *SweepService) GetSweepStatus(ctx context.Context, sweepID int64) (*SweepStatus, error) {
	sweepJob, err := s.sweepExecutionQueries.GetSweepJob(ctx, sweepID)
	if err != nil {
		return nil, fmt.Errorf("get sweep job: %w", err)
	}

	details, err := s.sweepExecutionQueries.GetSweepSimulationDetails(ctx, sweepID)
	if err != nil {
		return nil, fmt.Errorf("get sweep simulation details: %w", err)
	}

	status := &SweepStatus{
		Sweep:       sweepJob,
		Total:       len(details),
		Completed:   0,
		Running:     0,
		Failed:      0,
		Pending:     0,
		Simulations: details,
	}

	for _, detail := range details {
		switch detail.Status {
		case "completed":
			status.Completed++
		case "running":
			status.Running++
		case "failed":
			status.Failed++
		case "pending":
			status.Pending++
		}
	}

	return status, nil
}

func (s *SweepService) UpdateSweepProgress(ctx context.Context, sweepID int64) error {
	status, err := s.GetSweepStatus(ctx, sweepID)
	if err != nil {
		return err
	}

	// Determine overall status
	var overallStatus string
	if status.Failed > 0 && status.Completed+status.Failed == status.Total {
		overallStatus = "failed"
	} else if status.Completed == status.Total {
		overallStatus = "completed"
	} else if status.Running > 0 || status.Completed > 0 {
		overallStatus = "running"
	} else {
		overallStatus = "pending"
	}

	// Update progress
	err = s.sweepExecutionQueries.UpdateSweepJobProgress(ctx, sweep_execution.UpdateSweepJobProgressParams{
		CompletedSimulations: sql.NullInt64{Int64: int64(status.Completed), Valid: true},
		FailedSimulations:    sql.NullInt64{Int64: int64(status.Failed), Valid: true},
		Status:               overallStatus,
		ID:                   sweepID,
	})
	if err != nil {
		return fmt.Errorf("update sweep progress: %w", err)
	}

	// If completed, mark finish time
	if overallStatus == "completed" || overallStatus == "failed" {
		duration := int64(0)
		if status.Sweep.StartedAt.Valid {
			startTime, _ := time.Parse(time.RFC3339, status.Sweep.StartedAt.String)
			duration = time.Since(startTime).Milliseconds()
		}

		err = s.sweepExecutionQueries.FinishSweepJob(ctx, sweep_execution.FinishSweepJobParams{
			Status:        overallStatus,
			FinishedAt:    sql.NullString{String: time.Now().Format(time.RFC3339), Valid: true},
			RunDurationMs: sql.NullInt64{Int64: duration, Valid: true},
			ID:            sweepID,
		})
		if err != nil {
			return fmt.Errorf("finish sweep job: %w", err)
		}
	}

	return nil
}

func (s *SweepService) FindBest(
	ctx context.Context,
	sweepID int64,
	metric string,
) (*BestConfiguration, error) {
	// Validate metric
	if !s.isValidMetric(metric) {
		return nil, fmt.Errorf("invalid metric: %s", metric)
	}

	// Get sweep status to ensure it exists and get simulation details
	status, err := s.GetSweepStatus(ctx, sweepID)
	if err != nil {
		return nil, fmt.Errorf("get sweep status: %w", err)
	}

	if status.Sweep.Status != "completed" {
		return nil, fmt.Errorf("sweep %d is not completed (status: %s)", sweepID, status.Sweep.Status)
	}

	// Calculate metrics for all completed simulations
	type simResult struct {
		simulationID    int64
		variationParams map[string]any
		metrics         map[string]float64
	}

	results := make([]simResult, 0, len(status.Simulations))

	for _, sim := range status.Simulations {
		if sim.Status != "completed" || !sim.SummaryJson.Valid {
			continue
		}

		var summary Summary
		if err := json.Unmarshal([]byte(sim.SummaryJson.String), &summary); err != nil {
			s.logger.Error("failed to unmarshal summary", "simulation_id", sim.SimulationID, "error", err)
			continue
		}

		// Parse variation params
		var variationParams map[string]any
		if err := json.Unmarshal([]byte(sim.VariationParams), &variationParams); err != nil {
			s.logger.Error("failed to unmarshal variation params", "simulation_id", sim.SimulationID, "error", err)
			continue
		}

		// Calculate all metrics
		metrics := s.calculateMetrics(&summary)

		results = append(results, simResult{
			simulationID:    sim.SimulationID,
			variationParams: variationParams,
			metrics:         metrics,
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no completed simulations found in sweep %d", sweepID)
	}

	// Find the best result for the specified metric
	bestIndex := 0
	bestValue := results[0].metrics[metric]

	for i := 1; i < len(results); i++ {
		if results[i].metrics[metric] > bestValue {
			bestValue = results[i].metrics[metric]
			bestIndex = i
		}
	}

	best := results[bestIndex]

	// Calculate rank and percentile
	rank := 1
	for _, result := range results {
		if result.metrics[metric] > best.metrics[metric] {
			rank++
		}
	}

	percentile := 100.0 * float64(len(results)-rank+1) / float64(len(results))

	// Get the recipe for this simulation
	simulation, err := s.simulationService.GetSimulation(ctx, best.simulationID)
	if err != nil {
		return nil, fmt.Errorf("get simulation %d: %w", best.simulationID, err)
	}

	var recipe Recipe
	if err := json.Unmarshal([]byte(simulation.RecipeJson), &recipe); err != nil {
		return nil, fmt.Errorf("unmarshal recipe: %w", err)
	}

	return &BestConfiguration{
		SweepID:         sweepID,
		SimulationID:    best.simulationID,
		Recipe:          recipe,
		Metrics:         best.metrics,
		Rank:            rank,
		Percentile:      percentile,
		VariationParams: best.variationParams,
	}, nil
}

func (s *SweepService) calculateMetrics(summary *Summary) map[string]float64 {
	metrics := make(map[string]float64)

	metrics["quina_rate"] = summary.HitRateQuina
	metrics["quadra_rate"] = summary.HitRateQuadra
	metrics["terno_rate"] = summary.HitRateTerno
	metrics["avg_hits"] = summary.AverageHits
	metrics["total_quinaz"] = float64(summary.QuinaHits)
	metrics["total_quadras"] = float64(summary.QuadraHits)
	metrics["total_ternos"] = float64(summary.TernoHits)

	// Custom metrics
	if summary.TotalContests > 0 {
		metrics["hit_efficiency"] = summary.AverageHits / float64(summary.TotalContests)
	} else {
		metrics["hit_efficiency"] = 0
	}

	return metrics
}

func (s *SweepService) isValidMetric(metric string) bool {
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

func (s *SweepService) GetVisualizationData(
	ctx context.Context,
	sweepID int64,
	metrics []string,
) (*VisualizationData, error) {
	// Validate metrics
	for _, metric := range metrics {
		if !s.isValidMetric(metric) {
			return nil, fmt.Errorf("invalid metric: %s", metric)
		}
	}

	// If no metrics specified, use defaults
	if len(metrics) == 0 {
		metrics = []string{"quina_rate", "avg_hits"}
	}

	// Get sweep status to ensure it exists and get simulation details
	status, err := s.GetSweepStatus(ctx, sweepID)
	if err != nil {
		return nil, fmt.Errorf("get sweep status: %w", err)
	}

	// Parse sweep config to get parameter names
	var sweepConfig sweep.SweepConfig
	if err := json.Unmarshal([]byte(status.Sweep.SweepConfigJson), &sweepConfig); err != nil {
		return nil, fmt.Errorf("unmarshal sweep config: %w", err)
	}

	// Extract parameter names
	parameters := make([]string, 0, len(sweepConfig.Parameters))
	for _, param := range sweepConfig.Parameters {
		parameters = append(parameters, param.Name)
	}

	// Collect data points from completed simulations
	dataPoints := make([]VisualizationDataPoint, 0, len(status.Simulations))

	for _, sim := range status.Simulations {
		if sim.Status != "completed" || !sim.SummaryJson.Valid {
			continue
		}

		var summary Summary
		if err := json.Unmarshal([]byte(sim.SummaryJson.String), &summary); err != nil {
			s.logger.Error("failed to unmarshal summary", "simulation_id", sim.SimulationID, "error", err)
			continue
		}

		// Parse variation params
		var variationParams map[string]any
		if err := json.Unmarshal([]byte(sim.VariationParams), &variationParams); err != nil {
			s.logger.Error("failed to unmarshal variation params", "simulation_id", sim.SimulationID, "error", err)
			continue
		}

		// Calculate requested metrics
		pointMetrics := make(map[string]float64)
		allMetrics := s.calculateMetrics(&summary)

		for _, metric := range metrics {
			if value, exists := allMetrics[metric]; exists {
				pointMetrics[metric] = value
			}
		}

		dataPoints = append(dataPoints, VisualizationDataPoint{
			Params:  variationParams,
			Metrics: pointMetrics,
		})
	}

	return &VisualizationData{
		SweepID:    sweepID,
		Parameters: parameters,
		Metrics:    metrics,
		DataPoints: dataPoints,
	}, nil
}
