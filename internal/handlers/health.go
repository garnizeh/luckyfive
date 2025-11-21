package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
)

// SystemServiceInterface defines the interface for system service operations
type SystemServiceInterface interface {
	CheckHealth() (*services.HealthStatus, error)
}

// HealthCheck returns an HTTP handler for health checks
func HealthCheck(systemSvc SystemServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health, err := systemSvc.CheckHealth()
		if err != nil {
			WriteError(w, r, *models.NewAPIError("health_check_failed", "Health check failed"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if health.Status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(health)
	}
}
