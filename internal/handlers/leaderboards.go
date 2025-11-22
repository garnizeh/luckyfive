package handlers

import (
	"net/http"
	"strconv"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/go-chi/chi/v5"
)

// GetLeaderboard retrieves leaderboard for a specific metric
// @Summary Get leaderboard by metric
// @Description Get ranked list of simulations for a specific metric with optional filtering
// @Tags leaderboards
// @Accept json
// @Produce json
// @Param metric path string true "Metric name (quina_rate, quadra_rate, terno_rate, avg_hits, total_quinaz, total_quadras, total_ternos, hit_efficiency)"
// @Param mode query string false "Filter by simulation mode" Enums(simple,advanced,sweep,all) default(all)
// @Param date_from query string false "Filter simulations created after this date (RFC3339 format)"
// @Param date_to query string false "Filter simulations created before this date (RFC3339 format)"
// @Param limit query int false "Maximum number of results" default(50) maximum(1000)
// @Param offset query int false "Number of results to skip" default(0)
// @Success 200 {object} object{leaderboard=[]models.LeaderboardEntryResponse,total=integer,limit=integer,offset=integer} "Leaderboard results"
// @Failure 400 {object} models.APIError "Invalid parameters"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/leaderboards/{metric} [get]
func GetLeaderboard(leaderboardSvc services.LeaderboardServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get metric from URL path
		metric := chi.URLParam(r, "metric")
		if metric == "" {
			WriteError(w, r, *models.NewAPIError("invalid_metric", "Metric is required"))
			return
		}

		// Parse query parameters with defaults
		mode := r.URL.Query().Get("mode")
		if mode == "" {
			mode = "all"
		}
		dateFrom := r.URL.Query().Get("date_from")
		dateTo := r.URL.Query().Get("date_to")

		limit := 50 // default
		offset := 0 // default

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
				if limit > 1000 {
					limit = 1000
				}
			}
		}

		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		// Create request
		req := services.LeaderboardRequest{
			Metric:   metric,
			Mode:     mode,
			DateFrom: dateFrom,
			DateTo:   dateTo,
			Limit:    limit,
			Offset:   offset,
		}

		// Get leaderboard
		leaderboard, err := leaderboardSvc.GetLeaderboard(r.Context(), req)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("get_leaderboard_failed", err.Error()))
			return
		}

		response := convertLeaderboardEntriesToResponse(leaderboard)
		WriteJSON(w, http.StatusOK, map[string]any{
			"leaderboard": response,
			"total":       len(response),
			"limit":       limit,
			"offset":      offset,
		})
	}
}

// convertLeaderboardEntriesToResponse converts internal LeaderboardEntry slice to API response
func convertLeaderboardEntriesToResponse(entries []services.LeaderboardEntry) []models.LeaderboardEntryResponse {
	response := make([]models.LeaderboardEntryResponse, len(entries))
	for i, entry := range entries {
		response[i] = models.LeaderboardEntryResponse{
			Rank:         entry.Rank,
			SimulationID: entry.SimulationID,
			RecipeName:   entry.RecipeName,
			MetricValue:  entry.MetricValue,
			CreatedAt:    entry.CreatedAt,
		}
	}
	return response
}
