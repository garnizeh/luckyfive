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

func TestSweepService_GetVisualizationData(t *testing.T) {
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
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			started_at TEXT,
			finished_at TEXT,
			status TEXT NOT NULL,
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
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data
	sweepConfig := `{
		"name": "test_sweep",
		"parameters": [
			{"name": "alpha", "type": "range", "values": {"min": 0.0, "max": 1.0, "step": 0.1}},
			{"name": "beta", "type": "range", "values": {"min": 0.0, "max": 1.0, "step": 0.1}}
		]
	}`

	_, err = db.Exec(`
		INSERT INTO sweep_jobs (id, name, sweep_config_json, base_contest_range, status, total_combinations)
		VALUES (1, 'test_sweep', ?, '1-100', 'completed', 2)
	`, sweepConfig)
	if err != nil {
		t.Fatalf("Failed to insert sweep job: %v", err)
	}

	// Insert test simulations
	_, err = db.Exec(`
		INSERT INTO simulations (id, status, recipe_json, summary_json, mode, start_contest, end_contest)
		VALUES (1, 'completed', '{"version":"1.0","name":"test","parameters":{}}', '{"hitRateQuina":0.05,"hitRateQuadra":0.1,"hitRateTerno":0.2,"averageHits":2.3,"quinaHits":5,"quadraHits":10,"ternoHits":20,"totalContests":100}', 'historical', 1, 100)
	`)
	if err != nil {
		t.Fatalf("Failed to insert simulation 1: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO simulations (id, status, recipe_json, summary_json, mode, start_contest, end_contest)
		VALUES (2, 'completed', '{"version":"1.0","name":"test","parameters":{}}', '{"hitRateQuina":0.03,"hitRateQuadra":0.08,"hitRateTerno":0.15,"averageHits":1.8,"quinaHits":3,"quadraHits":8,"ternoHits":15,"totalContests":100}', 'historical', 1, 100)
	`)
	if err != nil {
		t.Fatalf("Failed to insert simulation 2: %v", err)
	}

	// Insert sweep simulations
	_, err = db.Exec(`
		INSERT INTO sweep_simulations (sweep_job_id, simulation_id, variation_index, variation_params)
		VALUES (1, 1, 0, '{"alpha":0.1,"beta":0.2}')
	`)
	if err != nil {
		t.Fatalf("Failed to insert sweep simulation 1: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO sweep_simulations (sweep_job_id, simulation_id, variation_index, variation_params)
		VALUES (1, 2, 1, '{"alpha":0.2,"beta":0.3}')
	`)
	if err != nil {
		t.Fatalf("Failed to insert sweep simulation 2: %v", err)
	}

	// Create service
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sweepQueries := sweep_execution.New(db)

	mockSimSvc := NewMockSimulationServicer(ctrl)
	mockSimSvc.EXPECT().GetSimulation(gomock.Any(), int64(1)).Return(&simulations.Simulation{
		ID:         1,
		Status:     "completed",
		RecipeJson: `{"version":"1.0","name":"test","parameters":{}}`,
		SummaryJson: sql.NullString{
			String: `{"hitRateQuina":0.05,"hitRateQuadra":0.1,"hitRateTerno":0.2,"averageHits":2.3,"quinaHits":5,"quadraHits":10,"ternoHits":20,"totalContests":100}`,
			Valid:  true,
		},
	}, nil).AnyTimes()

	mockSimSvc.EXPECT().GetSimulation(gomock.Any(), int64(2)).Return(&simulations.Simulation{
		ID:         2,
		Status:     "completed",
		RecipeJson: `{"version":"1.0","name":"test","parameters":{}}`,
		SummaryJson: sql.NullString{
			String: `{"hitRateQuina":0.03,"hitRateQuadra":0.08,"hitRateTerno":0.15,"averageHits":1.8,"quinaHits":3,"quadraHits":8,"ternoHits":15,"totalContests":100}`,
			Valid:  true,
		},
	}, nil).AnyTimes()

	svc := NewSweepService(sweepQueries, db, mockSimSvc, logger)

	// Test with specific metrics
	data, err := svc.GetVisualizationData(context.Background(), 1, []string{"quina_rate", "avg_hits"})
	if err != nil {
		t.Fatalf("GetVisualizationData failed: %v", err)
	}

	if data.SweepID != 1 {
		t.Errorf("Expected sweep ID 1, got %d", data.SweepID)
	}

	if len(data.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(data.Parameters))
	}

	if data.Parameters[0] != "alpha" || data.Parameters[1] != "beta" {
		t.Errorf("Expected parameters [alpha, beta], got %v", data.Parameters)
	}

	if len(data.Metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(data.Metrics))
	}

	if len(data.DataPoints) != 2 {
		t.Errorf("Expected 2 data points, got %d", len(data.DataPoints))
	}

	// Check first data point
	point1 := data.DataPoints[0]
	if point1.Params["alpha"] != 0.1 || point1.Params["beta"] != 0.2 {
		t.Errorf("Expected params {alpha:0.1, beta:0.2}, got %v", point1.Params)
	}

	if point1.Metrics["quina_rate"] != 0.05 || point1.Metrics["avg_hits"] != 2.3 {
		t.Errorf("Expected metrics {quina_rate:0.05, avg_hits:2.3}, got %v", point1.Metrics)
	}

	// Check second data point
	point2 := data.DataPoints[1]
	if point2.Params["alpha"] != 0.2 || point2.Params["beta"] != 0.3 {
		t.Errorf("Expected params {alpha:0.2, beta:0.3}, got %v", point2.Params)
	}

	if point2.Metrics["quina_rate"] != 0.03 || point2.Metrics["avg_hits"] != 1.8 {
		t.Errorf("Expected metrics {quina_rate:0.03, avg_hits:1.8}, got %v", point2.Metrics)
	}
}

