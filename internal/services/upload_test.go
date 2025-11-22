package services

import (
	"bytes"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestUploadService creates an upload service for testing
func createTestUploadService(t *testing.T) (*UploadService, func()) {
	t.Helper()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "upload_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &UploadService{
		logger:  logger,
		tempDir: tempDir,
		maxSize: 50 << 20, // 50MB
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return service, cleanup
}

// createMultipartFile creates a multipart file for testing
func createMultipartFile(t *testing.T, filename, content string) (multipart.File, *multipart.FileHeader) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = part.Write([]byte(content))
	if err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	reader := multipart.NewReader(&buf, writer.Boundary())
	form, err := reader.ReadForm(32 << 20) // 32MB max memory
	if err != nil {
		t.Fatalf("Failed to read form: %v", err)
	}

	fileHeader := form.File["file"][0]
	file, err := fileHeader.Open()
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	header := fileHeader

	return file, header
}

func TestNewUploadService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := NewUploadService(logger)

	if service == nil {
		t.Fatal("NewUploadService returned nil")
	}
	if service.logger != logger {
		t.Error("Logger not set correctly")
	}
	if service.tempDir != "data/temp" {
		t.Errorf("Expected tempDir to be 'data/temp', got '%s'", service.tempDir)
	}
	if service.maxSize != 50<<20 {
		t.Errorf("Expected maxSize to be 50MB, got %d", service.maxSize)
	}
}

func TestUploadService_UploadFile_ValidXLSX(t *testing.T) {
	service, cleanup := createTestUploadService(t)
	defer cleanup()

	content := "test xlsx content"
	file, header := createMultipartFile(t, "test.xlsx", content)
	defer file.Close()

	result, err := service.UploadFile(file, header)

	if err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
	if result == nil {
		t.Fatal("UploadFile returned nil result")
	}
	if result.Filename != "test.xlsx" {
		t.Errorf("Expected filename to be 'test.xlsx', got '%s'", result.Filename)
	}
	if result.Size != int64(len(content)) {
		t.Errorf("Expected size to be %d, got %d", len(content), result.Size)
	}
	if result.ArtifactID == "" {
		t.Error("ArtifactID should not be empty")
	}
	if !strings.HasPrefix(result.Path, service.tempDir) {
		t.Errorf("Expected path to start with tempDir, got '%s'", result.Path)
	}

	// Verify file was actually saved
	if _, err := os.Stat(result.Path); os.IsNotExist(err) {
		t.Errorf("File was not saved to disk: %s", result.Path)
	}
}

func TestUploadService_UploadFile_ValidXLS(t *testing.T) {
	service, cleanup := createTestUploadService(t)
	defer cleanup()

	content := "test xls content"
	file, header := createMultipartFile(t, "test.xls", content)
	defer file.Close()

	result, err := service.UploadFile(file, header)

	if err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
	if result == nil {
		t.Fatal("UploadFile returned nil result")
	}
	if result.Filename != "test.xls" {
		t.Errorf("Expected filename to be 'test.xls', got '%s'", result.Filename)
	}
	if result.ArtifactID == "" {
		t.Error("ArtifactID should not be empty")
	}
}

func TestUploadService_UploadFile_InvalidFileType(t *testing.T) {
	service, cleanup := createTestUploadService(t)
	defer cleanup()

	content := "test content"
	file, header := createMultipartFile(t, "test.txt", content)
	defer file.Close()

	result, err := service.UploadFile(file, header)

	if err == nil {
		t.Fatal("UploadFile should have returned error for invalid file type")
	}
	if result != nil {
		t.Error("UploadFile should have returned nil result for invalid file type")
	}
	if !strings.Contains(err.Error(), "invalid file type") {
		t.Errorf("Expected error to contain 'invalid file type', got '%s'", err.Error())
	}
}

