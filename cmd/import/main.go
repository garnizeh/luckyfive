package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/garnizeh/luckyfive/internal/config"
	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/garnizeh/luckyfive/internal/store"
)

func main() {
	envFile := flag.String("env-file", ".env", "Path to .env configuration file")
	xlsxPath := flag.String("xlsx", "data/results/Quina.xlsx", "path to XLSX file")
	sheetName := flag.String("sheet", "QUINA", "sheet name in XLSX file")
	flag.Parse()

	// Load configuration from .env (if provided)
	cfg, err := config.Load(*envFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Open database (use path from .env)
	db, err := store.Open(store.Config{
		ResultsPath:     cfg.Database.ResultsPath,
		SimulationsPath: ":memory:", // Not needed for import
		ConfigsPath:     ":memory:",
		FinancesPath:    ":memory:",
	})
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create import service
	importSvc := services.NewImportService(db, logger)

	// Open XLSX file
	file, err := os.Open(*xlsxPath)
	if err != nil {
		log.Fatalf("Failed to open XLSX file: %v", err)
	}
	defer file.Close()

	// Parse XLSX
	logger.Info("Parsing XLSX file", "path", *xlsxPath, "sheet", *sheetName)
	draws, err := importSvc.ParseXLSX(file, *sheetName)
	if err != nil {
		log.Fatalf("Failed to parse XLSX: %v", err)
	}

	logger.Info("Parsed draws", "count", len(draws))

	// Import draws
	logger.Info("Importing draws to database")
	err = importSvc.ImportDraws(context.Background(), draws)
	if err != nil {
		log.Fatalf("Failed to import draws: %v", err)
	}

	logger.Info("Import completed successfully")
}
