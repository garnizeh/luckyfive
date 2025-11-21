package integration

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/store/results"
	"github.com/garnizeh/luckyfive/migrations"
)

// TestResultsQueries_Integration verifies basic CRUD operations using the sqlc-generated
// queries against an in-memory SQLite database with the real migration schema applied.
func TestResultsQueries_Integration(t *testing.T) {
	// Read migration SQL and apply the Up section to an in-memory DB
	content, err := migrations.Files.ReadFile("001_create_results.sql")
	if err != nil {
		t.Fatalf("failed to read embedded migration: %v", err)
	}

	// Extract up section (before any -- Down marker)
	sqlText := string(content)
	upSQL := sqlText
	if idx := indexOfDown(sqlText); idx >= 0 {
		upSQL = sqlText[:idx]
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(upSQL); err != nil {
		t.Fatalf("failed to execute migration SQL: %v", err)
	}

	q := results.New(db)
	ctx := context.Background()

	// Ensure initially empty
	cnt, err := q.CountDraws(ctx)
	if err != nil {
		t.Fatalf("CountDraws failed: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("expected 0 draws, got %d", cnt)
	}

	// Insert a draw
	err = q.InsertDraw(ctx, results.InsertDrawParams{
		Contest:  1,
		DrawDate: "2024-01-01",
		Bola1:    5,
		Bola2:    10,
		Bola3:    15,
		Bola4:    20,
		Bola5:    25,
		Source:   sql.NullString{String: "test", Valid: true},
		RawRow:   sql.NullString{String: "raw", Valid: true},
	})
	if err != nil {
		t.Fatalf("InsertDraw failed: %v", err)
	}

	// Count should be 1
	cnt, err = q.CountDraws(ctx)
	if err != nil {
		t.Fatalf("CountDraws failed: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 draw, got %d", cnt)
	}

	// GetDraw
	d, err := q.GetDraw(ctx, 1)
	if err != nil {
		t.Fatalf("GetDraw failed: %v", err)
	}
	if d.Contest != 1 {
		t.Fatalf("expected contest 1, got %d", d.Contest)
	}

	// ListDraws
	list, err := q.ListDraws(ctx, results.ListDrawsParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListDraws failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 draw from ListDraws, got %d", len(list))
	}

	// UpsertDraw should update without error
	err = q.UpsertDraw(ctx, results.UpsertDrawParams{
		Contest:  1,
		DrawDate: "2024-01-01",
		Bola1:    6, // changed
		Bola2:    10,
		Bola3:    15,
		Bola4:    20,
		Bola5:    25,
		Source:   sql.NullString{String: "test2", Valid: true},
		RawRow:   sql.NullString{String: "raw2", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpsertDraw failed: %v", err)
	}

	d2, err := q.GetDraw(ctx, 1)
	if err != nil {
		t.Fatalf("GetDraw after upsert failed: %v", err)
	}
	if d2.Bola1 != 6 {
		t.Fatalf("expected Bola1 to be updated to 6, got %d", d2.Bola1)
	}

	// Insert import history
	ih, err := q.InsertImportHistory(ctx, results.InsertImportHistoryParams{
		Filename:     "file.xlsx",
		RowsInserted: 1,
		RowsSkipped:  0,
		RowsErrors:   0,
		SourceHash:   sql.NullString{String: "hash", Valid: true},
		Metadata:     sql.NullString{String: "meta", Valid: true},
	})
	if err != nil {
		t.Fatalf("InsertImportHistory failed: %v", err)
	}
	if ih.Filename != "file.xlsx" {
		t.Fatalf("expected import history filename file.xlsx, got %s", ih.Filename)
	}

	// GetImportHistory
	hist, err := q.GetImportHistory(ctx, results.GetImportHistoryParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("GetImportHistory failed: %v", err)
	}
	if len(hist) != 1 {
		t.Fatalf("expected 1 import history row, got %d", len(hist))
	}
}

// indexOfDown returns the byte index of the first occurrence of a line starting with "-- Down" (case-insensitive)
func indexOfDown(s string) int {
	// simple approach: look for "-- Down" in lowercased text
	lower := strings.ToLower(s)
	idx := strings.Index(lower, "-- down")
	return idx
}
