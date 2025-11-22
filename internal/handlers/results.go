package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/go-chi/chi/v5"
)

// UploadResponse represents the upload response
type UploadResponse struct {
	ArtifactID string `json:"artifact_id"`
	Filename   string `json:"filename"`
	Size       int64  `json:"size"`
	Message    string `json:"message"`
}

// UploadResults handles XLSX file uploads
// UploadResults godoc
// @Summary Upload XLSX results file
// @Description Accepts a multipart/form-data upload containing an XLSX file. Returns an artifact_id to be used for import.
// @Tags results
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "XLSX file"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /api/v1/results/upload [post]
func UploadResults(uploadSvc *services.UploadService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			WriteError(w, r, *models.NewAPIError("method_not_allowed", "Method not allowed. Use POST."))
			return
		}

		// Parse multipart form (max 50MB)
		const maxFileSize = 50 << 20 // 50MB
		err := r.ParseMultipartForm(maxFileSize)
		if err != nil {
			logger.Error("Failed to parse multipart form", "error", err)
			WriteError(w, r, *models.NewAPIError("invalid_form", "Failed to parse multipart form"))
			return
		}

		// Get the file from form
		file, header, err := r.FormFile("file")
		if err != nil {
			logger.Error("No file provided", "error", err)
			WriteError(w, r, *models.NewAPIError("no_file", "No file provided in 'file' field"))
			return
		}
		defer file.Close()

		// Upload file using service
		result, err := uploadSvc.UploadFile(file, header)
		if err != nil {
			logger.Error("File upload failed", "error", err, "filename", header.Filename)
			WriteError(w, r, *models.NewAPIError("upload_failed", fmt.Sprintf("Upload failed: %v", err)))
			return
		}

		// Return success response
		response := UploadResponse{
			ArtifactID: result.ArtifactID,
			Filename:   result.Filename,
			Size:       result.Size,
			Message:    "File uploaded successfully",
		}

		WriteJSON(w, http.StatusOK, response)
	}
}

// ImportRequest represents the import request body
type ImportRequest struct {
	ArtifactID string `json:"artifact_id" validate:"required"`
	Sheet      string `json:"sheet,omitempty"` // Optional, defaults to first sheet
}

// ImportResults handles result import requests
// ImportResults godoc
// @Summary Trigger import for uploaded artifact
// @Description Triggers import for a previously uploaded artifact identified by artifact_id. Optional sheet name may be provided.
// @Tags results
// @Accept application/json
// @Produce json
// @Param request body ImportRequest true "Import request"
// @Success 200 {object} services.ImportResult
// @Failure 400 {object} models.ValidationError
// @Failure 500 {object} models.APIError
// @Router /api/v1/results/import [post]
func ImportResults(resultsSvc *services.ResultsService, logger *slog.Logger) http.HandlerFunc {
	validate := validator.New()

	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			WriteError(w, r, *models.NewAPIError("method_not_allowed", "Method not allowed. Use POST."))
			return
		}

		// Parse JSON request body
		var req ImportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to parse JSON request", "error", err)
			WriteError(w, r, *models.NewAPIError("invalid_json", "Invalid JSON in request body"))
			return
		}

		// Validate request
		if err := validate.Struct(req); err != nil {
			logger.Error("Request validation failed", "error", err)
			fieldErrors := make(map[string]string)
			for _, err := range err.(validator.ValidationErrors) {
				fieldErrors[err.Field()] = fmt.Sprintf("Field validation failed on '%s' tag", err.Tag())
			}
			WriteValidationError(w, r, *models.NewValidationError("Request validation failed", fieldErrors))
			return
		}

		// Default sheet if not provided
		sheet := req.Sheet
		if sheet == "" {
			sheet = "" // Empty string means first sheet
		}

		logger.Info("Starting import", "artifact_id", req.ArtifactID, "sheet", sheet)

		// Import the artifact
		ctx := context.Background()
		result, err := resultsSvc.ImportArtifact(ctx, req.ArtifactID, sheet)
		if err != nil {
			logger.Error("Import failed", "artifact_id", req.ArtifactID, "error", err)
			WriteError(w, r, *models.NewAPIError("import_failed", fmt.Sprintf("Import failed: %v", err)))
			return
		}

		logger.Info("Import completed successfully",
			"artifact_id", req.ArtifactID,
			"rows_inserted", result.RowsInserted,
			"rows_skipped", result.RowsSkipped,
			"rows_errors", result.RowsErrors)

		// Return success response
		WriteJSON(w, http.StatusOK, result)
	}
}

// GetDraw handles GET /api/v1/results/{contest}
// GetDraw godoc
// @Summary Get a single draw
// @Description Retrieve a single draw by contest number
// @Tags results
// @Accept json
// @Produce json
// @Param contest path int true "Contest number"
// @Success 200 {object} models.Draw
// @Failure 400 {object} models.APIError
// @Failure 404 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /api/v1/results/{contest} [get]
func GetDraw(resultsSvc *services.ResultsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contestStr := chi.URLParam(r, "contest")
		if contestStr == "" {
			WriteError(w, r, *models.NewAPIError("missing_contest", "contest parameter is required"))
			return
		}

		contest, err := strconv.Atoi(contestStr)
		if err != nil {
			WriteError(w, r, *models.NewAPIError("invalid_contest", "contest must be a valid number"))
			return
		}

		logger.Info("Getting draw", "contest", contest)

		draw, err := resultsSvc.GetDraw(r.Context(), contest)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				WriteError(w, r, *models.NewAPIError("draw_not_found", "draw not found"))
				return
			}
			logger.Error("Failed to get draw", "error", err, "contest", contest)
			WriteError(w, r, *models.NewAPIError("get_draw_failed", "failed to get draw"))
			return
		}

		WriteJSON(w, http.StatusOK, draw)
	}
}

// ListDraws handles GET /api/v1/results
// ListDraws godoc
// @Summary List draws with pagination
// @Description Returns a paginated list of draws. Query params: limit (default 50), offset (default 0).
// @Tags results
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]any
// @Failure 400 {object} models.APIError
// @Failure 500 {object} models.APIError
// @Router /api/v1/results [get]
func ListDraws(resultsSvc *services.ResultsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		limit := 50 // default limit
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
				limit = l
			} else {
				WriteError(w, r, *models.NewAPIError("invalid_limit", "limit must be a number between 1 and 1000"))
				return
			}
		}

		offset := 0 // default offset
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			} else {
				WriteError(w, r, *models.NewAPIError("invalid_offset", "offset must be a non-negative number"))
				return
			}
		}

		logger.Info("Listing draws", "limit", limit, "offset", offset)

		draws, err := resultsSvc.ListDraws(r.Context(), limit, offset)
		if err != nil {
			logger.Error("Failed to list draws", "error", err)
			WriteError(w, r, *models.NewAPIError("list_draws_failed", "failed to list draws"))
			return
		}

		response := map[string]any{
			"draws":  draws,
			"limit":  limit,
			"offset": offset,
			"count":  len(draws),
		}

		WriteJSON(w, http.StatusOK, response)
	}
}
