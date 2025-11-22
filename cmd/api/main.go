// @title LuckyFive API
// @version 1.0
// @description Quina Lottery Simulation API — upload XLSX results, import into SQLite, and query draws.
// @termsOfService https://example.com/terms/
// @contact.name LuckyFive API Support
// @contact.url https://example.com/support
// @contact.email devops@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @schemes http

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

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/garnizeh/luckyfive/internal/config"
	"github.com/garnizeh/luckyfive/internal/handlers"
	"github.com/garnizeh/luckyfive/internal/logger"
	"github.com/garnizeh/luckyfive/internal/middleware"
	"github.com/garnizeh/luckyfive/internal/services"
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

	// Initialize services
	systemSvc := services.NewSystemService(db, startTime)
	uploadSvc := services.NewUploadService(logger)
	resultsSvc := services.NewResultsService(db, logger)
	engineSvc := services.NewEngineService(db.Results, logger)
	configSvc := services.NewConfigService(db.Configs, db.ConfigsDB, logger)
	simSvc := services.NewSimulationService(db.Simulations, db.SimulationsDB, engineSvc, logger)

	// Setup router
	router := setupRouter(logger, systemSvc, uploadSvc, resultsSvc, configSvc, simSvc)

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

func setupRouter(logger *slog.Logger, systemSvc *services.SystemService, uploadSvc *services.UploadService, resultsSvc *services.ResultsService, configSvc *services.ConfigService, simSvc *services.SimulationService) *chi.Mux {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(chimiddleware.RequestID)     // Request ID tracking
	r.Use(chimiddleware.RealIP)        // Real IP detection
	r.Use(middleware.Logging(logger))  // Custom logging
	r.Use(middleware.Recovery(logger)) // Panic recovery
	r.Use(middleware.CORS())           // CORS headers

	// Health check endpoint
	r.Get("/api/v1/health", handlers.HealthCheck(systemSvc))

	// Results upload endpoint
	r.Post("/api/v1/results/upload", handlers.UploadResults(uploadSvc, logger))

	// Results import endpoint
	r.Post("/api/v1/results/import", handlers.ImportResults(resultsSvc, logger))

	// Results query endpoints
	r.Get("/api/v1/results/{contest}", handlers.GetDraw(resultsSvc, logger))
	r.Get("/api/v1/results", handlers.ListDraws(resultsSvc, logger))

	// Simulation endpoints
	r.Post("/api/v1/simulations/simple", handlers.SimpleSimulation(configSvc, simSvc))
	r.Get("/api/v1/simulations/{id}", handlers.GetSimulation(simSvc))
	r.Get("/api/v1/simulations", handlers.ListSimulations(simSvc))
	r.Post("/api/v1/simulations/{id}/cancel", handlers.CancelSimulation(simSvc))

	// Config endpoints
	r.Get("/api/v1/configs", handlers.ListConfigs(configSvc))
	r.Post("/api/v1/configs", handlers.CreateConfig(configSvc))
	r.Get("/api/v1/configs/{id}", handlers.GetConfig(configSvc))
	r.Put("/api/v1/configs/{id}", handlers.UpdateConfig(configSvc))
	r.Delete("/api/v1/configs/{id}", handlers.DeleteConfig(configSvc))
	r.Post("/api/v1/configs/{id}/set-default", handlers.SetDefaultConfig(configSvc))

	// Swagger UI — serves UI and expects swagger JSON at /swagger/doc.json
	// If you generate docs with `swag init -g cmd/api/main.go -o api`,
	// the generated swagger.json will be placed under ./api and served here.
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "api/swagger.json")
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	return r
}
