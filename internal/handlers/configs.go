package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
)

type CreateConfigRequest struct {
	Name        string          `json:"name" validate:"required,min=1,max=100"`
	Description string          `json:"description" validate:"max=500"`
	Recipe      services.Recipe `json:"recipe" validate:"required"`
	Tags        string          `json:"tags" validate:"max=200"`
	Mode        string          `json:"mode" validate:"required,oneof=simple advanced"`
}

type UpdateConfigRequest struct {
	Description string          `json:"description" validate:"max=500"`
	Recipe      services.Recipe `json:"recipe" validate:"required"`
	Tags        string          `json:"tags" validate:"max=200"`
}

// ListConfigs godoc
// @Summary List configurations
// @Description Retrieve a paginated list of simulation configurations
// @Tags configs
// @Accept json
// @Produce json
// @Param limit query integer false "Maximum number of configs to return" default(50)
// @Param offset query integer false "Number of configs to skip" default(0)
// @Success 200 {object} object{configs=[]object{id=integer,name=string,description=string,recipe_json=string,tags=string,is_default=integer,mode=string,created_at=string,updated_at=string,created_by=string,times_used=integer,last_used_at=string},pagination=object{limit=integer,offset=integer}} "List of configurations"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/configs [get]
func ListConfigs(configSvc services.ConfigServicer) http.HandlerFunc {
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

		// Get configs
		configs, err := configSvc.List(r.Context(), limit, offset)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("list_configs_failed", "Failed to list configs"))
			return
		} // Return response
		response := map[string]any{
			"configs": configs,
			"pagination": map[string]any{
				"limit":  limit,
				"offset": offset,
			},
		}

		WriteJSON(w, http.StatusOK, response)
	}
}

// CreateConfig godoc
// @Summary Create a configuration
// @Description Create a new simulation configuration
// @Tags configs
// @Accept json
// @Produce json
// @Param request body handlers.CreateConfigRequest true "Configuration data"
// @Success 201 {object} object{id=integer,name=string,description=string,recipe_json=string,tags=string,is_default=integer,mode=string,created_at=string,updated_at=string,created_by=string,times_used=integer,last_used_at=string} "Configuration created"
// @Failure 400 {object} models.APIError "Invalid request"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/configs [post]
func CreateConfig(configSvc services.ConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid JSON"))
			return
		}

		// Basic validation
		if req.Name == "" {
			WriteError(w, r, *models.NewAPIError("validation_error", "Name is required"))
			return
		}
		if req.Mode == "" {
			WriteError(w, r, *models.NewAPIError("validation_error", "Mode is required"))
			return
		}
		if req.Recipe.Version == "" {
			WriteError(w, r, *models.NewAPIError("validation_error", "Recipe version is required"))
			return
		}

		// Create config
		config, err := configSvc.Create(r.Context(), services.CreateConfigRequest{
			Name:        req.Name,
			Description: req.Description,
			Recipe:      req.Recipe,
			Tags:        req.Tags,
			Mode:        req.Mode,
			CreatedBy:   "api", // TODO: get from auth context
		})
		if err != nil {
			WriteError(w, r, *models.NewAPIError("config_creation_failed", "Failed to create config"))
			return
		}

		WriteJSON(w, http.StatusCreated, config)
	}
}

// GetConfig godoc
// @Summary Get configuration details
// @Description Retrieve details of a specific configuration by ID
// @Tags configs
// @Accept json
// @Produce json
// @Param id path integer true "Configuration ID"
// @Success 200 {object} object{id=integer,name=string,description=string,recipe_json=string,tags=string,is_default=integer,mode=string,created_at=string,updated_at=string,created_by=string,times_used=integer,last_used_at=string} "Configuration details"
// @Failure 400 {object} models.APIError "Invalid configuration ID"
// @Failure 404 {object} models.APIError "Configuration not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/configs/{id} [get]
func GetConfig(configSvc services.ConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_config_id", "Invalid config ID"))
			return
		}

		// Get config
		config, err := configSvc.Get(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("config_not_found", "Config not found"))
			return
		}

		WriteJSON(w, http.StatusOK, config)
	}
}

// UpdateConfig godoc
// @Summary Update a configuration
// @Description Update an existing simulation configuration
// @Tags configs
// @Accept json
// @Produce json
// @Param id path integer true "Configuration ID"
// @Param request body object{name=string,description=string,recipe=object{version=string,algorithm=string,parameters=object{}},tags=[]string} true "Updated configuration data"
// @Success 200 {object} object{message=string} "Configuration updated"
// @Failure 400 {object} models.APIError "Invalid request"
// @Failure 404 {object} models.APIError "Configuration not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/configs/{id} [put]
func UpdateConfig(configSvc services.ConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_config_id", "Invalid config ID"))
			return
		}

		var req UpdateConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid JSON"))
			return
		}

		// Basic validation
		if req.Recipe.Version == "" {
			WriteError(w, r, *models.NewAPIError("validation_error", "Recipe version is required"))
			return
		}

		// Update config
		err = configSvc.Update(r.Context(), id, services.CreateConfigRequest{
			Description: req.Description,
			Recipe:      req.Recipe,
			Tags:        req.Tags,
		})
		if err != nil {
			WriteError(w, r, *models.NewAPIError("config_update_failed", "Failed to update config"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "Config updated successfully"})
	}
}

// DeleteConfig godoc
// @Summary Delete a configuration
// @Description Delete a simulation configuration
// @Tags configs
// @Accept json
// @Produce json
// @Param id path integer true "Configuration ID"
// @Success 204 "Configuration deleted"
// @Failure 400 {object} models.APIError "Invalid configuration ID"
// @Failure 404 {object} models.APIError "Configuration not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/configs/{id} [delete]
func DeleteConfig(configSvc services.ConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_config_id", "Invalid config ID"))
			return
		}

		// Delete config
		err = configSvc.Delete(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("config_delete_failed", "Failed to delete config"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "Config deleted successfully"})
	}
}

// SetDefaultConfig godoc
// @Summary Set default configuration
// @Description Set a configuration as the default for its mode
// @Tags configs
// @Accept json
// @Produce json
// @Param id path integer true "Configuration ID"
// @Success 200 {object} object{message=string} "Default configuration set"
// @Failure 400 {object} models.APIError "Invalid configuration ID"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/configs/{id}/set-default [post]
func SetDefaultConfig(configSvc services.ConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_config_id", "Invalid config ID"))
			return
		}

		// Set as default
		err = configSvc.SetDefault(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("set_default_config_failed", "Failed to set default config"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "Config set as default successfully"})
	}
}
