package services

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/store/simulations"
	"github.com/garnizeh/luckyfive/internal/store/sweep_execution"
	sweepmock "github.com/garnizeh/luckyfive/internal/store/sweep_execution/mock"
	"github.com/garnizeh/luckyfive/pkg/sweep"
)

func TestSweepService_CreateSweep(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create in-memory DB for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer db.Close()

	// Create required tables
	_, err = db.Exec(`
		CREATE TABLE sweep_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			sweep_config_json TEXT NOT NULL,
			base_contest_range TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			total_combinations INTEGER NOT NULL,
			completed_simulations INTEGER DEFAULT 0,
			failed_simulations INTEGER DEFAULT 0,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			started_at TEXT,
			finished_at TEXT,
			run_duration_ms INTEGER,
			created_by TEXT
		);
		CREATE TABLE sweep_simulations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sweep_job_id INTEGER NOT NULL,
			simulation_id INTEGER NOT NULL,
			variation_index INTEGER NOT NULL,
			variation_params TEXT NOT NULL
		);
		CREATE TABLE simulations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			recipe_name TEXT,
			recipe_json TEXT NOT NULL,
			mode TEXT NOT NULL,
			start_contest INTEGER NOT NULL,
			end_contest INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create real queriers
	sweepQueries := sweep_execution.New(db)
	mockSimService := NewMockSimulationServicer(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewSweepService(sweepQueries, db, mockSimService, logger)

	req := CreateSweepRequest{
		Name:        "test_sweep",
		Description: "Test sweep",
		SweepConfig: sweep.SweepConfig{
			Name: "test_sweep",
			BaseRecipe: sweep.Recipe{
				Version: "1.0",
				Name:    "test",
				Parameters: map[string]interface{}{
					"alpha": 0.5,
				},
			},
			Parameters: []sweep.ParameterSweep{
				{
					Name: "alpha",
					Type: "range",
					Values: sweep.RangeValues{
						Min:  0.0,
						Max:  1.0,
						Step: 0.5,
					},
				},
			},
		},
		StartContest: 1,
		EndContest:   10,
		CreatedBy:    "test_user",
	}

	expectedSim := simulations.Simulation{
		ID: 100,
	}

	mockSimService.EXPECT().
		CreateSimulation(gomock.Any(), gomock.Any()).
		Return(&expectedSim, nil).
		Times(3) // 3 combinations: alpha=0.0, 0.5, 1.0

	// Execute
	result, err := service.CreateSweep(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSweep failed: %v", err)
	}

	if result.ID == 0 {
		t.Errorf("Expected sweep ID to be set, got 0")
	}

	if result.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, result.Name)
	}

	if result.TotalCombinations != 3 {
		t.Errorf("Expected 3 combinations, got %d", result.TotalCombinations)
	}
}

func TestSweepService_GetSweepStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := sweepmock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(nil, nil))
	service := NewSweepService(mockQueries, nil, nil, logger)

	sweepID := int64(1)

	// Mock sweep job
	expectedSweepJob := sweep_execution.SweepJob{
		ID:                sweepID,
		Name:              "test_sweep",
		Status:            "running",
		TotalCombinations: 3,
	}

	mockQueries.EXPECT().
		GetSweepJob(gomock.Any(), sweepID).
		Return(expectedSweepJob, nil)

	// Mock simulation details
	details := []sweep_execution.GetSweepSimulationDetailsRow{
		{
			ID:              1,
			SweepJobID:      sweepID,
			SimulationID:    100,
			VariationIndex:  0,
			VariationParams: "{}",
			Status:          "completed",
		},
		{
			ID:              2,
			SweepJobID:      sweepID,
			SimulationID:    101,
			VariationIndex:  1,
			VariationParams: "{}",
			Status:          "running",
		},
		{
			ID:              3,
			SweepJobID:      sweepID,
			SimulationID:    102,
			VariationIndex:  2,
			VariationParams: "{}",
			Status:          "pending",
		},
	}

	mockQueries.EXPECT().
		GetSweepSimulationDetails(gomock.Any(), sweepID).
		Return(details, nil)

	// Execute
	status, err := service.GetSweepStatus(context.Background(), sweepID)
	if err != nil {
		t.Fatalf("GetSweepStatus failed: %v", err)
	}

	if status.Total != 3 {
		t.Errorf("Expected total 3, got %d", status.Total)
	}

	if status.Completed != 1 {
		t.Errorf("Expected completed 1, got %d", status.Completed)
	}

	if status.Running != 1 {
		t.Errorf("Expected running 1, got %d", status.Running)
	}

	if status.Pending != 1 {
		t.Errorf("Expected pending 1, got %d", status.Pending)
	}
}
