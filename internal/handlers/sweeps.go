package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/go-chi/chi/v5"
)

// CreateSweep creates a new sweep job
// @Summary Create a new sweep job
// @Description Create a parameter sweep job that will generate and execute multiple simulations
// @Tags sweeps
// @Accept json
// @Produce json
// @Param request body services.CreateSweepRequest true "Sweep creation request"
// @Success 202 {object} object{id=integer,name=string,description=string,sweep_config_json=string,base_contest_range=string,status=string,total_combinations=integer,completed_simulations=integer,failed_simulations=integer,created_at=string,started_at=string,finished_at=string,run_duration_ms=integer,created_by=string} "Sweep job created successfully"
// @Failure 400 {object} models.APIError "Invalid request"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweeps [post]
func CreateSweep(sweepSvc services.SweepServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req services.CreateSweepRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid JSON in request body"))
			return
		}

		sweep, err := sweepSvc.CreateSweep(r.Context(), req)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("create_sweep_failed", err.Error()))
			return
		}

		WriteJSON(w, http.StatusAccepted, sweep)
	}
}

// GetSweep retrieves a sweep job by ID
// @Summary Get sweep job by ID
// @Description Retrieve details of a specific sweep job
// @Tags sweeps
// @Accept json
// @Produce json
// @Param id path int true "Sweep job ID"
// @Success 200 {object} object{id=integer,name=string,description=string,sweep_config_json=string,base_contest_range=string,status=string,total_combinations=integer,completed_simulations=integer,failed_simulations=integer,created_at=string,started_at=string,finished_at=string,run_duration_ms=integer,created_by=string} "Sweep job details"
// @Failure 400 {object} models.APIError "Invalid ID"
// @Failure 404 {object} models.APIError "Sweep job not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweeps/{id} [get]
func GetSweep(sweepSvc services.SweepServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Sweep ID is required"))
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Invalid sweep ID"))
			return
		}

		// For now, we'll get the status which includes the sweep job details
		status, err := sweepSvc.GetSweepStatus(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("get_sweep_failed", err.Error()))
			return
		}

		WriteJSON(w, http.StatusOK, status.Sweep)
	}
}

// GetSweepStatus retrieves the status of a sweep job
// @Summary Get sweep job status
// @Description Retrieve the current status and progress of a sweep job
// @Tags sweeps
// @Accept json
// @Produce json
// @Param id path int true "Sweep job ID"
// @Success 200 {object} models.SweepStatusResponse "Sweep job status"
// @Failure 400 {object} models.APIError "Invalid ID"
// @Failure 404 {object} models.APIError "Sweep job not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweeps/{id}/status [get]
func GetSweepStatus(sweepSvc services.SweepServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Sweep ID is required"))
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Invalid sweep ID"))
			return
		}

		status, err := sweepSvc.GetSweepStatus(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("get_sweep_status_failed", err.Error()))
			return
		}

		response := convertSweepStatusToResponse(status)
		WriteJSON(w, http.StatusOK, response)
	}
}

// GetSweepResults retrieves the results of a completed sweep job
// @Summary Get sweep job results
// @Description Retrieve the results and simulation details of a completed sweep job
// @Tags sweeps
// @Accept json
// @Produce json
// @Param id path int true "Sweep job ID"
// @Success 200 {object} models.SweepStatusResponse "Sweep job results"
// @Failure 400 {object} models.APIError "Invalid ID"
// @Failure 404 {object} models.APIError "Sweep job not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweeps/{id}/results [get]
func GetSweepResults(sweepSvc services.SweepServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Sweep ID is required"))
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Invalid sweep ID"))
			return
		}

		// Get status which includes all simulation results
		status, err := sweepSvc.GetSweepStatus(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("get_sweep_results_failed", err.Error()))
			return
		}

		response := convertSweepStatusToResponse(status)
		WriteJSON(w, http.StatusOK, response)
	}
}

// CancelSweep cancels a running sweep job
// @Summary Cancel sweep job
// @Description Cancel a running sweep job and stop all pending simulations
// @Tags sweeps
// @Accept json
// @Produce json
// @Param id path int true "Sweep job ID"
// @Success 200 {object} object{message=string} "Sweep job cancelled successfully"
// @Failure 400 {object} models.APIError "Invalid ID"
// @Failure 404 {object} models.APIError "Sweep job not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweeps/{id}/cancel [post]
func CancelSweep(sweepSvc services.SweepServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Sweep ID is required"))
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Invalid sweep ID"))
			return
		}

		// For now, we'll just update the progress to mark as cancelled
		// In a real implementation, this would cancel running simulations
		err = sweepSvc.UpdateSweepProgress(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("cancel_sweep_failed", err.Error()))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{
			"message": "Sweep job cancellation requested",
		})
	}
}

// GetSweepBest finds the best configuration from a completed sweep
// @Summary Get best configuration from sweep
// @Description Find the best performing configuration from a completed sweep job based on a specified metric
// @Tags sweeps
// @Accept json
// @Produce json
// @Param id path int true "Sweep job ID"
// @Param metric query string true "Optimization metric (quina_rate, quadra_rate, terno_rate, avg_hits, total_quinaz, total_quadras, total_ternos, hit_efficiency)"
// @Success 200 {object} models.BestConfigurationResponse "Best configuration found"
// @Failure 400 {object} models.APIError "Invalid ID or metric"
// @Failure 404 {object} models.APIError "Sweep job not found or not completed"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweeps/{id}/best [get]
func GetSweepBest(sweepSvc services.SweepServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Sweep ID is required"))
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Invalid sweep ID"))
			return
		}

		metric := r.URL.Query().Get("metric")
		if metric == "" {
			WriteError(w, r, *models.NewAPIError("invalid_metric", "Metric parameter is required"))
			return
		}

		best, err := sweepSvc.FindBest(r.Context(), id, metric)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("find_best_failed", err.Error()))
			return
		}

		response := convertBestConfigurationToResponse(best)
		WriteJSON(w, http.StatusOK, response)
	}
}

