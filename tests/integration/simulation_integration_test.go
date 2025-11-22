package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/config"
	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/garnizeh/luckyfive/internal/store"
	"github.com/garnizeh/luckyfive/internal/store/configs"
	"github.com/garnizeh/luckyfive/internal/store/results"
	"github.com/garnizeh/luckyfive/internal/store/simulations"
	"github.com/garnizeh/luckyfive/internal/worker"
	"github.com/garnizeh/luckyfive/migrations"
)

// setupTestDB creates a test database with all required schemas
func setupTestDB(t *testing.T) (*store.DB, func()) {
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

	db := &store.DB{
		ResultsDB:     resultsDB,
		SimulationsDB: simulationsDB,
		ConfigsDB:     configsDB,
		FinancesDB:    financesDB,
	}

	// Configure connection pools
	for _, sqlDB := range []*sql.DB{resultsDB, simulationsDB, configsDB, financesDB} {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	}

	// Setup schemas
	if err := setupSchemas(t, db); err != nil {
		t.Fatalf("Failed to setup schemas: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

// setupSchemas creates all required database schemas
func setupSchemas(t *testing.T, db *store.DB) error {
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

	// Insert sample data
	if err := insertSampleData(t, db); err != nil {
		return fmt.Errorf("insert sample data: %w", err)
	}

	return nil
}

// insertSampleData inserts test data into the databases
func insertSampleData(t *testing.T, db *store.DB) error {
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

// setupServices creates all services with real database connections
func setupServices(t *testing.T, db *store.DB) (*services.SimulationService, *services.ConfigService, *services.EngineService, func()) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create queriers
	resultsQueries := results.New(db.ResultsDB)
	simulationsQueries := simulations.New(db.SimulationsDB)
	configsQueries := configs.New(db.ConfigsDB)

	// Create services
	engineService := services.NewEngineService(resultsQueries, logger)
	simulationService := services.NewSimulationService(simulationsQueries, db.SimulationsDB, engineService, logger)
	configService := services.NewConfigService(configsQueries, db.ConfigsDB, logger)

	cleanup := func() {
		// Services don't need explicit cleanup
	}

	return simulationService, configService, engineService, cleanup
}

func TestFullSimulationFlow(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	simSvc, configSvc, _, svcCleanup := setupServices(t, db)
	defer svcCleanup()

	ctx := context.Background()

	// Load test preset
	preset, err := configSvc.GetPreset(ctx, "test_preset")
	if err != nil {
		t.Fatalf("Failed to load test preset: %v", err)
	}

	// Parse preset recipe
	var recipe services.Recipe
	if err := json.Unmarshal([]byte(preset.RecipeJson), &recipe); err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	// Create simulation synchronously
	sim, err := simSvc.CreateSimulation(ctx, services.CreateSimulationRequest{
		Mode:         "simple",
		RecipeName:   "test_preset",
		Recipe:       recipe,
		StartContest: 1001, // Use contests that exist in sample data
		EndContest:   1003,
		Async:        false, // Synchronous execution
		CreatedBy:    "test_user",
	})
	if err != nil {
		t.Fatalf("Failed to create simulation: %v", err)
	}

	// Verify simulation was created and completed
	if sim.Status != "completed" {
		t.Errorf("Expected simulation status 'completed', got '%s'", sim.Status)
	}

	if !sim.RunDurationMs.Valid || sim.RunDurationMs.Int64 == 0 {
		t.Error("Expected non-zero run duration")
	}

	// Verify contest results were created
	results, err := simSvc.GetContestResults(ctx, sim.ID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get contest results: %v", err)
	}

	expectedContests := 3 // 1001, 1002, 1003
	if len(results) != expectedContests {
		t.Errorf("Expected %d contest results, got %d", expectedContests, len(results))
	}

	// Verify each result has predictions
	for i, result := range results {
		if result.BestHits < 0 || result.BestHits > 5 {
			t.Errorf("Result %d: invalid best hits %d", i, result.BestHits)
		}
		if len(result.PredictionsJson) == 0 {
			t.Errorf("Result %d: missing predictions JSON", i)
		}
	}
}

func TestWorkerJobProcessing(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	simSvc, configSvc, _, svcCleanup := setupServices(t, db)
	defer svcCleanup()

	ctx := context.Background()

	// Load test preset
	preset, err := configSvc.GetPreset(ctx, "test_preset")
	if err != nil {
		t.Fatalf("Failed to load test preset: %v", err)
	}

	// Parse preset recipe
	var recipe services.Recipe
	if err := json.Unmarshal([]byte(preset.RecipeJson), &recipe); err != nil {
		t.Fatalf("Failed to parse recipe: %v", err)
	}

	// Create simulation asynchronously
	sim, err := simSvc.CreateSimulation(ctx, services.CreateSimulationRequest{
		Mode:         "simple",
		RecipeName:   "test_preset",
		Recipe:       recipe,
		StartContest: 1001,
		EndContest:   1002, // Smaller range for faster execution
		Async:        true, // Asynchronous execution
		CreatedBy:    "test_user",
	})
	if err != nil {
		t.Fatalf("Failed to create async simulation: %v", err)
	}

	// Verify simulation is pending
	if sim.Status != "pending" {
		t.Errorf("Expected simulation status 'pending', got '%s'", sim.Status)
	}

	// Create worker and process the job
	cfg := &config.WorkerConfig{
		Concurrency:  1,
		PollInterval: 1 * time.Second,
	}

	jobWorker := worker.NewJobWorker(
		simulations.New(db.SimulationsDB),
		simSvc,
		"test-worker",
		cfg.PollInterval,
		cfg.Concurrency,
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})),
	)

	// Start worker in background
	done := make(chan bool)
	go func() {
		defer close(done)
		jobWorker.Start(ctx)
	}()

	// Wait for job to be processed (with timeout)
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

jobLoop:
	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for job completion")
		case <-ticker.C:
			// Check simulation status
			updatedSim, err := simSvc.GetSimulation(ctx, sim.ID)
			if err != nil {
				t.Fatalf("Failed to get simulation: %v", err)
			}

			if updatedSim.Status == "completed" {
				break jobLoop
			} else if updatedSim.Status == "failed" {
				t.Fatalf("Simulation failed: %s", updatedSim.ErrorMessage.String)
			}
		}
	}

	// Stop worker
	jobWorker.Stop()

	// Wait for worker to stop
	select {
	case <-done:
		// Worker stopped
	case <-time.After(5 * time.Second):
		t.Error("Worker did not stop gracefully")
	}

	// Verify final simulation state
	finalSim, err := simSvc.GetSimulation(ctx, sim.ID)
	if err != nil {
		t.Fatalf("Failed to get final simulation: %v", err)
	}

	if finalSim.Status != "completed" {
		t.Errorf("Expected final status 'completed', got '%s'", finalSim.Status)
	}

	if !finalSim.RunDurationMs.Valid || finalSim.RunDurationMs.Int64 == 0 {
		t.Error("Expected non-zero run duration")
	}

	// Verify worker ID was set
	if finalSim.WorkerID.String == "" {
		t.Error("Expected worker ID to be set")
	}
}

