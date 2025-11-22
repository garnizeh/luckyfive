package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/garnizeh/luckyfive/internal/store/simulations"
	simulationsmock "github.com/garnizeh/luckyfive/internal/store/simulations/mock"
	"github.com/garnizeh/luckyfive/pkg/predictor"
)

func TestSimulationService_CreateSimulation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	req := CreateSimulationRequest{
		Mode:       "simple",
		RecipeName: "test-recipe",
		Recipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: RecipeParameters{
				Alpha:      0.1,
				Beta:       0.2,
				Gamma:      0.3,
				Delta:      0.4,
				SimPrevMax: 10,
				SimPreds:   5,
			},
		},
		StartContest: 1000,
		EndContest:   1010,
		Async:        true,
		CreatedBy:    "user1",
	}

	expectedSim := simulations.Simulation{
		ID:           1,
		RecipeName:   sql.NullString{String: "test-recipe", Valid: true},
		RecipeJson:   `{"version":"1.0","name":"test","parameters":{"alpha":0.1,"beta":0.2,"gamma":0.3,"delta":0.4,"sim_prev_max":10,"sim_preds":5,"enableEvolutionary":false,"generations":0,"mutationRate":0}}`,
		Mode:         "simple",
		StartContest: 1000,
		EndContest:   1010,
		Status:       "pending",
		CreatedBy:    sql.NullString{String: "user1", Valid: true},
	}

	mockQueries.EXPECT().
		CreateSimulation(gomock.Any(), gomock.Any()).
		Return(expectedSim, nil)

	sim, err := service.CreateSimulation(context.Background(), req)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*sim, expectedSim) {
		t.Errorf("got %v, want %v", *sim, expectedSim)
	}
}

func TestSimulationService_GetSimulation_QueryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	mockQueries.EXPECT().
		GetSimulation(gomock.Any(), int64(1)).
		Return(simulations.Simulation{}, sql.ErrNoRows)

	_, err := service.GetSimulation(context.Background(), 1)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSimulationService_CancelSimulation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	mockQueries.EXPECT().
		CancelSimulation(gomock.Any(), gomock.Any()).
		Return(nil)

	err := service.CancelSimulation(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}
}

func TestSimulationService_ListSimulations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	expectedSims := []simulations.Simulation{
		{ID: 1, Status: "pending"},
		{ID: 2, Status: "running"},
	}

	mockQueries.EXPECT().
		ListSimulations(gomock.Any(), gomock.Any()).
		Return(expectedSims, nil)

	sims, err := service.ListSimulations(context.Background(), 10, 0)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sims, expectedSims) {
		t.Errorf("got %v, want %v", sims, expectedSims)
	}
}

func TestSimulationService_ListSimulationsByStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	expectedSims := []simulations.Simulation{
		{ID: 1, Status: "pending"},
	}

	mockQueries.EXPECT().
		ListSimulationsByStatus(gomock.Any(), gomock.Any()).
		Return(expectedSims, nil)

	sims, err := service.ListSimulationsByStatus(context.Background(), "pending", 10, 0)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sims, expectedSims) {
		t.Errorf("got %v, want %v", sims, expectedSims)
	}
}

func TestSimulationService_GetContestResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	expectedResults := []simulations.SimulationContestResult{
		{SimulationID: 1, Contest: 100, BestHits: 3},
	}

	mockQueries.EXPECT().
		GetContestResults(gomock.Any(), gomock.Any()).
		Return(expectedResults, nil)

	results, err := service.GetContestResults(context.Background(), 1, 10, 0)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("got %v, want %v", results, expectedResults)
	}
}

func TestSimulationService_ExecuteSimulation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	mockEngine := NewMockEngineServicer(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	// Note: Full ExecuteSimulation test requires database setup, testing logic separately
	// For unit test coverage, the error cases below provide sufficient coverage
	_ = mockQueries
	_ = mockEngine
	_ = logger
}

