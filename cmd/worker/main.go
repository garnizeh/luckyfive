package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/garnizeh/luckyfive/internal/config"
	"github.com/garnizeh/luckyfive/internal/logger"
	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/garnizeh/luckyfive/internal/store"
	"github.com/garnizeh/luckyfive/internal/worker"
	"github.com/google/uuid"
)

func main() {
	// Parse command line flags
	envFile := flag.String("env-file", ".env", "Path to .env configuration file")
	workerID := flag.String("worker-id", "", "Unique identifier for this worker instance (auto-generated if not provided)")
	flag.Parse()

	// Generate worker ID if not provided
	if *workerID == "" {
		*workerID = "worker-" + uuid.New().String()
	}

	// Load configuration
	cfg, err := config.Load(*envFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.LogLevel)

	// Initialize database connections
	db, err := store.Open(store.Config{
		ResultsPath:     cfg.Database.ResultsPath,
		SimulationsPath: cfg.Database.SimulationsPath,
		ConfigsPath:     cfg.Database.ConfigsPath,
		FinancesPath:    cfg.Database.FinancesPath,
	})
	if err != nil {
		logger.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize services
	engineSvc := services.NewEngineService(db.Results, logger)
	simSvc := services.NewSimulationService(db.Simulations, db.SimulationsDB, engineSvc, logger)

	// Create worker
	jobWorker := worker.NewJobWorker(
		db.Simulations,
		simSvc,
		*workerID,
		cfg.Worker.PollInterval,
		cfg.Worker.Concurrency,
		logger,
	)

	// Handle shutdown signals
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("Starting worker", "worker_id", *workerID, "concurrency", cfg.Worker.Concurrency)

	if err := jobWorker.Start(ctx); err != nil {
		logger.Error("Worker error", "error", err)
		os.Exit(1)
	}

	logger.Info("Worker stopped", "worker_id", *workerID)
}