func TestConfigManagement(t *testing.T) {
	db, dbCleanup := setupTestDB(t)
	defer dbCleanup()

	_, configSvc, _, svcCleanup := setupServices(t, db)
	defer svcCleanup()

	ctx := context.Background()

	// Test preset loading
	preset, err := configSvc.GetPreset(ctx, "test_preset")
	if err != nil {
		t.Fatalf("Failed to load preset: %v", err)
	}

	if preset.Name != "test_preset" {
		t.Errorf("Expected preset name 'test_preset', got '%s'", preset.Name)
	}

	// Test listing presets
	presets, err := configSvc.ListPresets(ctx)
	if err != nil {
		t.Fatalf("Failed to list presets: %v", err)
	}

	if len(presets) == 0 {
		t.Error("Expected at least one preset")
	}

	// Test creating a custom config
	testRecipe := services.Recipe{
		Version: "1.0",
		Name:    "Custom Test",
		Parameters: services.RecipeParameters{
			Alpha:      0.5,
			Beta:       0.3,
			Gamma:      0.2,
			Delta:      0.1,
			SimPrevMax: 20,
			SimPreds:   10,
		},
	}

	config, err := configSvc.Create(ctx, services.CreateConfigRequest{
		Name:        "integration_test_config",
		Description: "Config created during integration test",
		Recipe:      testRecipe,
		Mode:        "advanced",
		CreatedBy:   "test_user",
	})
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	configID := config.ID

	if config.Name != "integration_test_config" {
		t.Errorf("Expected config name 'integration_test_config', got '%s'", config.Name)
	}

	// Test updating config
	updatedConfig := services.CreateConfigRequest{
		Name:        config.Name,
		Description: "Updated description",
		Recipe:      testRecipe,
		Mode:        config.Mode,
		CreatedBy:   "test_user",
	}
	err = configSvc.Update(ctx, configID, updatedConfig)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Verify update
	updated, err := configSvc.Get(ctx, configID)
	if err != nil {
		t.Fatalf("Failed to get updated config: %v", err)
	}

	if !updated.Description.Valid || updated.Description.String != "Updated description" {
		t.Errorf("Expected updated description, got '%s'", updated.Description.String)
	}

	// Test listing configs
	configs, err := configSvc.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list configs: %v", err)
	}

	if len(configs) == 0 {
		t.Error("Expected at least one config")
	}

	// Test setting as default
	err = configSvc.SetDefault(ctx, configID)
	if err != nil {
		t.Fatalf("Failed to set default config: %v", err)
	}

	// Test getting default config
	defaultConfig, err := configSvc.GetDefault(ctx, "advanced")
	if err != nil {
		t.Fatalf("Failed to get default config: %v", err)
	}

	if defaultConfig.ID != configID {
		t.Errorf("Expected default config ID %d, got %d", configID, defaultConfig.ID)
	}
}