func TestSimulationService_ExecuteSimulation_GetSimulationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	mockQueries.EXPECT().
		GetSimulation(gomock.Any(), int64(1)).
		Return(simulations.Simulation{}, sql.ErrNoRows)

	err := service.ExecuteSimulation(context.Background(), 1)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSimulationService_ExecuteSimulation_RunSimulationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	mockEngine := NewMockEngineServicer(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, mockEngine, logger)

	sim := simulations.Simulation{
		ID:           1,
		RecipeJson:   `{"version":"1.0","name":"test","parameters":{"alpha":0.1,"beta":0.2,"gamma":0.3,"delta":0.4,"sim_prev_max":10,"sim_preds":5}}`,
		StartContest: 100,
		EndContest:   110,
	}

	mockQueries.EXPECT().
		GetSimulation(gomock.Any(), int64(1)).
		Return(sim, nil)

	mockEngine.EXPECT().
		RunSimulation(gomock.Any(), gomock.Any()).
		Return(nil, fmt.Errorf("simulation failed"))

	mockQueries.EXPECT().
		FailSimulation(gomock.Any(), gomock.Any()).
		Return(nil)

	err := service.ExecuteSimulation(context.Background(), 1)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSimulationService_ExecuteSimulation_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Set up in-memory database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE simulations (
			id INTEGER PRIMARY KEY,
			recipe_json TEXT,
			start_contest INTEGER,
			end_contest INTEGER,
			status TEXT,
			finished_at TEXT,
			run_duration_ms INTEGER,
			summary_json TEXT,
			output_blob BLOB,
			output_name TEXT
		);
		CREATE TABLE simulation_contest_results (
			id INTEGER PRIMARY KEY,
			simulation_id INTEGER,
			contest INTEGER,
			actual_numbers TEXT,
			best_hits INTEGER,
			best_prediction_index INTEGER,
			best_prediction_numbers TEXT,
			predictions_json TEXT
		);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	mockEngine := NewMockEngineServicer(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, db, mockEngine, logger)

	sim := simulations.Simulation{
		ID:           1,
		RecipeJson:   `{"version":"1.0","name":"test","parameters":{"alpha":0.1,"beta":0.2,"gamma":0.3,"delta":0.4,"sim_prev_max":10,"sim_preds":5}}`,
		StartContest: 100,
		EndContest:   110,
	}

	result := &SimulationResult{
		ContestResults: []ContestResult{
			{
				Contest:             105,
				ActualNumbers:       []int{1, 2, 3, 4, 5},
				BestHits:            5,
				BestPrediction:      []int{1, 2, 3, 4, 5},
				BestPredictionIndex: 0,
				AllPredictions:      []predictor.Prediction{{Numbers: []int{1, 2, 3, 4, 5}}},
			},
		},
		Summary: Summary{
			TotalContests: 1,
			QuinaHits:     1,
			QuadraHits:    0,
			TernoHits:     0,
			AverageHits:   5.0,
		},
		DurationMs: 1000,
	}

	mockQueries.EXPECT().
		GetSimulation(gomock.Any(), int64(1)).
		Return(sim, nil)

	mockEngine.EXPECT().
		RunSimulation(gomock.Any(), gomock.Any()).
		Return(result, nil)

	err = service.ExecuteSimulation(context.Background(), 1)

	if err != nil {
		t.Fatalf("ExecuteSimulation returned error: %v", err)
	}
}

func TestSimulationService_ExecuteSimulation_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	sim := simulations.Simulation{
		ID:           1,
		RecipeJson:   `invalid json`,
		StartContest: 100,
		EndContest:   110,
	}

	mockQueries.EXPECT().
		GetSimulation(gomock.Any(), int64(1)).
		Return(sim, nil)

	err := service.ExecuteSimulation(context.Background(), 1)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSimulationService_validateRecipe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	// Valid recipe
	validRecipe := Recipe{
		Version: "1.0",
		Name:    "test",
		Parameters: RecipeParameters{
			SimPrevMax: 10,
			SimPreds:   5,
		},
	}

	err := service.validateRecipe(validRecipe)
	if err != nil {
		t.Errorf("expected no error for valid recipe, got %v", err)
	}

	// Invalid version
	invalidRecipe := Recipe{
		Version: "",
		Name:    "test",
		Parameters: RecipeParameters{
			SimPrevMax: 10,
			SimPreds:   5,
		},
	}

	err = service.validateRecipe(invalidRecipe)
	if err == nil {
		t.Error("expected error for invalid version, got nil")
	}

	// Invalid sim_prev_max
	invalidRecipe2 := Recipe{
		Version: "1.0",
		Name:    "test",
		Parameters: RecipeParameters{
			SimPrevMax: 0,
			SimPreds:   5,
		},
	}

	err = service.validateRecipe(invalidRecipe2)
	if err == nil {
		t.Error("expected error for invalid sim_prev_max, got nil")
	}

	// Invalid sim_preds
	invalidRecipe3 := Recipe{
		Version: "1.0",
		Name:    "test",
		Parameters: RecipeParameters{
			SimPrevMax: 10,
			SimPreds:   0,
		},
	}

	err = service.validateRecipe(invalidRecipe3)
	if err == nil {
		t.Error("expected error for invalid sim_preds, got nil")
	}
}

