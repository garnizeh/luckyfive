package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/garnizeh/luckyfive/internal/services"
)

// UploadResponse represents the upload response
type UploadResponse struct {
	ArtifactID string `json:"artifact_id"`
	Filename   string `json:"filename"`
	Size       int64  `json:"size"`
	Message    string `json:"message"`
}

// UploadResults handles XLSX file uploads
func UploadResults(uploadSvc *services.UploadService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, APIError{
				Code:    "method_not_allowed",
				Message: "Method not allowed. Use POST.",
			})
			return
		}

		// Parse multipart form (max 50MB)
		const maxFileSize = 50 << 20 // 50MB
		err := r.ParseMultipartForm(maxFileSize)
		if err != nil {
			logger.Error("Failed to parse multipart form", "error", err)
			WriteError(w, http.StatusBadRequest, APIError{
				Code:    "invalid_form",
				Message: "Failed to parse multipart form",
			})
			return
		}

		// Get the file from form
		file, header, err := r.FormFile("file")
		if err != nil {
			logger.Error("No file provided", "error", err)
			WriteError(w, http.StatusBadRequest, APIError{
				Code:    "no_file",
				Message: "No file provided in 'file' field",
			})
			return
		}
		defer file.Close()

		// Upload file using service
		result, err := uploadSvc.UploadFile(file, header)
		if err != nil {
			logger.Error("File upload failed", "error", err, "filename", header.Filename)
			WriteError(w, http.StatusBadRequest, APIError{
				Code:    "upload_failed",
				Message: fmt.Sprintf("Upload failed: %v", err),
			})
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
	ArtifactID string `json:"artifact_id"`
	Sheet      string `json:"sheet,omitempty"` // Optional, defaults to first sheet
}

// ImportResults handles result import requests
func ImportResults(resultsSvc *services.ResultsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, APIError{
				Code:    "method_not_allowed",
				Message: "Method not allowed. Use POST.",
			})
			return
		}

		// Parse JSON request body
		var req ImportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to parse JSON request", "error", err)
			WriteError(w, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "Invalid JSON in request body",
			})
			return
		}

		// Validate required fields
		if req.ArtifactID == "" {
			WriteError(w, http.StatusBadRequest, APIError{
				Code:    "missing_artifact_id",
				Message: "artifact_id is required",
			})
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
			WriteError(w, http.StatusInternalServerError, APIError{
				Code:    "import_failed",
				Message: fmt.Sprintf("Import failed: %v", err),
			})
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
