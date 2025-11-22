package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/go-chi/chi/v5"
)

type ComparisonHandler struct {
	comparisonService services.ComparisonServicer
}

func NewComparisonHandler(comparisonService services.ComparisonServicer) *ComparisonHandler {
	return &ComparisonHandler{
		comparisonService: comparisonService,
	}
}

// CreateComparison creates a new comparison
// @Summary Create a new simulation comparison
// @Description Compare multiple simulations across various metrics
// @Tags comparisons
// @Accept json
// @Produce json
// @Param request body services.CompareRequest true "Comparison request"
// @Success 201 {object} services.ComparisonResult "Comparison created successfully"
// @Failure 400 {object} models.APIError "Invalid request"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/comparisons [post]
func CreateComparison(comparisonSvc services.ComparisonServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req services.CompareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid JSON in request body"))
			return
		}

		result, err := comparisonSvc.Compare(r.Context(), req)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("comparison_failed", err.Error()))
			return
		}

		WriteJSON(w, http.StatusCreated, result)
	}
}

// GetComparison retrieves a comparison by ID
// @Summary Get comparison by ID
// @Description Retrieve a specific comparison with full results
// @Tags comparisons
// @Accept json
// @Produce json
// @Param id path int true "Comparison ID"
// @Success 200 {object} services.ComparisonResult "Comparison details"
// @Failure 400 {object} models.APIError "Invalid ID"
// @Failure 404 {object} models.APIError "Comparison not found"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/comparisons/{id} [get]
func GetComparison(comparisonSvc services.ComparisonServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			WriteError(w, r, *models.NewAPIError("invalid_comparison_id", "Comparison ID is required"))
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_comparison_id", "Invalid comparison ID"))
			return
		}

		result, err := comparisonSvc.GetComparison(r.Context(), id)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("get_comparison_failed", err.Error()))
			return
		}

		WriteJSON(w, http.StatusOK, result)
	}
}

// ListComparisons lists comparisons with pagination
// @Summary List comparisons
// @Description Get a paginated list of comparisons
// @Tags comparisons
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of results" default(50)
// @Param offset query int false "Number of results to skip" default(0)
// @Success 200 {object} object{comparisons=[]object{id=integer,name=string,description=string,simulation_ids=[]integer,metrics=[]string,created_at=string},limit=integer,offset=integer} "List of comparisons"
// @Failure 500 {object} models.APIError "Internal server error"
// @Router /api/v1/comparisons [get]
func ListComparisons(comparisonSvc services.ComparisonServicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		limit := 50 // default
		offset := 0 // default

		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
				limit = l
			}
		}

		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		comparisons, err := comparisonSvc.ListComparisons(r.Context(), limit, offset)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("list_comparisons_failed", err.Error()))
			return
		}

		WriteJSON(w, http.StatusOK, map[string]any{
			"comparisons": comparisons,
			"limit":       limit,
			"offset":      offset,
		})
	}
}
