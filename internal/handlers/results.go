package handlers

import (
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

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, code int, err APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(err)
}

// APIError represents an API error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
