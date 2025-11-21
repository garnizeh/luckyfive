package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/garnizeh/luckyfive/internal/config"
	"github.com/garnizeh/luckyfive/internal/handlers"
	"github.com/garnizeh/luckyfive/internal/logger"
	"github.com/garnizeh/luckyfive/internal/middleware"
	"github.com/garnizeh/luckyfive/internal/store"
)

func main() {
	// Record start time for uptime tracking
	startTime := time.Now()

	// Parse command line flags
	envFile := flag.String("env-file", ".env", "Path to .env configuration file")
	flag.Parse()

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

	// Setup router
	router := setupRouter(logger, db, startTime)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited")
}

func setupRouter(logger *slog.Logger, db *store.DB, startTime time.Time) *chi.Mux {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(chimiddleware.RequestID)     // Request ID tracking
	r.Use(chimiddleware.RealIP)        // Real IP detection
	r.Use(middleware.Logging(logger))  // Custom logging
	r.Use(middleware.Recovery(logger)) // Panic recovery
	r.Use(middleware.CORS())           // CORS headers

	// Health check endpoint
	r.Get("/api/v1/health", handlers.HealthCheck(db, startTime))

	return r
}
