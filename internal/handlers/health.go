package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/garnizeh/luckyfive/internal/services"
)

// HealthCheck returns an HTTP handler for health checks
func HealthCheck(systemSvc *services.SystemService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health, err := systemSvc.CheckHealth()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, APIError{
				Code:    "health_check_failed",
				Message: "Health check failed",
			})
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
