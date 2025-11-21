package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/store"
)

// ImportResult represents the result of an import operation
type ImportResult struct {
	ArtifactID   string    `json:"artifact_id"`
	Filename     string    `json:"filename"`
	ImportedAt   time.Time `json:"imported_at"`
	RowsInserted int       `json:"rows_inserted"`
	RowsSkipped  int       `json:"rows_skipped"`
	RowsErrors   int       `json:"rows_errors"`
	Duration     string    `json:"duration"`
	Message      string    `json:"message"`
}

// ResultsService handles results-related operations including import
type ResultsService struct {
	db            *store.DB
	logger        *slog.Logger
	uploadService *UploadService
	importService *ImportService
	tempDir       string
}

// NewResultsService creates a new results service
func NewResultsService(db *store.DB, logger *slog.Logger) *ResultsService {
	uploadService := NewUploadService(logger)
	importService := NewImportService(db, logger)

	return &ResultsService{
		db:            db,
		logger:        logger,
		uploadService: uploadService,
		importService: importService,
		tempDir:       "data/temp",
	}
}

// ImportArtifact imports data from a previously uploaded artifact
func (s *ResultsService) ImportArtifact(ctx context.Context, artifactID string, sheet string) (*ImportResult, error) {
	startTime := time.Now()

	// Find the artifact file
	artifactPath, filename, err := s.findArtifactFile(artifactID)
	if err != nil {
		return nil, fmt.Errorf("failed to find artifact: %w", err)
	}

	s.logger.Info("Starting import", "artifact_id", artifactID, "path", artifactPath)

	// Open the file
	file, err := os.Open(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open artifact file: %w", err)
	}
	defer file.Close()

	// Parse the XLSX file
	draws, err := s.importService.ParseXLSX(file, sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XLSX: %w", err)
	}

	s.logger.Info("Parsed XLSX", "draws_count", len(draws))

	// Import the draws
	importErr := s.importService.ImportDraws(ctx, draws)
	if importErr != nil {
		s.logger.Error("Import failed", "error", importErr)
		// For now, assume all draws failed if there's an error
		// In a more sophisticated implementation, we could track individual failures
		result := &ImportResult{
			ArtifactID:   artifactID,
			Filename:     filename,
			ImportedAt:   startTime,
			RowsInserted: 0,
			RowsSkipped:  0,
			RowsErrors:   len(draws),
			Duration:     time.Since(startTime).String(),
			Message:      fmt.Sprintf("Import failed: %v", importErr),
		}
		return result, importErr
	}

	duration := time.Since(startTime)

	result := &ImportResult{
		ArtifactID:   artifactID,
		Filename:     filename,
		ImportedAt:   startTime,
		RowsInserted: len(draws),
		RowsSkipped:  0,
		RowsErrors:   0,
		Duration:     duration.String(),
		Message:      s.getImportMessage(len(draws), 0, 0),
	}

	// Clean up the artifact file after successful import
	if err := os.Remove(artifactPath); err != nil {
		s.logger.Warn("Failed to clean up artifact file", "path", artifactPath, "error", err)
	}

	s.logger.Info("Import completed",
		"artifact_id", artifactID,
		"draws_count", len(draws),
		"duration", duration)

	return result, importErr
}

// findArtifactFile finds the artifact file by ID
func (s *ResultsService) findArtifactFile(artifactID string) (string, string, error) {
	// List all files in temp directory
	files, err := os.ReadDir(s.tempDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("temp directory does not exist")
		}
		return "", "", fmt.Errorf("failed to read temp directory: %w", err)
	}

	// Find file that starts with the artifact ID
	for _, file := range files {
		if !file.IsDir() && s.fileMatchesArtifact(file.Name(), artifactID) {
			path := filepath.Join(s.tempDir, file.Name())
			filename := s.extractOriginalFilename(file.Name(), artifactID)
			return path, filename, nil
		}
	}

	return "", "", fmt.Errorf("artifact %s not found", artifactID)
}

// fileMatchesArtifact checks if a filename matches the artifact ID pattern
func (s *ResultsService) fileMatchesArtifact(filename, artifactID string) bool {
	// File should start with artifactID and have a valid extension
	if !strings.HasPrefix(filename, artifactID) {
		return false
	}

	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".xlsx" || ext == ".xls"
}

// extractOriginalFilename extracts the original filename from the stored filename
func (s *ResultsService) extractOriginalFilename(storedName, artifactID string) string {
	// Remove artifact ID prefix and keep the extension
	// This is a simplified approach - in production you might want to store metadata
	withoutPrefix := strings.TrimPrefix(storedName, artifactID)
	if withoutPrefix == "" {
		return "unknown.xlsx" // fallback
	}
	// Remove leading underscore if present
	withoutPrefix = strings.TrimPrefix(withoutPrefix, "_")
	if withoutPrefix == "" {
		return "unknown.xlsx" // fallback
	}
	return withoutPrefix
}

// getImportMessage generates a human-readable import message
func (s *ResultsService) getImportMessage(inserted, skipped, errors int) string {
	total := inserted + skipped + errors

	if errors > 0 {
		return fmt.Sprintf("Import completed with errors: %d/%d rows imported successfully", inserted, total)
	}

	if skipped > 0 {
		return fmt.Sprintf("Import completed: %d/%d rows imported (%d skipped)", inserted, total, skipped)
	}

	return fmt.Sprintf("Import completed successfully: %d rows imported", inserted)
}

// GetDraw retrieves a single draw by contest number
func (s *ResultsService) GetDraw(ctx context.Context, contest int) (*models.Draw, error) {
	return s.importService.GetDraw(ctx, contest)
}

// ListDraws retrieves draws with pagination
func (s *ResultsService) ListDraws(ctx context.Context, limit, offset int) ([]models.Draw, error) {
	return s.importService.ListDraws(ctx, limit, offset)
}
