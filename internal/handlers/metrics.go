package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
)

// MetricsServiceInterface defines the interface for metrics service operations
type MetricsServiceInterface interface {
	GetSystemMetrics(ctx context.Context) (*services.SystemMetrics, error)
}

// GetMetrics returns an HTTP handler for retrieving system metrics
// GetMetrics godoc
// @Summary Get system performance metrics
// @Description Returns comprehensive performance metrics including database stats, query performance, and HTTP statistics
// @Tags metrics
// @Accept json
// @Produce json
// @Success 200 {object} services.SystemMetrics
// @Router /api/v1/metrics [get]
func GetMetrics(metricsSvc MetricsServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := metricsSvc.GetSystemMetrics(r.Context())
		if err != nil {
			WriteError(w, r, *models.NewAPIError("metrics_retrieval_failed", "Failed to retrieve system metrics"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			WriteError(w, r, *models.NewAPIError("metrics_encoding_failed", "Failed to encode metrics response"))
			return
		}
	}
}