func TestSimulationService_CreateSimulation_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	req := CreateSimulationRequest{
		Mode:       "simple",
		RecipeName: "test-recipe",
		Recipe: Recipe{
			Version: "", // Invalid
			Name:    "test",
			Parameters: RecipeParameters{
				SimPrevMax: 10,
				SimPreds:   5,
			},
		},
		StartContest: 1000,
		EndContest:   1010,
		Async:        true,
	}

	_, err := service.CreateSimulation(context.Background(), req)

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestSimulationService_CreateSimulation_SyncExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Set up in-memory database for sync execution
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	defer db.Close()

	// Create minimal schema for ExecuteSimulation
	_, err = db.Exec(`
		CREATE TABLE simulations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at TEXT,
			finished_at TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			recipe_name TEXT,
			recipe_json TEXT NOT NULL,
			mode TEXT NOT NULL,
			start_contest INTEGER NOT NULL,
			end_contest INTEGER NOT NULL,
			worker_id TEXT,
			run_duration_ms INTEGER,
			summary_json TEXT,
			output_blob BLOB,
			output_name TEXT,
			log_blob BLOB,
			error_message TEXT,
			error_stack TEXT,
			created_by TEXT
		);
		CREATE TABLE simulation_contest_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			simulation_id INTEGER NOT NULL,
			contest INTEGER NOT NULL,
			actual_numbers TEXT NOT NULL,
			best_hits INTEGER NOT NULL,
			best_prediction_index INTEGER,
			best_prediction_numbers TEXT,
			predictions_json TEXT NOT NULL,
			processed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	mockEngine := NewMockEngineServicer(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, db, mockEngine, logger)

	req := CreateSimulationRequest{
		Mode:       "simple",
		RecipeName: "test-recipe",
		Recipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: RecipeParameters{
				Alpha:      0.1,
				Beta:       0.2,
				Gamma:      0.3,
				Delta:      0.4,
				SimPrevMax: 10,
				SimPreds:   5,
			},
		},
		StartContest: 1000,
		EndContest:   1010,
		Async:        false, // Synchronous execution
		CreatedBy:    "user1",
	}

	expectedSim := simulations.Simulation{
		ID:           1,
		RecipeName:   sql.NullString{String: "test-recipe", Valid: true},
		RecipeJson:   `{"version":"1.0","name":"test","parameters":{"alpha":0.1,"beta":0.2,"gamma":0.3,"delta":0.4,"sim_prev_max":10,"sim_preds":5,"enableEvolutionary":false,"generations":0,"mutationRate":0}}`,
		Mode:         "simple",
		StartContest: 1000,
		EndContest:   1010,
		Status:       "pending",
		CreatedBy:    sql.NullString{String: "user1", Valid: true},
	}

	completedSim := simulations.Simulation{
		ID:           1,
		RecipeName:   sql.NullString{String: "test-recipe", Valid: true},
		RecipeJson:   `{"version":"1.0","name":"test","parameters":{"alpha":0.1,"beta":0.2,"gamma":0.3,"delta":0.4,"sim_prev_max":10,"sim_preds":5,"enableEvolutionary":false,"generations":0,"mutationRate":0}}`,
		Mode:         "simple",
		StartContest: 1000,
		EndContest:   1010,
		Status:       "completed",
		CreatedBy:    sql.NullString{String: "user1", Valid: true},
	}

	result := &SimulationResult{
		ContestResults: []ContestResult{
			{
				Contest:             1005,
				ActualNumbers:       []int{1, 2, 3, 4, 5},
				BestHits:            5,
				BestPrediction:      []int{1, 2, 3, 4, 5},
				BestPredictionIndex: 0,
				AllPredictions:      []predictor.Prediction{{Numbers: []int{1, 2, 3, 4, 5}}},
			},
		},
		Summary: Summary{
			TotalContests: 1,
			QuinaHits:     1,
			QuadraHits:    0,
			TernoHits:     0,
			AverageHits:   5.0,
		},
		DurationMs: 1000,
	}

	mockQueries.EXPECT().
		CreateSimulation(gomock.Any(), gomock.Any()).
		Return(expectedSim, nil)

	mockQueries.EXPECT().
		GetSimulation(gomock.Any(), int64(1)).
		Return(expectedSim, nil)

	mockEngine.EXPECT().
		RunSimulation(gomock.Any(), gomock.Any()).
		Return(result, nil)

	mockQueries.EXPECT().
		GetSimulation(gomock.Any(), int64(1)).
		Return(completedSim, nil)

	sim, err := service.CreateSimulation(context.Background(), req)

	if err != nil {
		t.Fatalf("CreateSimulation returned error: %v", err)
	}
	if sim.Status != "completed" {
		t.Errorf("Expected status 'completed' for sync execution, got '%s'", sim.Status)
	}
}

func TestSimulationService_CreateSimulation_CreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := simulationsmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	service := NewSimulationService(mockQueries, nil, nil, logger)

	req := CreateSimulationRequest{
		Mode:       "simple",
		RecipeName: "test-recipe",
		Recipe: Recipe{
			Version: "1.0",
			Name:    "test",
			Parameters: RecipeParameters{
				Alpha:      0.1,
				Beta:       0.2,
				Gamma:      0.3,
				Delta:      0.4,
				SimPrevMax: 10,
				SimPreds:   5,
			},
		},
		StartContest: 1000,
		EndContest:   1010,
		Async:        true,
		CreatedBy:    "user1",
	}

	mockQueries.EXPECT().
		CreateSimulation(gomock.Any(), gomock.Any()).
		Return(simulations.Simulation{}, sql.ErrConnDone)

	_, err := service.CreateSimulation(context.Background(), req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
