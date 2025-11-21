package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/config"
	"github.com/garnizeh/luckyfive/pkg/migrator"
)

func main() {
	envFile := flag.String("env-file", ".env", "Path to .env configuration file")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: migrate [up|down] [--env-file=path]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	// Load configuration
	cfg, err := config.Load(*envFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	action := strings.ToLower(flag.Arg(0))
	dbs := []string{cfg.Database.ResultsPath, cfg.Database.SimulationsPath, cfg.Database.ConfigsPath, cfg.Database.FinancesPath}
	logger := slog.Default()

	for _, dbPath := range dbs {
		dbPath = strings.TrimSpace(dbPath)
		if dbPath == "" {
			continue
		}
		// Ensure parent dir exists for file dbs
		if !strings.HasPrefix(dbPath, ":") {
			parent := filepath.Dir(dbPath)
			if parent == "." || parent == "" {
				parent = "."
			}
			// create parent dir recursively (handles data/ and data/db/)
			if err := os.MkdirAll(parent, 0o755); err != nil {
				logger.Error("create parent dir failed", "dir", parent, "err", err)
				os.Exit(1)
			}
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			logger.Error("open db", "db", dbPath, "err", err)
			os.Exit(1)
		}
		// Choose migrations directory for this DB. If a subdirectory exists under
		// migrations/ matching the DB filename (without extension), use it.
		migrationsDir := "migrations"
		if dbPath != ":memory:" {
			base := filepath.Base(dbPath)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			candidate := filepath.Join(migrationsDir, name)
			if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
				migrationsDir = candidate
				logger.Info("using db-specific migrations directory", "db", dbPath, "dir", candidate)
			}
		}
		migrator := migrator.New(db, migrationsDir, logger)
		switch action {
		case "up":
			logger.Info("migrating up", "db", dbPath)
			if err := migrator.Up(); err != nil {
				logger.Error("migration up failed", "db", dbPath, "err", err)
				db.Close()
				os.Exit(1)
			}
		case "down":
			logger.Info("migrating down", "db", dbPath)
			if err := migrator.Down(); err != nil {
				logger.Error("migration down failed", "db", dbPath, "err", err)
				db.Close()
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "unknown action: %s (use 'up' or 'down')\n", action)
			db.Close()
			os.Exit(2)
		}
		db.Close()
	}
}
