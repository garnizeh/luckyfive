package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/garnizeh/luckyfive/internal/store"
	"github.com/garnizeh/luckyfive/internal/store/comparisons"
	"github.com/garnizeh/luckyfive/internal/store/results"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
	"github.com/garnizeh/luckyfive/internal/store/sweep_execution"
	"github.com/garnizeh/luckyfive/migrations"
	"github.com/garnizeh/luckyfive/pkg/sweep"
)

// setupSweepTestDB creates a test database with all required schemas including sweeps
func setupSweepTestDB(t *testing.T) *store.DB {
	t.Helper()

	// Create temporary databases
	resultsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create results DB: %v", err)
	}

	simulationsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create simulations DB: %v", err)
	}

	configsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create configs DB: %v", err)
	}

	financesDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create finances DB: %v", err)
	}

	sweepsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create sweeps DB: %v", err)
	}

	db := &store.DB{
		ResultsDB:     resultsDB,
		SimulationsDB: simulationsDB,
		ConfigsDB:     configsDB,
		FinancesDB:    financesDB,
		SweepsDB:      sweepsDB,
	}

	// Configure connection pools
	for _, sqlDB := range []*sql.DB{resultsDB, simulationsDB, configsDB, financesDB, sweepsDB} {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	}

	// Setup schemas
	if err := setupSweepSchemas(t, db); err != nil {
		t.Fatalf("Failed to setup schemas: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// setupSweepSchemas creates all required database schemas including sweep tables
func setupSweepSchemas(t *testing.T, db *store.DB) error {
	t.Helper()

	// Load and execute migration files
	migrationFiles := map[string]string{
		"results":     "001_create_results.sql",
		"simulations": "002_create_simulations.sql",
		"configs":     "003_create_configs.sql",
		"finances":    "004_create_finances.sql",
	}

	dbs := map[string]*sql.DB{
		"results":     db.ResultsDB,
		"simulations": db.SimulationsDB,
		"configs":     db.ConfigsDB,
		"finances":    db.FinancesDB,
	}

	for dbName, migrationFile := range migrationFiles {
		sqlDB, exists := dbs[dbName]
		if !exists {
			continue
		}

		// Read migration file
		content, err := migrations.Files.ReadFile(migrationFile)
		if err != nil {
			return fmt.Errorf("read migration file %s: %w", migrationFile, err)
		}

		// Execute migration
		if _, err := sqlDB.Exec(string(content)); err != nil {
			return fmt.Errorf("execute migration %s: %w", migrationFile, err)
		}
	}

	// Create sweep tables manually (since they might not be in migrations yet)
	sweepTables := `
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
    variation_params TEXT NOT NULL,
    FOREIGN KEY (sweep_job_id) REFERENCES sweep_jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
);

CREATE INDEX idx_sweep_simulations_sweep_job_id ON sweep_simulations(sweep_job_id);
CREATE INDEX idx_sweep_jobs_status ON sweep_jobs(status);

CREATE TABLE comparisons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    simulation_ids TEXT NOT NULL,
    metric TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    result_json TEXT
);

CREATE TABLE comparison_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    comparison_id INTEGER NOT NULL,
    simulation_id INTEGER NOT NULL,
    metric_name TEXT NOT NULL,
    metric_value REAL NOT NULL,
    rank INTEGER,
    percentile REAL,
    FOREIGN KEY (comparison_id) REFERENCES comparisons(id) ON DELETE CASCADE,
    FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
);

CREATE INDEX idx_comparison_metrics_comparison_id ON comparison_metrics(comparison_id);
`
	if _, err := db.SimulationsDB.Exec(sweepTables); err != nil {
		return fmt.Errorf("create sweep tables: %w", err)
	}

	// Insert sample data
	if err := insertSweepSampleData(t, db); err != nil {
		return fmt.Errorf("insert sample data: %w", err)
	}

	return nil
}

// insertSweepSampleData inserts test data into the databases
func insertSweepSampleData(t *testing.T, db *store.DB) error {
	t.Helper()

	// Sample draws for results DB
	resultsData := `
INSERT INTO draws (contest, draw_date, bola1, bola2, bola3, bola4, bola5, source) VALUES
(1000, '2024-01-01', 1, 2, 3, 4, 5, 'test'),
(1001, '2024-01-02', 6, 7, 8, 9, 10, 'test'),
(1002, '2024-01-03', 11, 12, 13, 14, 15, 'test'),
(1003, '2024-01-04', 16, 17, 18, 19, 20, 'test'),
(1004, '2024-01-05', 21, 22, 23, 24, 25, 'test');
`
	if _, err := db.ResultsDB.Exec(resultsData); err != nil {
		return fmt.Errorf("insert results data: %w", err)
	}

	// Sample config preset for configs DB
	configsData := `
INSERT INTO config_presets (name, display_name, description, recipe_json, risk_level, sort_order) VALUES
('test_preset', 'Test Preset', 'Test preset for integration tests', '{"version":"1.0","name":"Test","parameters":{"alpha":0.1,"beta":0.2,"gamma":0.3,"delta":0.4,"sim_prev_max":10,"sim_preds":5,"enableEvolutionary":false,"generations":0,"mutationRate":0.0}}', 'low', 1);
`
	if _, err := db.ConfigsDB.Exec(configsData); err != nil {
		return fmt.Errorf("insert configs data: %w", err)
	}

	return nil
}

// setupSweepServices creates all services with real database connections including sweep services
func setupSweepServices(t *testing.T, db *store.DB) (*services.SimulationService, *services.SweepService, *services.ComparisonService) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create queriers
	resultsQueries := results.New(db.ResultsDB)
	simulationsQueries := simulations.New(db.SimulationsDB)
	sweepExecutionQueries := sweep_execution.New(db.SimulationsDB)
	comparisonQueries := comparisons.New(db.SimulationsDB)

	// Create services
	engineService := services.NewEngineService(resultsQueries, logger)
	simulationService := services.NewSimulationService(simulationsQueries, db.SimulationsDB, engineService, logger)
	sweepService := services.NewSweepService(sweepExecutionQueries, db.SimulationsDB, simulationService, logger)
	comparisonService := services.NewComparisonService(comparisonQueries, simulationsQueries, db.SimulationsDB, logger)

	return simulationService, sweepService, comparisonService
}

func TestFullSweepFlow(t *testing.T) {
	db := setupSweepTestDB(t)

	simSvc, sweepSvc, _ := setupSweepServices(t, db)

	ctx := context.Background()

	// Create sweep configuration
	sweepConfig := sweep.SweepConfig{
		Name:        "integration_test_sweep",
		Description: "Test sweep for integration testing",
		BaseRecipe: sweep.Recipe{
			Version: "1.0",
			Name:    "base",
			Parameters: map[string]any{
				"sim_prev_max": 10,
				"sim_preds":    5,
				"gamma":        0.5,
				"delta":        0.5,
			},
		},
		Parameters: []sweep.ParameterSweep{
			{
				Name: "alpha",
				Type: "range",
				Values: sweep.RangeValues{
					Min:  0.0,
					Max:  0.2,
					Step: 0.1,
				},
			},
		},
	}

	// Create sweep
	sweepReq := services.CreateSweepRequest{
		Name:         "integration_test_sweep",
		Description:  "Test sweep for integration testing",
		SweepConfig:  sweepConfig,
		StartContest: 1001,
		EndContest:   1002, // Small range for fast testing
		CreatedBy:    "test_user",
	}

	sweepJob, err := sweepSvc.CreateSweep(ctx, sweepReq)
	if err != nil {
		t.Fatalf("Failed to create sweep: %v", err)
	}

	// Verify sweep was created
	if sweepJob.Name != "integration_test_sweep" {
		t.Errorf("Expected sweep name 'integration_test_sweep', got '%s'", sweepJob.Name)
	}

	if sweepJob.TotalCombinations != 3 { // alpha: 0.0, 0.1, 0.2
		t.Errorf("Expected 3 combinations, got %d", sweepJob.TotalCombinations)
	}

	if sweepJob.Status != "pending" {
		t.Errorf("Expected sweep status 'pending', got '%s'", sweepJob.Status)
	}

	// Get sweep status
	status, err := sweepSvc.GetSweepStatus(ctx, sweepJob.ID)
	if err != nil {
		t.Fatalf("Failed to get sweep status: %v", err)
	}

	if status.Total != 3 {
		t.Errorf("Expected 3 total simulations, got %d", status.Total)
	}

	if status.Completed != 0 {
		t.Errorf("Expected 0 completed simulations, got %d", status.Completed)
	}

	// For integration testing, we'll simulate the sweep processing
	// instead of using the complex worker system to avoid SQLite concurrency issues

	// Get the sweep simulations that were created
	sweepSims, err := sweepSvc.GetSweepStatus(ctx, sweepJob.ID)
	if err != nil {
		t.Fatalf("Failed to get sweep simulations: %v", err)
	}

	// Process each simulation synchronously
	for _, sweepSim := range sweepSims.Simulations {
		// Get the simulation
		sim, err := simSvc.GetSimulation(ctx, sweepSim.SimulationID)
		if err != nil {
			t.Fatalf("Failed to get simulation %d: %v", sweepSim.SimulationID, err)
		}

		// If not completed, run it synchronously
		if sim.Status != "completed" {
			err := simSvc.ExecuteSimulation(ctx, sweepSim.SimulationID)
			if err != nil {
				t.Fatalf("Failed to execute simulation %d synchronously: %v", sweepSim.SimulationID, err)
			}
		}
	}

	// Update sweep progress to mark as completed
	if err := sweepSvc.UpdateSweepProgress(ctx, sweepJob.ID); err != nil {
		t.Fatalf("Failed to update sweep progress: %v", err)
	}

	// Verify final sweep state
	finalStatus, err := sweepSvc.GetSweepStatus(ctx, sweepJob.ID)
	if err != nil {
		t.Fatalf("Failed to get final sweep status: %v", err)
	}

	if finalStatus.Completed != 3 {
		t.Errorf("Expected 3 completed simulations, got %d", finalStatus.Completed)
	}

	if finalStatus.Failed != 0 {
		t.Errorf("Expected 0 failed simulations, got %d", finalStatus.Failed)
	}

	// Test finding best configuration
	best, err := sweepSvc.FindBest(ctx, sweepJob.ID, "quina_rate")
	if err != nil {
		t.Fatalf("Failed to find best configuration: %v", err)
	}

	if best.SweepID != sweepJob.ID {
		t.Errorf("Expected best sweep ID %d, got %d", sweepJob.ID, best.SweepID)
	}

	if best.Rank != 1 {
		t.Errorf("Expected best rank 1, got %d", best.Rank)
	}

	// Test visualization data
	vizData, err := sweepSvc.GetVisualizationData(ctx, sweepJob.ID, []string{"quina_rate"})
	if err != nil {
		t.Fatalf("Failed to get visualization data: %v", err)
	}

	if vizData.SweepID != sweepJob.ID {
		t.Errorf("Expected viz sweep ID %d, got %d", sweepJob.ID, vizData.SweepID)
	}

	if len(vizData.DataPoints) != 3 {
		t.Errorf("Expected 3 data points, got %d", len(vizData.DataPoints))
	}

	// Verify data points have required fields
	for i, point := range vizData.DataPoints {
		if point.Params == nil {
			t.Errorf("Data point %d: missing parameters", i)
		}
		if point.Metrics == nil {
			t.Errorf("Data point %d: missing metrics", i)
		}
		if _, exists := point.Metrics["quina_rate"]; !exists {
			t.Errorf("Data point %d: missing quina_rate metric", i)
		}
	}
}

func TestComparisonFlow(t *testing.T) {
	db := setupSweepTestDB(t)

	simSvc, _, compSvc := setupSweepServices(t, db)

	ctx := context.Background()

	// Create multiple simulations for comparison
	var simulationIDs []int64

	// Create 3 different simulations with varying parameters
	recipes := []services.Recipe{
		{
			Version: "1.0",
			Name:    "sim1",
			Parameters: services.RecipeParameters{
				Alpha:      0.1,
				Beta:       0.2,
				Gamma:      0.3,
				Delta:      0.4,
				SimPrevMax: 10,
				SimPreds:   5,
			},
		},
		{
			Version: "1.0",
			Name:    "sim2",
			Parameters: services.RecipeParameters{
				Alpha:      0.2,
				Beta:       0.3,
				Gamma:      0.2,
				Delta:      0.3,
				SimPrevMax: 15,
				SimPreds:   7,
			},
		},
		{
			Version: "1.0",
			Name:    "sim3",
			Parameters: services.RecipeParameters{
				Alpha:      0.3,
				Beta:       0.4,
				Gamma:      0.1,
				Delta:      0.2,
				SimPrevMax: 20,
				SimPreds:   10,
			},
		},
	}

	for i, recipe := range recipes {
		sim, err := simSvc.CreateSimulation(ctx, services.CreateSimulationRequest{
			Mode:         "simple",
			RecipeName:   fmt.Sprintf("comparison_sim_%d", i+1),
			Recipe:       recipe,
			StartContest: 1001,
			EndContest:   1003,
			Async:        false, // Synchronous for simplicity
			CreatedBy:    "test_user",
		})
		if err != nil {
			t.Fatalf("Failed to create simulation %d: %v", i+1, err)
		}

		simulationIDs = append(simulationIDs, sim.ID)
	} // Create comparison
	compReq := services.CompareRequest{
		Name:          "integration_test_comparison",
		Description:   "Test comparison for integration testing",
		SimulationIDs: simulationIDs,
		Metrics:       []string{"quina_rate", "avg_hits"},
	}

	result, err := compSvc.Compare(ctx, compReq)
	if err != nil {
		t.Fatalf("Failed to create comparison: %v", err)
	}

	t.Log("Comparison created successfully")

	// Verify comparison result
	if result.Name != "integration_test_comparison" {
		t.Errorf("Expected comparison name 'integration_test_comparison', got '%s'", result.Name)
	}

	if len(result.Rankings) != 2 {
		t.Errorf("Expected 2 metrics in rankings, got %d", len(result.Rankings))
	}

	// Check quina_rate ranking
	quinaRankings, exists := result.Rankings["quina_rate"]
	if !exists {
		t.Error("Expected quina_rate rankings")
	} else {
		if len(quinaRankings) != 3 {
			t.Errorf("Expected 3 quina_rate rankings, got %d", len(quinaRankings))
		}

		// Verify rankings are sorted (descending by value)
		for i := 1; i < len(quinaRankings); i++ {
			if quinaRankings[i-1].Value < quinaRankings[i].Value {
				t.Errorf("Rankings not sorted correctly: %v vs %v", quinaRankings[i-1].Value, quinaRankings[i].Value)
			}
		}
	}

	// Check avg_hits ranking
	avgHitsRankings, exists := result.Rankings["avg_hits"]
	if !exists {
		t.Error("Expected avg_hits rankings")
	} else {
		if len(avgHitsRankings) != 3 {
			t.Errorf("Expected 3 avg_hits rankings, got %d", len(avgHitsRankings))
		}
		// Verify avg_hits has positive values since simulations are hitting some numbers
		for i, ranking := range avgHitsRankings {
			if ranking.Value <= 0 {
				t.Errorf("Expected positive avg_hits value for ranking %d, got %v", i+1, ranking.Value)
			}
		}
	}

	// Check winner by metric
	if _, exists := result.WinnerByMetric["quina_rate"]; !exists {
		t.Error("Expected winner for quina_rate")
	}

	if _, exists := result.WinnerByMetric["avg_hits"]; !exists {
		t.Error("Expected winner for avg_hits")
	}

	// Check statistics
	if len(result.Statistics) != 2 {
		t.Errorf("Expected 2 metrics in statistics, got %d", len(result.Statistics))
	}

	quinaStats, exists := result.Statistics["quina_rate"]
	if !exists {
		t.Error("Expected quina_rate statistics")
	} else {
		if quinaStats.Count != 3 {
			t.Errorf("Expected count 3, got %d", quinaStats.Count)
		}
		// Quina rate can be 0 if no quinas are hit, which is expected with test data
		if quinaStats.Mean < 0 {
			t.Errorf("Expected non-negative mean, got %v", quinaStats.Mean)
		}
	}
}

func TestLeaderboardGeneration(t *testing.T) {
	db := setupSweepTestDB(t)

	simSvc, _, _ := setupSweepServices(t, db)

	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	leaderboardSvc := services.NewLeaderboardService(simulations.New(db.SimulationsDB), logger)

	// Create multiple simulations with different performance levels
	recipes := []services.Recipe{
		{
			Version: "1.0",
			Name:    "high_perf",
			Parameters: services.RecipeParameters{
				Alpha:      0.5,
				Beta:       0.3,
				Gamma:      0.1,
				Delta:      0.1,
				SimPrevMax: 50,
				SimPreds:   20,
			},
		},
		{
			Version: "1.0",
			Name:    "med_perf",
			Parameters: services.RecipeParameters{
				Alpha:      0.3,
				Beta:       0.4,
				Gamma:      0.2,
				Delta:      0.1,
				SimPrevMax: 30,
				SimPreds:   15,
			},
		},
		{
			Version: "1.0",
			Name:    "low_perf",
			Parameters: services.RecipeParameters{
				Alpha:      0.1,
				Beta:       0.2,
				Gamma:      0.3,
				Delta:      0.4,
				SimPrevMax: 10,
				SimPreds:   5,
			},
		},
	}

	var simulationIDs []int64
	for i, recipe := range recipes {
		sim, err := simSvc.CreateSimulation(ctx, services.CreateSimulationRequest{
			Mode:         "simple",
			RecipeName:   fmt.Sprintf("leaderboard_sim_%d", i+1),
			Recipe:       recipe,
			StartContest: 1001,
			EndContest:   1004,
			Async:        false,
			CreatedBy:    "test_user",
		})
		if err != nil {
			t.Fatalf("Failed to create leaderboard simulation %d: %v", i+1, err)
		}

		simulationIDs = append(simulationIDs, sim.ID)
	}

	// Test leaderboard generation for quina_rate
	leaderboard, err := leaderboardSvc.GetLeaderboard(ctx, services.LeaderboardRequest{
		Metric: "quina_rate",
		Mode:   "all",
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to get leaderboard: %v", err)
	}

	if len(leaderboard) != 3 {
		t.Errorf("Expected 3 leaderboard entries, got %d", len(leaderboard))
	}

	// Verify leaderboard is sorted (descending by metric value)
	for i := 1; i < len(leaderboard); i++ {
		if leaderboard[i-1].MetricValue < leaderboard[i].MetricValue {
			t.Errorf("Leaderboard not sorted correctly: %v vs %v",
				leaderboard[i-1].MetricValue, leaderboard[i].MetricValue)
		}
		if leaderboard[i-1].Rank >= leaderboard[i].Rank {
			t.Errorf("Ranks not sequential: %d vs %d",
				leaderboard[i-1].Rank, leaderboard[i].Rank)
		}
	}

	// Test leaderboard with mode filter
	simpleLeaderboard, err := leaderboardSvc.GetLeaderboard(ctx, services.LeaderboardRequest{
		Metric: "quina_rate",
		Mode:   "simple",
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to get simple mode leaderboard: %v", err)
	}

	if len(simpleLeaderboard) != 3 {
		t.Errorf("Expected 3 simple mode entries, got %d", len(simpleLeaderboard))
	}

	// Test leaderboard with date filter (use current year to include test simulations)
	currentYear := time.Now().Year()
	fromDate := time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(currentYear, 12, 31, 23, 59, 59, 0, time.UTC)

	dateFiltered, err := leaderboardSvc.GetLeaderboard(ctx, services.LeaderboardRequest{
		Metric:   "quina_rate",
		Mode:     "all",
		DateFrom: fromDate.Format(time.RFC3339),
		DateTo:   toDate.Format(time.RFC3339),
		Limit:    10,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("Failed to get date-filtered leaderboard: %v", err)
	}

	if len(dateFiltered) != 3 {
		t.Errorf("Expected 3 date-filtered entries, got %d", len(dateFiltered))
	}

	// Test pagination
	paged, err := leaderboardSvc.GetLeaderboard(ctx, services.LeaderboardRequest{
		Metric: "quina_rate",
		Mode:   "all",
		Limit:  2,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to get paged leaderboard: %v", err)
	}

	if len(paged) != 2 {
		t.Errorf("Expected 2 paged entries, got %d", len(paged))
	}

	// Test different metric
	avgHitsLeaderboard, err := leaderboardSvc.GetLeaderboard(ctx, services.LeaderboardRequest{
		Metric: "avg_hits",
		Mode:   "all",
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to get avg_hits leaderboard: %v", err)
	}

	if len(avgHitsLeaderboard) != 3 {
		t.Errorf("Expected 3 avg_hits entries, got %d", len(avgHitsLeaderboard))
	}
}