// GetSweepVisualization returns data suitable for visualization from a sweep
// @Summary Get sweep visualization data
// @Description Export sweep data formatted for heatmaps, scatter plots, and other visualizations
// @Tags sweeps
// @Accept json
// @Produce json
// @Param id path int true "Sweep job ID"
// @Param metrics query []string false "Metrics to include (comma-separated). Defaults to quina_rate,avg_hits"
// @Success 200 {object} models.VisualizationDataResponse "Visualization data"
// @Failure 400 {object} models.APIError "Invalid ID or metrics"
// @Failure 404 {object} models.APIError "Sweep job not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweeps/{id}/visualization [get]
func GetSweepVisualization(sweepSvc services.SweepServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Sweep ID is required"))
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_id", "Invalid sweep ID"))
			return
		}

		// Parse metrics query parameter (comma-separated)
		metricsStr := r.URL.Query().Get("metrics")
		var metrics []string
		if metricsStr != "" {
			metrics = strings.Split(metricsStr, ",")
			// Trim spaces
			for i, m := range metrics {
				metrics[i] = strings.TrimSpace(m)
			}
		}

		data, err := sweepSvc.GetVisualizationData(r.Context(), id, metrics)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("get_visualization_failed", err.Error()))
			return
		}

		response := convertVisualizationDataToResponse(data)
		WriteJSON(w, http.StatusOK, response)
	}
}

// convertSweepStatusToResponse converts internal SweepStatus to API response
func convertSweepStatusToResponse(status *services.SweepStatus) *models.SweepStatusResponse {
	response := &models.SweepStatusResponse{
		Sweep: models.SweepJobResponse{
			ID:                status.Sweep.ID,
			Name:              status.Sweep.Name,
			SweepConfigJson:   status.Sweep.SweepConfigJson,
			BaseContestRange:  status.Sweep.BaseContestRange,
			Status:            status.Sweep.Status,
			TotalCombinations: status.Sweep.TotalCombinations,
		},
		Total:     status.Total,
		Completed: status.Completed,
		Running:   status.Running,
		Failed:    status.Failed,
		Pending:   status.Pending,
	}

	// Handle nullable fields
	if status.Sweep.Description.Valid {
		response.Sweep.Description = status.Sweep.Description.String
	}
	if status.Sweep.CompletedSimulations.Valid {
		response.Sweep.CompletedSimulations = status.Sweep.CompletedSimulations.Int64
	}
	if status.Sweep.FailedSimulations.Valid {
		response.Sweep.FailedSimulations = status.Sweep.FailedSimulations.Int64
	}
	if status.Sweep.CreatedAt.Valid {
		response.Sweep.CreatedAt = status.Sweep.CreatedAt.String
	}
	if status.Sweep.StartedAt.Valid {
		response.Sweep.StartedAt = status.Sweep.StartedAt.String
	}
	if status.Sweep.FinishedAt.Valid {
		response.Sweep.FinishedAt = status.Sweep.FinishedAt.String
	}
	if status.Sweep.RunDurationMs.Valid {
		response.Sweep.RunDurationMs = status.Sweep.RunDurationMs.Int64
	}
	if status.Sweep.CreatedBy.Valid {
		response.Sweep.CreatedBy = status.Sweep.CreatedBy.String
	}

	// Convert simulations
	response.Simulations = make([]models.SweepSimulationDetailResponse, len(status.Simulations))
	for i, sim := range status.Simulations {
		response.Simulations[i] = models.SweepSimulationDetailResponse{
			ID:              sim.ID,
			SweepJobID:      sim.SweepJobID,
			SimulationID:    sim.SimulationID,
			VariationIndex:  sim.VariationIndex,
			VariationParams: sim.VariationParams,
			Status:          sim.Status,
		}

		// Handle nullable fields
		if sim.SummaryJson.Valid {
			response.Simulations[i].SummaryJson = sim.SummaryJson.String
		}
		if sim.RunDurationMs.Valid {
			response.Simulations[i].RunDurationMs = sim.RunDurationMs.Int64
		}
	}

	return response
}

// convertBestConfigurationToResponse converts internal BestConfiguration to API response
func convertBestConfigurationToResponse(best *services.BestConfiguration) *models.BestConfigurationResponse {
	return &models.BestConfigurationResponse{
		SweepID:      best.SweepID,
		SimulationID: best.SimulationID,
		Recipe: models.RecipeResponse{
			Version:    best.Recipe.Version,
			Name:       best.Recipe.Name,
			Parameters: best.Recipe.Parameters,
		},
		Metrics:         best.Metrics,
		Rank:            best.Rank,
		Percentile:      best.Percentile,
		VariationParams: best.VariationParams,
	}
}

// convertVisualizationDataToResponse converts internal VisualizationData to API response
func convertVisualizationDataToResponse(data *services.VisualizationData) *models.VisualizationDataResponse {
	response := &models.VisualizationDataResponse{
		SweepID:    data.SweepID,
		Parameters: data.Parameters,
		Metrics:    data.Metrics,
		DataPoints: make([]models.VisualizationDataPointResponse, len(data.DataPoints)),
	}

	for i, point := range data.DataPoints {
		response.DataPoints[i] = models.VisualizationDataPointResponse{
			Params:  point.Params,
			Metrics: point.Metrics,
		}
	}

	return response
}
