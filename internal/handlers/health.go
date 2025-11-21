package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/garnizeh/luckyfive/internal/store"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Services  map[string]string `json:"services"`
}

// HealthCheck returns an HTTP handler for health checks
func HealthCheck(db *store.DB, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		services := make(map[string]string)

		// Check database connectivity
		if err := checkDatabaseHealth(db); err != nil {
			services["database"] = "unhealthy"
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			services["database"] = "healthy"
		}

		// Check other services here (e.g., external APIs, caches, etc.)
		services["api"] = "healthy"

		response := HealthResponse{
			Status:    getOverallStatus(services),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   "1.0.0", // TODO: Get from build info or config
			Uptime:    time.Since(startTime).String(),
			Services:  services,
		}

		w.Header().Set("Content-Type", "application/json")
		if response.Status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(response)
	}
}

// checkDatabaseHealth verifies connectivity to all databases
func checkDatabaseHealth(db *store.DB) error {
	// Test Results database
	if err := db.ResultsDB.Ping(); err != nil {
		return err
	}

	// Test Simulations database
	if err := db.SimulationsDB.Ping(); err != nil {
		return err
	}

	// Test Configs database
	if err := db.ConfigsDB.Ping(); err != nil {
		return err
	}

	// Test Finances database
	if err := db.FinancesDB.Ping(); err != nil {
		return err
	}

	return nil
}

// getOverallStatus determines the overall health status
func getOverallStatus(services map[string]string) string {
	for _, status := range services {
		if status == "unhealthy" {
			return "unhealthy"
		}
	}
	return "healthy"
}
