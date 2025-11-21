package services

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/store"
)

// createTestResultsService creates a results service for testing
func createTestResultsService(t *testing.T) (*ResultsService, func()) {
	t.Helper()

	// Create test database
	resultsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create results DB: %v", err)
	}

	simulationsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create simulations DB: %v", err)
	}

	configsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create configs DB: %v", err)
	}

	financesDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create finances DB: %v", err)
	}

	db := &store.DB{
		ResultsDB:     resultsDB,
		SimulationsDB: simulationsDB,
		ConfigsDB:     configsDB,
		FinancesDB:    financesDB,
	}

	// Configure connection pools
	for _, sqlDB := range []*sql.DB{resultsDB, simulationsDB, configsDB, financesDB} {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "results_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := &ResultsService{
		db:            db,
		logger:        logger,
		uploadService: NewUploadService(logger),
		importService: NewImportService(db, logger),
		tempDir:       tempDir,
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
		db.Close()
	}

	return service, cleanup
}

func TestNewResultsService(t *testing.T) {
	// Create test database
	resultsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create results DB: %v", err)
	}
	defer resultsDB.Close()

	simulationsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create simulations DB: %v", err)
	}
	defer simulationsDB.Close()

	configsDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create configs DB: %v", err)
	}
	defer configsDB.Close()

	financesDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create finances DB: %v", err)
	}
	defer financesDB.Close()

	db := &store.DB{
		ResultsDB:     resultsDB,
		SimulationsDB: simulationsDB,
		ConfigsDB:     configsDB,
		FinancesDB:    financesDB,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewResultsService(db, logger)

	if service == nil {
		t.Fatal("NewResultsService returned nil")
	}
	if service.db != db {
		t.Error("Expected db to match")
	}
	if service.logger != logger {
		t.Error("Expected logger to match")
	}
	if service.tempDir != "data/temp" {
		t.Errorf("Expected tempDir to be 'data/temp', got '%s'", service.tempDir)
	}
	if service.uploadService == nil {
		t.Error("Expected uploadService to be initialized")
	}
	if service.importService == nil {
		t.Error("Expected importService to be initialized")
	}
}

func TestResultsService_findArtifactFile_ArtifactNotFound(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	_, _, err := service.findArtifactFile("nonexistent")

	if err == nil {
		t.Fatal("Expected error when artifact not found")
	}
	if err.Error() != "artifact nonexistent not found" {
		t.Errorf("Expected 'artifact nonexistent not found', got '%s'", err.Error())
	}
}

func TestResultsService_findArtifactFile_TempDirNotExist(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	// Remove temp directory
	os.RemoveAll(service.tempDir)

	_, _, err := service.findArtifactFile("test")

	if err == nil {
		t.Fatal("Expected error when temp directory doesn't exist")
	}
	if err.Error() != "temp directory does not exist" {
		t.Errorf("Expected 'temp directory does not exist', got '%s'", err.Error())
	}
}

func TestResultsService_fileMatchesArtifact(t *testing.T) {
	service, _ := createTestResultsService(t)

	tests := []struct {
		filename   string
		artifactID string
		expected   bool
	}{
		{"abc123_file.xlsx", "abc123", true},
		{"abc123_file.xls", "abc123", true},
		{"abc123.XLSX", "abc123", true},
		{"abc123.XLS", "abc123", true},
		{"def456_file.xlsx", "abc123", false},
		{"abc123_file.txt", "abc123", false},
		{"abc123", "abc123", false},
		{"file.xlsx", "abc123", false},
	}

	for _, test := range tests {
		result := service.fileMatchesArtifact(test.filename, test.artifactID)
		if result != test.expected {
			t.Errorf("fileMatchesArtifact(%s, %s) = %v, expected %v",
				test.filename, test.artifactID, result, test.expected)
		}
	}
}

func TestResultsService_extractOriginalFilename(t *testing.T) {
	service, _ := createTestResultsService(t)

	tests := []struct {
		storedName string
		artifactID string
		expected   string
	}{
		{"abc123.xlsx", "abc123", ".xlsx"},
		{"abc123.xls", "abc123", ".xls"},
		{"abc123", "abc123", "unknown.xlsx"},
	}

	for _, test := range tests {
		result := service.extractOriginalFilename(test.storedName, test.artifactID)
		if result != test.expected {
			t.Errorf("extractOriginalFilename(%s, %s) = %s, expected %s",
				test.storedName, test.artifactID, result, test.expected)
		}
	}
}

func TestResultsService_getImportMessage(t *testing.T) {
	service, _ := createTestResultsService(t)

	tests := []struct {
		inserted int
		skipped  int
		errors   int
		expected string
	}{
		{100, 0, 0, "Import completed successfully: 100 rows imported"},
		{95, 5, 0, "Import completed: 95/100 rows imported (5 skipped)"},
		{90, 0, 10, "Import completed with errors: 90/100 rows imported successfully"},
		{85, 5, 10, "Import completed with errors: 85/100 rows imported successfully"},
	}

	for _, test := range tests {
		result := service.getImportMessage(test.inserted, test.skipped, test.errors)
		if result != test.expected {
			t.Errorf("getImportMessage(%d, %d, %d) = %s, expected %s",
				test.inserted, test.skipped, test.errors, result, test.expected)
		}
	}
}

func TestResultsService_ImportArtifact_ArtifactNotFound(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	ctx := context.Background()
	result, err := service.ImportArtifact(ctx, "nonexistent", "")

	if err == nil {
		t.Fatal("Expected error when artifact not found")
	}
	if result != nil {
		t.Error("Expected nil result when artifact not found")
	}
	if err.Error() != "failed to find artifact: artifact nonexistent not found" {
		t.Errorf("Expected specific error message, got '%s'", err.Error())
	}
}
