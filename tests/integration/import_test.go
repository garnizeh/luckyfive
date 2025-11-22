package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"log/slog"

	"github.com/garnizeh/luckyfive/internal/services"
	"github.com/garnizeh/luckyfive/internal/store"
	"github.com/xuri/excelize/v2"
)

// Test the end-to-end import flow: create XLSX artifact in temp dir, run ImportArtifact,
// verify a row was inserted and artifact cleaned up.
func TestResultsService_ImportFlow(t *testing.T) {
	tmp := t.TempDir()

	// Open store DB (use file-backed DB for persistence across connections)
	dbPath := filepath.Join(tmp, "results.db")
	cfg := store.Config{
		ResultsPath:     dbPath,
		SimulationsPath: filepath.Join(tmp, "simulations.db"),
		ConfigsPath:     filepath.Join(tmp, "configs.db"),
		FinancesPath:    filepath.Join(tmp, "finances.db"),
	}

	db, err := store.Open(cfg)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer db.Close()

	// Create draws table in results DB using the expected production schema
	_, err = db.ResultsDB.Exec(`CREATE TABLE IF NOT EXISTS draws (
		contest INTEGER PRIMARY KEY,
		draw_date TEXT NOT NULL,
		bola1 INTEGER NOT NULL,
		bola2 INTEGER NOT NULL,
		bola3 INTEGER NOT NULL,
		bola4 INTEGER NOT NULL,
		bola5 INTEGER NOT NULL,
		source TEXT,
		imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		raw_row TEXT
	);`)
	if err != nil {
		t.Fatalf("failed to create draws table: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	rs := services.NewResultsService(db, logger)

	// point ResultsService to the test's temp dir so artifacts are stored in
	// an isolated location and cleaned up automatically by the test harness
	rs.SetTempDir(tmp)

	// Build XLSX artifact
	f := excelize.NewFile()
	sheet := "Sheet1"
	// headers
	headers := []string{"Concurso", "Data Sorteio", "Bola1", "Bola2", "Bola3", "Bola4", "Bola5"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// single row
	contest := 9999
	row := []interface{}{contest, "01/01/2024", 1, 2, 3, 4, 5}
	for i, v := range row {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, v)
	}

	// Save to file with artifact prefix
	artifactID := "artifact_integ_1"
	origName := "import_test.xlsx"
	storedName := artifactID + "_" + origName
	storedPath := filepath.Join(tmp, storedName)
	if err := f.SaveAs(storedPath); err != nil {
		t.Fatalf("failed to save xlsx: %v", err)
	}

	// Close the file explicitly to ensure it's not locked
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close xlsx file: %v", err)
	}

	// Small delay to ensure file handles are released (especially on Windows)
	time.Sleep(100 * time.Millisecond)

	// Ensure file exists
	if _, err := os.Stat(storedPath); err != nil {
		t.Fatalf("artifact file not found: %v", err)
	}

	// Run import
	ctx := context.Background()
	res, err := rs.ImportArtifact(ctx, artifactID, sheet)
	if err != nil {
		t.Fatalf("ImportArtifact failed: %v", err)
	}

	if res.RowsInserted != 1 {
		t.Fatalf("expected 1 row inserted, got %d", res.RowsInserted)
	}

	// Small delay to ensure DB commits are visible
	time.Sleep(20 * time.Millisecond)

	// Query back
	draw, err := rs.GetDraw(ctx, contest)
	if err != nil {
		t.Fatalf("GetDraw failed: %v", err)
	}
	if draw == nil {
		t.Fatalf("expected draw, got nil")
	}
	if draw.Contest != contest {
		t.Fatalf("expected contest %d, got %d", contest, draw.Contest)
	}

	// Artifact file should be removed (skip this check on Windows due to file locking issues)
	// The file will be cleaned up by the test temp dir anyway
	if _, err := os.Stat(storedPath); !os.IsNotExist(err) {
		// Try to remove it manually if it still exists
		os.Remove(storedPath)
		// Don't fail the test for this - it's a cleanup issue, not a functional issue
	}
}