func TestSweepService_GetVisualizationData_DefaultMetrics(t *testing.T) {
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
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			started_at TEXT,
			finished_at TEXT,
			status TEXT NOT NULL,
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
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data
	sweepConfig := `{
		"name": "test_sweep",
		"parameters": [
			{"name": "alpha", "type": "range", "values": {"min": 0.0, "max": 1.0, "step": 0.1}}
		]
	}`

	_, err = db.Exec(`
		INSERT INTO sweep_jobs (id, name, sweep_config_json, base_contest_range, status, total_combinations)
		VALUES (1, 'test_sweep', ?, '1-100', 'completed', 1)
	`, sweepConfig)
	if err != nil {
		t.Fatalf("Failed to insert sweep job: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO simulations (id, status, recipe_json, summary_json, mode, start_contest, end_contest)
		VALUES (1, 'completed', '{"version":"1.0","name":"test","parameters":{}}', '{"hitRateQuina":0.05,"hitRateQuadra":0.1,"hitRateTerno":0.2,"averageHits":2.3,"quinaHits":5,"quadraHits":10,"ternoHits":20,"totalContests":100}', 'historical', 1, 100)
	`)
	if err != nil {
		t.Fatalf("Failed to insert simulation: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO sweep_simulations (sweep_job_id, simulation_id, variation_index, variation_params)
		VALUES (1, 1, 0, '{"alpha":0.1}')
	`)
	if err != nil {
		t.Fatalf("Failed to insert sweep simulation: %v", err)
	}

	// Create service
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sweepQueries := sweep_execution.New(db)

	mockSimSvc := NewMockSimulationServicer(ctrl)
	mockSimSvc.EXPECT().GetSimulation(gomock.Any(), int64(1)).Return(&simulations.Simulation{
		ID:         1,
		Status:     "completed",
		RecipeJson: `{"version":"1.0","name":"test","parameters":{}}`,
		SummaryJson: sql.NullString{
			String: `{"hitRateQuina":0.05,"hitRateQuadra":0.1,"hitRateTerno":0.2,"averageHits":2.3,"quinaHits":5,"quadraHits":10,"ternoHits":20,"totalContests":100}`,
			Valid:  true,
		},
	}, nil).AnyTimes()

	svc := NewSweepService(sweepQueries, db, mockSimSvc, logger)

	// Test with no metrics specified (should use defaults)
	data, err := svc.GetVisualizationData(context.Background(), 1, []string{})
	if err != nil {
		t.Fatalf("GetVisualizationData failed: %v", err)
	}

	if len(data.Metrics) != 2 {
		t.Errorf("Expected 2 default metrics, got %d", len(data.Metrics))
	}

	if data.Metrics[0] != "quina_rate" || data.Metrics[1] != "avg_hits" {
		t.Errorf("Expected default metrics [quina_rate, avg_hits], got %v", data.Metrics)
	}
}

func TestSweepService_GetVisualizationData_InvalidMetric(t *testing.T) {
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
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data
	sweepConfig := `{"name": "test_sweep", "parameters": []}`

	_, err = db.Exec(`
		INSERT INTO sweep_jobs (id, name, sweep_config_json, base_contest_range, status, total_combinations)
		VALUES (1, 'test_sweep', ?, '1-100', 'completed', 0)
	`, sweepConfig)
	if err != nil {
		t.Fatalf("Failed to insert sweep job: %v", err)
	}

	// Create service
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sweepQueries := sweep_execution.New(db)

	mockSimSvc := NewMockSimulationServicer(ctrl)

	svc := NewSweepService(sweepQueries, db, mockSimSvc, logger)

	// Test with invalid metric
	_, err = svc.GetVisualizationData(context.Background(), 1, []string{"invalid_metric"})
	if err == nil {
		t.Error("Expected error for invalid metric, got nil")
	}

	if err.Error() != "invalid metric: invalid_metric" {
		t.Errorf("Expected 'invalid metric: invalid_metric', got '%s'", err.Error())
	}
}
