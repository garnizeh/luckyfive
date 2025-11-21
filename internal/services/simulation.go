package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/garnizeh/luckyfive/internal/store/simulations"
	"github.com/garnizeh/luckyfive/pkg/predictor"
)

type SimulationService struct {
	simulationsQueries simulations.Querier // Mockable
	simulationsDB      *sql.DB             // For transactions
	engineService      *EngineService
	logger             *slog.Logger
}

func NewSimulationService(
	simulationsQueries simulations.Querier,
	simulationsDB *sql.DB,
	engineService *EngineService,
	logger *slog.Logger,
) *SimulationService {
	return &SimulationService{
		simulationsQueries: simulationsQueries,
		simulationsDB:      simulationsDB,
		engineService:      engineService,
		logger:             logger,
	}
}

type CreateSimulationRequest struct {
	Mode         string
	RecipeName   string
	Recipe       Recipe
	StartContest int
	EndContest   int
	Async        bool
	CreatedBy    string
}

type Recipe struct {
	Version    string           `json:"version"`
	Name       string           `json:"name"`
	Parameters RecipeParameters `json:"parameters"`
}

type RecipeParameters struct {
	Alpha              float64 `json:"alpha"`
	Beta               float64 `json:"beta"`
	Gamma              float64 `json:"gamma"`
	Delta              float64 `json:"delta"`
	SimPrevMax         int     `json:"sim_prev_max"`
	SimPreds           int     `json:"sim_preds"`
	EnableEvolutionary bool    `json:"enableEvolutionary"`
	Generations        int     `json:"generations"`
	MutationRate       float64 `json:"mutationRate"`
}

