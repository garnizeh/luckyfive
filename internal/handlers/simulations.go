package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
)

func SimpleSimulation(
	configSvc services.ConfigServicer,
	simSvc services.SimulationServicer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Preset       string `json:"preset"`
			StartContest int    `json:"start_contest"`
			EndContest   int    `json:"end_contest"`
			Async        bool   `json:"async"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid request format"))
			return
		}

		// Validate request
		if req.Preset == "" {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Preset is required"))
			return
		}
		if req.StartContest <= 0 || req.EndContest <= 0 || req.StartContest > req.EndContest {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid contest range"))
			return
		}

		// Load preset
		preset, err := configSvc.GetPreset(r.Context(), req.Preset)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("not_found", "Preset not found"))
			return
		}

		// Parse preset recipe
		var recipe services.Recipe
		if err := json.Unmarshal([]byte(preset.RecipeJson), &recipe); err != nil {
			WriteError(w, r, *models.NewAPIError("upload_failed", "Invalid preset configuration"))
			return
		}

		// Create simulation
		sim, err := simSvc.CreateSimulation(r.Context(), services.CreateSimulationRequest{
			Mode:         "simple",
			RecipeName:   req.Preset,
			Recipe:       recipe,
			StartContest: req.StartContest,
			EndContest:   req.EndContest,
			Async:        req.Async,
		})
		if err != nil {
			WriteError(w, r, *models.NewAPIError("upload_failed", "Simulation creation failed"))
			return
		}

		// Return response
		if req.Async {
			WriteJSON(w, http.StatusAccepted, map[string]any{
				"simulation_id": sim.ID,
				"status":        sim.Status,
				"message":       "Simulation queued for processing",
			})
		} else {
			WriteJSON(w, http.StatusOK, sim)
		}
	}
}

func GetSimulation(simSvc services.SimulationServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Simulation ID is required"))
			return
		}

		// Parse ID (assuming int64)
		var id int64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid simulation ID"))
			return
		}

		sim, err := simSvc.GetSimulation(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("not_found", "Simulation not found"))
			return
		}

		WriteJSON(w, http.StatusOK, sim)
	}
}

func ListSimulations(simSvc services.SimulationServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters for pagination
		limit := 10 // default
		offset := 0 // default

		if l := r.URL.Query().Get("limit"); l != "" {
			if _, err := fmt.Sscanf(l, "%d", &limit); err != nil || limit <= 0 {
				limit = 10
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if _, err := fmt.Sscanf(o, "%d", &offset); err != nil || offset < 0 {
				offset = 0
			}
		}

		sims, err := simSvc.ListSimulations(r.Context(), limit, offset)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("upload_failed", "Failed to list simulations"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]any{
			"simulations": sims,
			"limit":       limit,
			"offset":      offset,
		})
	}
}

func CancelSimulation(simSvc services.SimulationServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Simulation ID is required"))
			return
		}

		var id int64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid simulation ID"))
			return
		}

		if err := simSvc.CancelSimulation(r.Context(), id); err != nil {
			WriteError(w, r, *models.NewAPIError("upload_failed", "Failed to cancel simulation"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "Simulation cancelled"})
	}
}

func GetContestResults(simSvc services.SimulationServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Simulation ID is required"))
			return
		}

		var id int64
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid simulation ID"))
			return
		}

		// Parse query parameters for pagination
		limit := 50 // default for contest results
		offset := 0 // default

		if l := r.URL.Query().Get("limit"); l != "" {
			if _, err := fmt.Sscanf(l, "%d", &limit); err != nil || limit <= 0 {
				limit = 50
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if _, err := fmt.Sscanf(o, "%d", &offset); err != nil || offset < 0 {
				offset = 0
			}
		}

		results, err := simSvc.GetContestResults(r.Context(), id, limit, offset)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("upload_failed", "Failed to get contest results"))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]any{
			"simulation_id": id,
			"results":       results,
			"limit":         limit,
			"offset":        offset,
		})
	}
}

func validateRecipe(recipe services.Recipe) error {
	if recipe.Version == "" {
		return fmt.Errorf("recipe version is required")
	}
	if recipe.Name == "" {
		return fmt.Errorf("recipe name is required")
	}
	if recipe.Parameters.SimPrevMax <= 0 {
		return fmt.Errorf("sim_prev_max must be positive")
	}
	if recipe.Parameters.SimPreds <= 0 {
		return fmt.Errorf("sim_preds must be positive")
	}
	if recipe.Parameters.Alpha < 0 || recipe.Parameters.Beta < 0 || recipe.Parameters.Gamma < 0 || recipe.Parameters.Delta < 0 {
		return fmt.Errorf("weights must be non-negative")
	}
	return nil
}

func AdvancedSimulation(
	configSvc services.ConfigServicer,
	simSvc services.SimulationServicer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Recipe       services.Recipe `json:"recipe"`
			StartContest int             `json:"start_contest"`
			EndContest   int             `json:"end_contest"`
			Async        bool            `json:"async"`
			SaveAsConfig bool            `json:"save_as_config,omitempty"`
			ConfigName   string          `json:"config_name,omitempty"`
			ConfigDesc   string          `json:"config_description,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid request format"))
			return
		}

		// Validate request
		if req.StartContest <= 0 || req.EndContest <= 0 || req.StartContest > req.EndContest {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid contest range"))
			return
		}

		// Validate recipe
		if err := validateRecipe(req.Recipe); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", err.Error()))
			return
		}

		// Optionally save as config
		if req.SaveAsConfig {
			if req.ConfigName == "" {
				WriteError(w, r, *models.NewAPIError("invalid_json", "Config name is required when saving as config"))
				return
			}

			// Create config
			_, err := configSvc.Create(r.Context(), services.CreateConfigRequest{
				Name:        req.ConfigName,
				Description: req.ConfigDesc,
				Recipe:      req.Recipe,
				Mode:        "advanced",
			})
			if err != nil {
				WriteError(w, r, *models.NewAPIError("upload_failed", "Failed to save config"))
				return
			}
		}

		// Create simulation
		sim, err := simSvc.CreateSimulation(r.Context(), services.CreateSimulationRequest{
			Mode:         "advanced",
			RecipeName:   req.Recipe.Name,
			Recipe:       req.Recipe,
			StartContest: req.StartContest,
			EndContest:   req.EndContest,
			Async:        req.Async,
		})
		if err != nil {
			WriteError(w, r, *models.NewAPIError("upload_failed", "Simulation creation failed"))
			return
		}

		// Return response
		response := map[string]any{
			"simulation_id": sim.ID,
			"status":        sim.Status,
		}
		if req.Async {
			response["message"] = "Simulation queued for processing"
			WriteJSON(w, http.StatusAccepted, response)
		} else {
			response["simulation"] = sim
			WriteJSON(w, http.StatusOK, response)
		}
	}
}
