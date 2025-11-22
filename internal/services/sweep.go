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

	// Start transaction
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

	// Create child simulations
	for i, recipe := range recipes {
		// Convert sweep recipe to service recipe
		serviceRecipe := s.convertToServiceRecipe(recipe)

		// Create simulation (async mode)
		sim, err := s.simulationService.CreateSimulation(ctx, CreateSimulationRequest{
			Mode:         "sweep",
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

		// Link to sweep
		paramsJSON, _ := json.Marshal(recipe.Parameters)
		err = txQueries.CreateSweepSimulation(ctx, sweep_execution.CreateSweepSimulationParams{
			SweepJobID:      sweepJob.ID,
			SimulationID:    sim.ID,
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
	params := RecipeParameters{}

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
