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

// ListConfigs handles GET /api/v1/configs
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
			WriteError(w, r, *models.NewAPIError("INTERNAL_ERROR", "Failed to list configs"))
			return
		} // Return response
		response := map[string]interface{}{
			"configs": configs,
			"pagination": map[string]interface{}{
				"limit":  limit,
				"offset": offset,
			},
		}

		WriteJSON(w, http.StatusOK, response)
	}
}

// CreateConfig handles POST /api/v1/configs
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
			WriteError(w, r, *models.NewAPIError("INTERNAL_ERROR", "Failed to create config"))
			return
		}

		WriteJSON(w, http.StatusCreated, config)
	}
}

// GetConfig handles GET /api/v1/configs/{id}
func GetConfig(configSvc services.ConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_request", "Invalid config ID"))
			return
		}

		// Get config
		config, err := configSvc.Get(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("not_found", "Config not found"))
			return
		}

		WriteJSON(w, http.StatusOK, config)
	}
}

// UpdateConfig handles PUT /api/v1/configs/{id}
func UpdateConfig(configSvc services.ConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_request", "Invalid config ID"))
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
			WriteError(w, r, *models.NewAPIError("INTERNAL_ERROR", "Failed to update config"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "Config updated successfully"})
	}
}

// DeleteConfig handles DELETE /api/v1/configs/{id}
func DeleteConfig(configSvc services.ConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_request", "Invalid config ID"))
			return
		}

		// Delete config
		err = configSvc.Delete(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("INTERNAL_ERROR", "Failed to delete config"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "Config deleted successfully"})
	}
}

// SetDefaultConfig handles POST /api/v1/configs/{id}/set-default
func SetDefaultConfig(configSvc services.ConfigServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_request", "Invalid config ID"))
			return
		}

		// Set as default
		err = configSvc.SetDefault(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("INTERNAL_ERROR", "Failed to set default config"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "Config set as default successfully"})
	}
}