func TestUploadService_UploadFile_FileTooLarge(t *testing.T) {
	service, cleanup := createTestUploadService(t)
	defer cleanup()

	// Create a service with very small max size for testing
	service.maxSize = 10 // 10 bytes

	content := "this content is way too long for the limit"
	file, header := createMultipartFile(t, "test.xlsx", content)
	defer file.Close()

	// Manually set the header size to exceed limit
	header.Size = int64(len(content))

	result, err := service.UploadFile(file, header)

	if err == nil {
		t.Fatal("UploadFile should have returned error for file too large")
	}
	if result != nil {
		t.Error("UploadFile should have returned nil result for file too large")
	}
	if !strings.Contains(err.Error(), "exceeds maximum") {
		t.Errorf("Expected error to contain 'exceeds maximum', got '%s'", err.Error())
	}
}

func TestUploadService_IsValidFileType(t *testing.T) {
	service, _ := createTestUploadService(t)

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.xlsx", true},
		{"test.XLSX", true},
		{"test.xls", true},
		{"test.XLS", true},
		{"test.txt", false},
		{"test.xlsx.exe", false},
		{"test", false},
		{"", false},
	}

	for _, test := range tests {
		result := service.isValidFileType(test.filename)
		if result != test.expected {
			t.Errorf("isValidFileType(%s) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}

func TestUploadService_GenerateArtifactID(t *testing.T) {
	service, cleanup := createTestUploadService(t)
	defer cleanup()

	id1, err := service.generateArtifactID()
	if err != nil {
		t.Fatalf("generateArtifactID returned error: %v", err)
	}
	if len(id1) != 32 { // 16 bytes * 2 hex chars per byte
		t.Errorf("Expected artifact ID length to be 32, got %d", len(id1))
	}

	id2, err := service.generateArtifactID()
	if err != nil {
		t.Fatalf("generateArtifactID returned error: %v", err)
	}

	// IDs should be unique
	if id1 == id2 {
		t.Error("Generated artifact IDs should be unique")
	}

	// Should be valid hex
	for _, char := range id1 {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("Artifact ID contains invalid character: %c", char)
		}
	}
}

func TestUploadService_UploadFile_CreatesTempDirectory(t *testing.T) {
	// Create service with non-existent temp directory
	tempDir := filepath.Join(os.TempDir(), "upload_test_nonexistent")
	defer os.RemoveAll(tempDir) // cleanup

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &UploadService{
		logger:  logger,
		tempDir: tempDir,
		maxSize: 50 << 20,
	}

	content := "test content"
	file, header := createMultipartFile(t, "test.xlsx", content)
	defer file.Close()

	// Directory shouldn't exist yet
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Fatalf("Temp directory should not exist before upload")
	}

	result, err := service.UploadFile(file, header)

	if err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
	if result == nil {
		t.Fatal("UploadFile returned nil result")
	}

	// Directory should now exist
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Temp directory was not created: %s", tempDir)
	}
}

func TestUploadService_SetTempDir(t *testing.T) {
	service, cleanup := createTestUploadService(t)
	defer cleanup()

	// Test setting a valid directory
	service.SetTempDir("/tmp/test")
	if service.tempDir != "/tmp/test" {
		t.Errorf("Expected tempDir to be '/tmp/test', got '%s'", service.tempDir)
	}

	// Test setting an empty directory (should not change)
	service.SetTempDir("")
	if service.tempDir != "/tmp/test" {
		t.Errorf("Expected tempDir to remain '/tmp/test', got '%s'", service.tempDir)
	}
}

func TestUploadService_SetMaxSize(t *testing.T) {
	service, cleanup := createTestUploadService(t)
	defer cleanup()

	// Test setting a valid max size
	service.SetMaxSize(100 << 20) // 100MB
	if service.maxSize != 100<<20 {
		t.Errorf("Expected maxSize to be %d, got %d", 100<<20, service.maxSize)
	}

	// Test setting an invalid max size (should not change)
	service.SetMaxSize(0)
	if service.maxSize != 100<<20 {
		t.Errorf("Expected maxSize to remain %d, got %d", 100<<20, service.maxSize)
	}

	// Test setting a negative max size (should not change)
	service.SetMaxSize(-1)
	if service.maxSize != 100<<20 {
		t.Errorf("Expected maxSize to remain %d, got %d", 100<<20, service.maxSize)
	}
}
