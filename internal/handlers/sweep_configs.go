package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/garnizeh/luckyfive/pkg/sweep"
)

type CreateSweepConfigRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=100"`
	Description string            `json:"description" validate:"max=500"`
	Config      sweep.SweepConfig `json:"config" validate:"required"`
}

type UpdateSweepConfigRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=100"`
	Description string            `json:"description" validate:"max=500"`
	Config      sweep.SweepConfig `json:"config" validate:"required"`
}

// ListSweepConfigs godoc
// @Summary List sweep configurations
// @Description Retrieve a paginated list of sweep configurations
// @Tags sweep-configs
// @Accept json
// @Produce json
// @Param limit query integer false "Maximum number of sweep configs to return" default(50)
// @Param offset query integer false "Number of sweep configs to skip" default(0)
// @Success 200 {object} object{sweep_configs=[]object{id=integer,name=string,description=string,config_json=string,created_at=string,updated_at=string,created_by=string,times_used=integer,last_used_at=string},pagination=object{limit=integer,offset=integer}} "List of sweep configurations"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweep-configs [get]
func ListSweepConfigs(sweepSvc services.SweepConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		limit := int64(50) // default limit
		if limitStr != "" {
			if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		offset := int64(0) // default offset
		if offsetStr != "" {
			if o, err := strconv.ParseInt(offsetStr, 10, 64); err == nil && o >= 0 {
				offset = o
			}
		}

		// Get sweep configs
		sweepConfigs, err := sweepSvc.List(r.Context(), limit, offset)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("list_sweep_configs_failed", "Failed to list sweep configs"))
			return
		}

		// Return response
		response := map[string]interface{}{
			"sweep_configs": sweepConfigs,
			"pagination": map[string]interface{}{
				"limit":  limit,
				"offset": offset,
			},
		}

		WriteJSON(w, http.StatusOK, response)
	}
}

// CreateSweepConfig godoc
// @Summary Create a sweep configuration
// @Description Create a new sweep configuration
// @Tags sweep-configs
// @Accept json
// @Produce json
// @Param request body handlers.CreateSweepConfigRequest true "Sweep configuration data"
// @Success 201 {object} object{id=integer,name=string,description=string,config_json=string,created_at=string,updated_at=string,created_by=string,times_used=integer,last_used_at=string} "Sweep configuration created"
// @Failure 400 {object} models.APIError "Invalid request"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweep-configs [post]
func CreateSweepConfig(sweepSvc services.SweepConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateSweepConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid JSON"))
			return
		}

		// Basic validation
		if req.Name == "" {
			WriteError(w, r, *models.NewAPIError("validation_error", "Name is required"))
			return
		}
		if req.Config.Name == "" {
			WriteError(w, r, *models.NewAPIError("validation_error", "Config name is required"))
			return
		}

		// Create sweep config
		sweepConfig, err := sweepSvc.Create(r.Context(), services.CreateSweepConfigRequest{
			Name:        req.Name,
			Description: req.Description,
			Config:      req.Config,
			CreatedBy:   "api", // TODO: get from auth context
		})
		if err != nil {
			WriteError(w, r, *models.NewAPIError("sweep_config_creation_failed", "Failed to create sweep config"))
			return
		}

		WriteJSON(w, http.StatusCreated, sweepConfig)
	}
}

// GetSweepConfig godoc
// @Summary Get sweep configuration details
// @Description Retrieve details of a specific sweep configuration by ID
// @Tags sweep-configs
// @Accept json
// @Produce json
// @Param id path integer true "Sweep configuration ID"
// @Success 200 {object} object{id=integer,name=string,description=string,config_json=string,created_at=string,updated_at=string,created_by=string,times_used=integer,last_used_at=string} "Sweep configuration details"
// @Failure 400 {object} models.APIError "Invalid sweep configuration ID"
// @Failure 404 {object} models.APIError "Sweep configuration not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweep-configs/{id} [get]
func GetSweepConfig(sweepSvc services.SweepConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_config_id", "Invalid sweep config ID"))
			return
		}

		// Get sweep config
		sweepConfig, err := sweepSvc.Get(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("sweep_config_not_found", "Sweep config not found"))
			return
		}

		WriteJSON(w, http.StatusOK, sweepConfig)
	}
}

// UpdateSweepConfig godoc
// @Summary Update a sweep configuration
// @Description Update an existing sweep configuration
// @Tags sweep-configs
// @Accept json
// @Produce json
// @Param id path integer true "Sweep configuration ID"
// @Param request body handlers.UpdateSweepConfigRequest true "Updated sweep configuration data"
// @Success 200 {object} object{message=string} "Sweep configuration updated"
// @Failure 400 {object} models.APIError "Invalid request"
// @Failure 404 {object} models.APIError "Sweep configuration not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweep-configs/{id} [put]
func UpdateSweepConfig(sweepSvc services.SweepConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_config_id", "Invalid sweep config ID"))
			return
		}

		var req UpdateSweepConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid JSON"))
			return
		}

		// Basic validation
		if req.Name == "" {
			WriteError(w, r, *models.NewAPIError("validation_error", "Name is required"))
			return
		}
		if req.Config.Name == "" {
			WriteError(w, r, *models.NewAPIError("validation_error", "Config name is required"))
			return
		}

		// Update sweep config
		err = sweepSvc.Update(r.Context(), id, services.CreateSweepConfigRequest{
			Name:        req.Name,
			Description: req.Description,
			Config:      req.Config,
			CreatedBy:   "api", // TODO: get from auth context
		})
		if err != nil {
			WriteError(w, r, *models.NewAPIError("sweep_config_update_failed", "Failed to update sweep config"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "Sweep config updated successfully"})
	}
}

// DeleteSweepConfig godoc
// @Summary Delete a sweep configuration
// @Description Delete a sweep configuration
// @Tags sweep-configs
// @Accept json
// @Produce json
// @Param id path integer true "Sweep configuration ID"
// @Success 200 {object} object{message=string} "Sweep configuration deleted"
// @Failure 400 {object} models.APIError "Invalid sweep configuration ID"
// @Failure 404 {object} models.APIError "Sweep configuration not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/sweep-configs/{id} [delete]
func DeleteSweepConfig(sweepSvc services.SweepConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_sweep_config_id", "Invalid sweep config ID"))
			return
		}

		// Delete sweep config
		err = sweepSvc.Delete(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("sweep_config_delete_failed", "Failed to delete sweep config"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "Sweep config deleted successfully"})
	}
}
