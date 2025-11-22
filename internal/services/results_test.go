package services

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/store"
	"github.com/xuri/excelize/v2"
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

func TestResultsService_SetTempDir(t *testing.T) {
	service, cleanup := createTestResultsService(t)
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

func TestResultsService_GetDraw(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	ctx := context.Background()

	// Test getting a draw (this will likely return an error since no data is set up)
	_, err := service.GetDraw(ctx, 1000)

	// We expect an error since we're using an in-memory DB with no data
	if err == nil {
		t.Error("Expected error when getting draw from empty database")
	}
}

func TestResultsService_ListDraws(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	ctx := context.Background()

	// Test listing draws (this will return an error since no schema is set up)
	_, err := service.ListDraws(ctx, 10, 0)

	// We expect an error since we're using an in-memory DB with no schema
	if err == nil {
		t.Error("Expected error when listing draws from database with no schema")
	}
}

func TestResultsService_findArtifactFile_Success(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	// Create a test file
	testFile := "abc123_test.xlsx"
	filePath := filepath.Join(service.tempDir, testFile)
	err := os.WriteFile(filePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	path, filename, err := service.findArtifactFile("abc123")

	if err != nil {
		t.Errorf("Expected no error finding artifact, got %v", err)
	}
	if path != filePath {
		t.Errorf("Expected path %s, got %s", filePath, path)
	}
	if filename != "test.xlsx" {
		t.Errorf("Expected filename 'test.xlsx', got '%s'", filename)
	}
}

func TestResultsService_ImportArtifact_FindArtifactSuccess(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	// Create a test XLSX file (invalid content, but findArtifactFile should succeed)
	testFile := "abc123_test.xlsx"
	filePath := filepath.Join(service.tempDir, testFile)
	err := os.WriteFile(filePath, []byte("invalid xlsx content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ctx := context.Background()
	result, err := service.ImportArtifact(ctx, "abc123", "")

	// We expect it to fail at XLSX parsing, not at findArtifactFile
	if err == nil {
		t.Fatal("Expected error during XLSX parsing")
	}
	if result != nil {
		t.Fatal("Expected nil result when XLSX parsing fails")
	}
	// The key point is that findArtifactFile succeeded, so we got past that step
	// and the error is from XLSX parsing, not artifact finding
	if !strings.Contains(err.Error(), "failed to parse XLSX") {
		t.Errorf("Expected error to contain 'failed to parse XLSX', got '%s'", err.Error())
	}
}

func TestResultsService_ImportArtifact_ParseXLSXSuccess(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	// Create a valid XLSX file using excelize
	f := excelize.NewFile()
	defer f.Close()

	// Set headers
	f.SetCellValue("Sheet1", "A1", "Concurso")
	f.SetCellValue("Sheet1", "B1", "Data")
	f.SetCellValue("Sheet1", "C1", "Bola1")
	f.SetCellValue("Sheet1", "D1", "Bola2")
	f.SetCellValue("Sheet1", "E1", "Bola3")
	f.SetCellValue("Sheet1", "F1", "Bola4")
	f.SetCellValue("Sheet1", "G1", "Bola5")

	// Set data row
	f.SetCellValue("Sheet1", "A2", 1000)
	f.SetCellValue("Sheet1", "B2", "2024-01-01")
	f.SetCellValue("Sheet1", "C2", 1)
	f.SetCellValue("Sheet1", "D2", 2)
	f.SetCellValue("Sheet1", "E2", 3)
	f.SetCellValue("Sheet1", "F2", 4)
	f.SetCellValue("Sheet1", "G2", 5)

	// Save to temp file
	testFile := "def456_test.xlsx"
	filePath := filepath.Join(service.tempDir, testFile)
	err := f.SaveAs(filePath)
	if err != nil {
		t.Fatalf("Failed to create test XLSX file: %v", err)
	}

	ctx := context.Background()
	result, err := service.ImportArtifact(ctx, "def456", "")

	// We expect it to succeed in finding artifact and parsing XLSX
	// It may fail at import due to DB schema, but ParseXLSX should succeed
	if err == nil {
		t.Log("Import succeeded completely")
	} else {
		// Check that the error is not from findArtifactFile or ParseXLSX
		errStr := err.Error()
		if strings.Contains(errStr, "failed to find artifact") {
			t.Errorf("findArtifactFile failed: %s", errStr)
		}
		if strings.Contains(errStr, "failed to parse XLSX") {
			t.Errorf("ParseXLSX failed: %s", errStr)
		}
		// If it fails at import, that's expected since DB may not have schema
		if strings.Contains(errStr, "failed to import draws") {
			t.Log("ParseXLSX succeeded, failed at import (expected)")
		}
	}

	if result != nil {
		if result.ArtifactID != "def456" {
			t.Errorf("Expected artifact ID 'def456', got '%s'", result.ArtifactID)
		}
		if result.Filename != "test.xlsx" {
			t.Errorf("Expected filename 'test.xlsx', got '%s'", result.Filename)
		}
	}
}

func TestResultsService_ImportArtifact_ImportDrawsSuccess(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	// Setup database schema
	schema := `
CREATE TABLE IF NOT EXISTS draws (
  contest INTEGER PRIMARY KEY,
  draw_date TEXT NOT NULL,
  bola1 INTEGER NOT NULL CHECK(bola1 BETWEEN 1 AND 80),
  bola2 INTEGER NOT NULL CHECK(bola2 BETWEEN 1 AND 80),
  bola3 INTEGER NOT NULL CHECK(bola3 BETWEEN 1 AND 80),
  bola4 INTEGER NOT NULL CHECK(bola4 BETWEEN 1 AND 80),
  bola5 INTEGER NOT NULL CHECK(bola5 BETWEEN 1 AND 80),
  source TEXT,
  imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  raw_row TEXT,
  CHECK(bola1 < bola2 AND bola2 < bola3 AND bola3 < bola4 AND bola4 < bola5)
);
CREATE INDEX IF NOT EXISTS idx_draws_draw_date ON draws(draw_date);
CREATE INDEX IF NOT EXISTS idx_draws_imported_at ON draws(imported_at);
CREATE TABLE IF NOT EXISTS import_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  filename TEXT NOT NULL,
  imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  rows_inserted INTEGER NOT NULL DEFAULT 0,
  rows_skipped INTEGER NOT NULL DEFAULT 0,
  rows_errors INTEGER NOT NULL DEFAULT 0,
  source_hash TEXT,
  metadata TEXT
);
`
	_, err := service.db.ResultsDB.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to setup database schema: %v", err)
	}

	// Create a valid XLSX file using excelize
	f := excelize.NewFile()
	defer f.Close()

	// Set headers
	f.SetCellValue("Sheet1", "A1", "Concurso")
	f.SetCellValue("Sheet1", "B1", "Data")
	f.SetCellValue("Sheet1", "C1", "Bola1")
	f.SetCellValue("Sheet1", "D1", "Bola2")
	f.SetCellValue("Sheet1", "E1", "Bola3")
	f.SetCellValue("Sheet1", "F1", "Bola4")
	f.SetCellValue("Sheet1", "G1", "Bola5")

	// Set data row
	f.SetCellValue("Sheet1", "A2", 1000)
	f.SetCellValue("Sheet1", "B2", "2024-01-01")
	f.SetCellValue("Sheet1", "C2", 1)
	f.SetCellValue("Sheet1", "D2", 2)
	f.SetCellValue("Sheet1", "E2", 3)
	f.SetCellValue("Sheet1", "F2", 4)
	f.SetCellValue("Sheet1", "G2", 5)

	// Save to temp file
	testFile := "ghi789_test.xlsx"
	filePath := filepath.Join(service.tempDir, testFile)
	err = f.SaveAs(filePath)
	if err != nil {
		t.Fatalf("Failed to create test XLSX file: %v", err)
	}

	ctx := context.Background()
	result, err := service.ImportArtifact(ctx, "ghi789", "")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.RowsInserted != 1 {
		t.Errorf("Expected 1 row inserted, got %d", result.RowsInserted)
	}
	if result.RowsErrors != 0 {
		t.Errorf("Expected 0 errors, got %d", result.RowsErrors)
	}
	if result.ArtifactID != "ghi789" {
		t.Errorf("Expected artifact ID 'ghi789', got '%s'", result.ArtifactID)
	}
	if result.Filename != "test.xlsx" {
		t.Errorf("Expected filename 'test.xlsx', got '%s'", result.Filename)
	}
}

func TestResultsService_ImportArtifact_ParseXLSX_GetRowsFailure(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	// Create an XLSX file with invalid sheet
	f := excelize.NewFile()
	defer f.Close()

	// Add a sheet with data
	f.SetCellValue("ValidSheet", "A1", "Concurso")

	// Save to temp file
	testFile := "jkl012_test.xlsx"
	filePath := filepath.Join(service.tempDir, testFile)
	err := f.SaveAs(filePath)
	if err != nil {
		t.Fatalf("Failed to create test XLSX file: %v", err)
	}

	ctx := context.Background()
	result, err := service.ImportArtifact(ctx, "jkl012", "InvalidSheet")

	if err == nil {
		t.Fatal("Expected error when sheet does not exist")
	}
	if result != nil {
		t.Error("Expected nil result when ParseXLSX fails")
	}
	if !strings.Contains(err.Error(), "failed to parse XLSX") {
		t.Errorf("Expected error to contain 'failed to parse XLSX', got '%s'", err.Error())
	}
}

func TestResultsService_ImportArtifact_ParseXLSX_InsufficientRows(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	// Create an XLSX file with only headers (no data rows)
	f := excelize.NewFile()
	defer f.Close()

	// Set headers only
	f.SetCellValue("Sheet1", "A1", "Concurso")
	f.SetCellValue("Sheet1", "B1", "Data")
	f.SetCellValue("Sheet1", "C1", "Bola1")

	// Save to temp file
	testFile := "mno345_test.xlsx"
	filePath := filepath.Join(service.tempDir, testFile)
	err := f.SaveAs(filePath)
	if err != nil {
		t.Fatalf("Failed to create test XLSX file: %v", err)
	}

	ctx := context.Background()
	result, err := service.ImportArtifact(ctx, "mno345", "")

	if err == nil {
		t.Fatal("Expected error when XLSX has insufficient rows")
	}
	if result != nil {
		t.Error("Expected nil result when ParseXLSX fails")
	}
	if !strings.Contains(err.Error(), "failed to parse XLSX") {
		t.Errorf("Expected error to contain 'failed to parse XLSX', got '%s'", err.Error())
	}
}

func TestResultsService_ImportArtifact_ParseXLSX_DetectColumnsFailure(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	// Create an XLSX file with invalid headers
	f := excelize.NewFile()
	defer f.Close()

	// Set invalid headers
	f.SetCellValue("Sheet1", "A1", "InvalidCol1")
	f.SetCellValue("Sheet1", "B1", "InvalidCol2")
	f.SetCellValue("Sheet1", "C1", "InvalidCol3")

	// Add a data row
	f.SetCellValue("Sheet1", "A2", "data")

	// Save to temp file
	testFile := "pqr678_test.xlsx"
	filePath := filepath.Join(service.tempDir, testFile)
	err := f.SaveAs(filePath)
	if err != nil {
		t.Fatalf("Failed to create test XLSX file: %v", err)
	}

	ctx := context.Background()
	result, err := service.ImportArtifact(ctx, "pqr678", "")

	if err == nil {
		t.Fatal("Expected error when column detection fails")
	}
	if result != nil {
		t.Error("Expected nil result when ParseXLSX fails")
	}
	if !strings.Contains(err.Error(), "failed to parse XLSX") {
		t.Errorf("Expected error to contain 'failed to parse XLSX', got '%s'", err.Error())
	}
}

func TestResultsService_ImportArtifact_ParseXLSX_ParseRowColumnOutOfRange(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	// Create an XLSX file with headers but data row has fewer columns
	f := excelize.NewFile()
	defer f.Close()

	// Set headers
	f.SetCellValue("Sheet1", "A1", "Concurso")
	f.SetCellValue("Sheet1", "B1", "Data")
	f.SetCellValue("Sheet1", "C1", "Bola1")
	f.SetCellValue("Sheet1", "D1", "Bola2")
	f.SetCellValue("Sheet1", "E1", "Bola3")
	f.SetCellValue("Sheet1", "F1", "Bola4")
	f.SetCellValue("Sheet1", "G1", "Bola5")

	// Set data row with only 2 columns (insufficient for contest and date)
	f.SetCellValue("Sheet1", "A2", 1000)
	f.SetCellValue("Sheet1", "B2", "2024-01-01")

	// Save to temp file
	testFile := "stu901_test.xlsx"
	filePath := filepath.Join(service.tempDir, testFile)
	err := f.SaveAs(filePath)
	if err != nil {
		t.Fatalf("Failed to create test XLSX file: %v", err)
	}

	ctx := context.Background()
	result, err := service.ImportArtifact(ctx, "stu901", "")

	if err == nil {
		t.Fatal("Expected error when parsing row with insufficient columns")
	}
	if result == nil {
		t.Error("Expected result even when import fails")
	}
	if !strings.Contains(err.Error(), "no draws to import") {
		t.Errorf("Expected error to contain 'no draws to import', got '%s'", err.Error())
	}
}

func TestResultsService_ImportArtifact_ImportDraws_EmptyDraws(t *testing.T) {
	service, cleanup := createTestResultsService(t)
	defer cleanup()

	// Setup database schema
	schema := `
CREATE TABLE IF NOT EXISTS draws (
  contest INTEGER PRIMARY KEY,
  draw_date TEXT NOT NULL,
  bola1 INTEGER NOT NULL CHECK(bola1 BETWEEN 1 AND 80),
  bola2 INTEGER NOT NULL CHECK(bola2 BETWEEN 1 AND 80),
  bola3 INTEGER NOT NULL CHECK(bola3 BETWEEN 1 AND 80),
  bola4 INTEGER NOT NULL CHECK(bola4 BETWEEN 1 AND 80),
  bola5 INTEGER NOT NULL CHECK(bola5 BETWEEN 1 AND 80),
  source TEXT,
  imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  raw_row TEXT,
  CHECK(bola1 < bola2 AND bola2 < bola3 AND bola3 < bola4 AND bola4 < bola5)
);
`
	_, err := service.db.ResultsDB.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to setup database schema: %v", err)
	}

	// Create an XLSX file with headers and invalid data row that gets skipped
	f := excelize.NewFile()
	defer f.Close()

	// Set headers
	f.SetCellValue("Sheet1", "A1", "Concurso")
	f.SetCellValue("Sheet1", "B1", "Data")
	f.SetCellValue("Sheet1", "C1", "Bola1")
	f.SetCellValue("Sheet1", "D1", "Bola2")
	f.SetCellValue("Sheet1", "E1", "Bola3")
	f.SetCellValue("Sheet1", "F1", "Bola4")
	f.SetCellValue("Sheet1", "G1", "Bola5")

	// Set invalid data row (invalid contest - not a number)
	f.SetCellValue("Sheet1", "A2", "invalid_contest")
	f.SetCellValue("Sheet1", "B2", "2024-01-01")
	f.SetCellValue("Sheet1", "C2", 1)
	f.SetCellValue("Sheet1", "D2", 2)
	f.SetCellValue("Sheet1", "E2", 3)
	f.SetCellValue("Sheet1", "F2", 4)
	f.SetCellValue("Sheet1", "G2", 5)

	// Save to temp file
	testFile := "vwx234_test.xlsx"
	filePath := filepath.Join(service.tempDir, testFile)
	err = f.SaveAs(filePath)
	if err != nil {
		t.Fatalf("Failed to create test XLSX file: %v", err)
	}

	ctx := context.Background()
	result, err := service.ImportArtifact(ctx, "vwx234", "")

	if err == nil {
		t.Fatal("Expected error when no valid draws to import")
	}
	if result == nil {
		t.Error("Expected result even when import fails")
	}
	if !strings.Contains(err.Error(), "no draws to import") {
		t.Errorf("Expected error to contain 'no draws to import', got '%s'", err.Error())
	}
}
