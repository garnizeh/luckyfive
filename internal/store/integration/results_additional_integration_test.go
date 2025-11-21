package integration

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/garnizeh/luckyfive/internal/store/results"
	"github.com/garnizeh/luckyfive/migrations"
)

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
