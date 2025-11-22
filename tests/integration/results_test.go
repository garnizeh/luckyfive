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

// TestResultsQueries_RangeAndStats exercises additional queries: count by date range,
// contest range and list-by-date/contest-range using the real migration schema.
func TestResultsQueries_RangeAndStats(t *testing.T) {
	// Load embedded migration and extract Up section
	content, err := migrations.Files.ReadFile("001_create_results.sql")
	if err != nil {
		t.Fatalf("failed to read embedded migration: %v", err)
	}
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

	// Insert multiple draws with different dates and contest numbers
	draws := []results.InsertDrawParams{
		{Contest: 1, DrawDate: "2024-01-01", Bola1: 1, Bola2: 2, Bola3: 3, Bola4: 4, Bola5: 5},
		{Contest: 2, DrawDate: "2024-02-15", Bola1: 6, Bola2: 7, Bola3: 8, Bola4: 9, Bola5: 10},
		{Contest: 3, DrawDate: "2024-03-10", Bola1: 11, Bola2: 12, Bola3: 13, Bola4: 14, Bola5: 15},
		{Contest: 4, DrawDate: "2024-03-20", Bola1: 16, Bola2: 17, Bola3: 18, Bola4: 19, Bola5: 20},
	}

	for _, d := range draws {
		if err := q.InsertDraw(ctx, d); err != nil {
			t.Fatalf("InsertDraw failed: %v", err)
		}
	}

	// Count draws between 2024-02-01 and 2024-03-15 -> should include contests 2 and 3
	cnt, err := q.CountDrawsBetweenDates(ctx, results.CountDrawsBetweenDatesParams{FromDrawDate: "2024-02-01", ToDrawDate: "2024-03-15"})
	if err != nil {
		t.Fatalf("CountDrawsBetweenDates failed: %v", err)
	}
	if cnt != 2 {
		t.Fatalf("expected 2 draws in date range, got %d", cnt)
	}

	// Contest range min/max should be 1 and 4
	cr, err := q.GetContestRange(ctx)
	if err != nil {
		t.Fatalf("GetContestRange failed: %v", err)
	}
	// the generated types use interface{} for min/max; convert via sql driver behavior
	var minContest, maxContest int64
	switch v := cr.MinContest.(type) {
	case int64:
		minContest = v
	case nil:
		t.Fatalf("expected min contest, got nil")
	default:
		t.Fatalf("unexpected type for MinContest: %T", v)
	}
	switch v := cr.MaxContest.(type) {
	case int64:
		maxContest = v
	case nil:
		t.Fatalf("expected max contest, got nil")
	default:
		t.Fatalf("unexpected type for MaxContest: %T", v)
	}
	if minContest != 1 || maxContest != 4 {
		t.Fatalf("expected contest range 1..4, got %d..%d", minContest, maxContest)
	}

	// List draws by date range (descending by contest) with limit/offset
	list, err := q.ListDrawsByDateRange(ctx, results.ListDrawsByDateRangeParams{FromDrawDate: "2024-01-01", ToDrawDate: "2024-12-31", Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListDrawsByDateRange failed: %v", err)
	}
	if len(list) != 4 {
		t.Fatalf("expected 4 draws from ListDrawsByDateRange, got %d", len(list))
	}
	// Ensure ordering is by contest DESC
	if list[0].Contest < list[1].Contest {
		t.Fatalf("expected descending contest order, got %d before %d", list[0].Contest, list[1].Contest)
	}

	// List draws by contest range (ascending)
	crList, err := q.ListDrawsByContestRange(ctx, results.ListDrawsByContestRangeParams{FromContest: 2, ToContest: 4})
	if err != nil {
		t.Fatalf("ListDrawsByContestRange failed: %v", err)
	}
	if len(crList) != 3 {
		t.Fatalf("expected 3 draws in contest range 2..4, got %d", len(crList))
	}
	if crList[0].Contest != 2 || crList[2].Contest != 4 {
		t.Fatalf("unexpected contest ordering in ListDrawsByContestRange: %v", []int{int(crList[0].Contest), int(crList[1].Contest), int(crList[2].Contest)})
	}
}