func (s *SimulationService) CreateSimulation(
	ctx context.Context,
	req CreateSimulationRequest,
) (*simulations.Simulation, error) {
	// Validate recipe
	if err := s.validateRecipe(req.Recipe); err != nil {
		return nil, fmt.Errorf("invalid recipe: %w", err)
	}

	// Marshal recipe to JSON
	recipeJSON, err := json.Marshal(req.Recipe)
	if err != nil {
		return nil, fmt.Errorf("marshal recipe: %w", err)
	}

	// Create simulation record
	sim, err := s.simulationsQueries.CreateSimulation(ctx, simulations.CreateSimulationParams{
		RecipeName:   sql.NullString{String: req.RecipeName, Valid: req.RecipeName != ""},
		RecipeJson:   string(recipeJSON),
		Mode:         req.Mode,
		StartContest: int64(req.StartContest),
		EndContest:   int64(req.EndContest),
		CreatedBy:    sql.NullString{String: req.CreatedBy, Valid: req.CreatedBy != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("create simulation: %w", err)
	}

	// If sync mode, execute immediately
	if !req.Async {
		if err := s.ExecuteSimulation(ctx, sim.ID); err != nil {
			return nil, fmt.Errorf("execute simulation: %w", err)
		}

		// Reload to get updated status
		sim, err = s.simulationsQueries.GetSimulation(ctx, sim.ID)
		if err != nil {
			return nil, fmt.Errorf("reload simulation: %w", err)
		}
	}

	return &sim, nil
}

func (s *SimulationService) ExecuteSimulation(ctx context.Context, simID int64) error {
	// Get simulation
	sim, err := s.simulationsQueries.GetSimulation(ctx, simID)
	if err != nil {
		return fmt.Errorf("get simulation: %w", err)
	}

	// Parse recipe
	var recipe Recipe
	if err := json.Unmarshal([]byte(sim.RecipeJson), &recipe); err != nil {
		return fmt.Errorf("unmarshal recipe: %w", err)
	}

	// Build engine config
	engineCfg := SimulationConfig{
		StartContest: int(sim.StartContest),
		EndContest:   int(sim.EndContest),
		SimPrevMax:   recipe.Parameters.SimPrevMax,
		SimPreds:     recipe.Parameters.SimPreds,
		Weights: predictor.Weights{
			Alpha: recipe.Parameters.Alpha,
			Beta:  recipe.Parameters.Beta,
			Gamma: recipe.Parameters.Gamma,
			Delta: recipe.Parameters.Delta,
		},
		Seed:            sim.ID, // Use simulation ID as seed for reproducibility
		EnableEvolution: recipe.Parameters.EnableEvolutionary,
		Generations:     recipe.Parameters.Generations,
		MutationRate:    recipe.Parameters.MutationRate,
	}

	// Run simulation
	result, err := s.engineService.RunSimulation(ctx, engineCfg)
	if err != nil {
		// Mark as failed
		s.simulationsQueries.FailSimulation(ctx, simulations.FailSimulationParams{
			ID:           simID,
			FinishedAt:   sql.NullString{String: time.Now().Format(time.RFC3339), Valid: true},
			ErrorMessage: sql.NullString{String: err.Error(), Valid: true},
		})
		return fmt.Errorf("run simulation: %w", err)
	}

	// Save results in transaction
	tx, err := s.simulationsDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txQueries := simulations.New(tx)

	// Insert contest results
	for _, cr := range result.ContestResults {
		actualJSON, _ := json.Marshal(cr.ActualNumbers)
		predJSON, _ := json.Marshal(cr.BestPrediction)
		allPredsJSON, _ := json.Marshal(cr.AllPredictions)

		err = txQueries.InsertContestResult(ctx, simulations.InsertContestResultParams{
			SimulationID:          simID,
			Contest:               int64(cr.Contest),
			ActualNumbers:         string(actualJSON),
			BestHits:              int64(cr.BestHits),
			BestPredictionIndex:   sql.NullInt64{Int64: int64(cr.BestPredictionIndex), Valid: true},
			BestPredictionNumbers: sql.NullString{String: string(predJSON), Valid: true},
			PredictionsJson:       string(allPredsJSON),
		})
		if err != nil {
			return fmt.Errorf("insert contest result: %w", err)
		}
	}

	// Update simulation status
	summaryJSON, _ := json.Marshal(result.Summary)
	outputJSON, _ := json.Marshal(result)

	err = txQueries.CompleteSimulation(ctx, simulations.CompleteSimulationParams{
		ID:            simID,
		FinishedAt:    sql.NullString{String: time.Now().Format(time.RFC3339), Valid: true},
		RunDurationMs: sql.NullInt64{Int64: result.DurationMs, Valid: true},
		SummaryJson:   sql.NullString{String: string(summaryJSON), Valid: true},
		OutputBlob:    outputJSON,
		OutputName:    sql.NullString{String: fmt.Sprintf("simulation_%d.json", simID), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("complete simulation: %w", err)
	}

	return tx.Commit()
}

func (s *SimulationService) GetSimulation(ctx context.Context, id int64) (*simulations.Simulation, error) {
	sim, err := s.simulationsQueries.GetSimulation(ctx, id)
	if err != nil {
		return nil, err
	}
	return &sim, nil
}

func (s *SimulationService) CancelSimulation(ctx context.Context, id int64) error {
	return s.simulationsQueries.CancelSimulation(ctx, simulations.CancelSimulationParams{
		ID:         id,
		FinishedAt: sql.NullString{String: time.Now().Format(time.RFC3339), Valid: true},
	})
}

// validateRecipe validates the recipe parameters
func (s *SimulationService) validateRecipe(recipe Recipe) error {
	if recipe.Version == "" {
		return fmt.Errorf("recipe version is required")
	}
	if recipe.Parameters.SimPrevMax <= 0 {
		return fmt.Errorf("sim_prev_max must be positive")
	}
	if recipe.Parameters.SimPreds <= 0 {
		return fmt.Errorf("sim_preds must be positive")
	}
	// Add more validations as needed
	return nil
}

// ListSimulations lists simulations with pagination
func (s *SimulationService) ListSimulations(ctx context.Context, limit, offset int) ([]simulations.Simulation, error) {
	return s.simulationsQueries.ListSimulations(ctx, simulations.ListSimulationsParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
}

// ListSimulationsByStatus lists simulations by status with pagination
func (s *SimulationService) ListSimulationsByStatus(ctx context.Context, status string, limit, offset int) ([]simulations.Simulation, error) {
	return s.simulationsQueries.ListSimulationsByStatus(ctx, simulations.ListSimulationsByStatusParams{
		Status: status,
		Limit:  int64(limit),
		Offset: int64(offset),
	})
}
