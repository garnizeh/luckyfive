package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// UploadService handles file upload operations
type UploadService struct {
	logger  *slog.Logger
	tempDir string
	maxSize int64
}

// NewUploadService creates a new upload service
func NewUploadService(logger *slog.Logger) *UploadService {
	return &UploadService{
		logger:  logger,
		tempDir: "data/temp",
		maxSize: 50 << 20, // 50MB
	}
}

// SetTempDir overrides the temporary directory used to store uploaded files.
// Useful for tests and alternative runtime configurations.
func (s *UploadService) SetTempDir(dir string) {
	if dir == "" {
		return
	}
	s.tempDir = dir
}

// SetMaxSize allows overriding the maximum accepted file size (in bytes).
// Useful for tests that want to simulate larger/smaller limits.
func (s *UploadService) SetMaxSize(size int64) {
	if size <= 0 {
		return
	}
	s.maxSize = size
}

// UploadResult represents the result of a file upload
type UploadResult struct {
	ArtifactID string
	Filename   string
	Size       int64
	Path       string
}

// UploadFile validates and saves an uploaded file
func (s *UploadService) UploadFile(file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// Validate file type
	if !s.isValidFileType(header.Filename) {
		return nil, fmt.Errorf("invalid file type: %s", header.Filename)
	}

	// Validate file size
	if header.Size > s.maxSize {
		return nil, fmt.Errorf("file size %d bytes exceeds maximum %d bytes", header.Size, s.maxSize)
	}

	// Generate artifact ID
	artifactID, err := s.generateArtifactID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate artifact ID: %w", err)
	}

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(s.tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Save file temporarily
	tempPath := filepath.Join(s.tempDir, artifactID+filepath.Ext(header.Filename))
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Copy file content
	size, err := io.Copy(tempFile, file)
	if err != nil {
		os.Remove(tempPath) // Clean up on error
		return nil, fmt.Errorf("failed to save file content: %w", err)
	}

	s.logger.Info("File uploaded successfully",
		"artifact_id", artifactID,
		"filename", header.Filename,
		"size", size,
		"temp_path", tempPath)

	return &UploadResult{
		ArtifactID: artifactID,
		Filename:   header.Filename,
		Size:       size,
		Path:       tempPath,
	}, nil
}

// isValidFileType checks if the file extension is allowed
func (s *UploadService) isValidFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".xlsx" || ext == ".xls"
}

// generateArtifactID generates a random artifact ID
func (s *UploadService) generateArtifactID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
