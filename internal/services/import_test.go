package services

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/store/results"
)

func TestImportService_ParseXLSX(t *testing.T) {
	// Create a test XLSX file
	f := excelize.NewFile()
	defer f.Close()

	// Create a sheet with test data
	sheetName := "Sheet1"
	f.SetSheetName("Sheet1", sheetName)

	// Set headers
	headers := []string{"Concurso", "Data Sorteio", "Bola1", "Bola2", "Bola3", "Bola4", "Bola5"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Set test data
	testData := [][]interface{}{
		{1, "02/01/2024", 5, 12, 23, 34, 45},
		{2, "09/01/2024", 3, 15, 27, 38, 49},
		{3, "16/01/2024", 7, 18, 29, 41, 52},
	}

	for rowIdx, row := range testData {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("Failed to write XLSX to buffer: %v", err)
	}

	// Create service
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(nil, logger)

	// Parse XLSX
	reader := bytes.NewReader(buf.Bytes())
	draws, err := service.ParseXLSX(reader, sheetName)
	if err != nil {
		t.Fatalf("ParseXLSX failed: %v", err)
	}

	// Verify results
	expected := []models.Draw{
		{
			Contest:  1,
			DrawDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			Bola1:    5,
			Bola2:    12,
			Bola3:    23,
			Bola4:    34,
			Bola5:    45,
			Source:   "xlsx:Sheet1",
		},
		{
			Contest:  2,
			DrawDate: time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC),
			Bola1:    3,
			Bola2:    15,
			Bola3:    27,
			Bola4:    38,
			Bola5:    49,
			Source:   "xlsx:Sheet1",
		},
		{
			Contest:  3,
			DrawDate: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			Bola1:    7,
			Bola2:    18,
			Bola3:    29,
			Bola4:    41,
			Bola5:    52,
			Source:   "xlsx:Sheet1",
		},
	}

	if len(draws) != len(expected) {
		t.Fatalf("Expected %d draws, got %d", len(expected), len(draws))
	}

	for i, draw := range draws {
		exp := expected[i]

		if draw.Contest != exp.Contest {
			t.Errorf("Draw %d: expected contest %d, got %d", i, exp.Contest, draw.Contest)
		}

		if !draw.DrawDate.Equal(exp.DrawDate) {
			t.Errorf("Draw %d: expected date %v, got %v", i, exp.DrawDate, draw.DrawDate)
		}

		if draw.Bola1 != exp.Bola1 || draw.Bola2 != exp.Bola2 || draw.Bola3 != exp.Bola3 ||
			draw.Bola4 != exp.Bola4 || draw.Bola5 != exp.Bola5 {
			t.Errorf("Draw %d: expected balls [%d,%d,%d,%d,%d], got [%d,%d,%d,%d,%d]",
				i, exp.Bola1, exp.Bola2, exp.Bola3, exp.Bola4, exp.Bola5,
				draw.Bola1, draw.Bola2, draw.Bola3, draw.Bola4, draw.Bola5)
		}

		if draw.Source != exp.Source {
			t.Errorf("Draw %d: expected source %s, got %s", i, exp.Source, draw.Source)
		}

		if err := draw.Validate(); err != nil {
			t.Errorf("Draw %d validation failed: %v", i, err)
		}
	}
}

func TestImportService_ParseXLSX_ColumnDetection(t *testing.T) {
	// Test with different column layouts
	testCases := []struct {
		name     string
		headers  []string
		data     []interface{}
		expected models.Draw
	}{
		{
			name:    "Portuguese headers",
			headers: []string{"Concurso", "Data Sorteio", "Bola1", "Bola2", "Bola3", "Bola4", "Bola5"},
			data:    []interface{}{100, "15/03/2024", 10, 20, 30, 40, 50},
			expected: models.Draw{
				Contest:  100,
				DrawDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
				Bola1:    10, Bola2: 20, Bola3: 30, Bola4: 40, Bola5: 50,
			},
		},
		{
			name:    "English headers",
			headers: []string{"Contest", "Draw Date", "Ball1", "Ball2", "Ball3", "Ball4", "Ball5"},
			data:    []interface{}{101, "2024-03-22", 5, 15, 25, 35, 45},
			expected: models.Draw{
				Contest:  101,
				DrawDate: time.Date(2024, 3, 22, 0, 0, 0, 0, time.UTC),
				Bola1:    5, Bola2: 15, Bola3: 25, Bola4: 35, Bola5: 45,
			},
		},
		{
			name:    "Position-based detection",
			headers: []string{"ID", "Date", "N1", "N2", "N3", "N4", "N5"},
			data:    []interface{}{102, "30/03/2024", 2, 12, 22, 32, 42},
			expected: models.Draw{
				Contest:  102,
				DrawDate: time.Date(2024, 3, 30, 0, 0, 0, 0, time.UTC),
				Bola1:    2, Bola2: 12, Bola3: 22, Bola4: 32, Bola5: 42,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := excelize.NewFile()
			defer f.Close()

			sheetName := "Sheet1"

			// Set headers
			for i, header := range tc.headers {
				cell, _ := excelize.CoordinatesToCellName(i+1, 1)
				f.SetCellValue(sheetName, cell, header)
			}

			// Set data
			for i, value := range tc.data {
				cell, _ := excelize.CoordinatesToCellName(i+1, 2)
				f.SetCellValue(sheetName, cell, value)
			}

			var buf bytes.Buffer
			if err := f.Write(&buf); err != nil {
				t.Fatalf("Failed to write XLSX: %v", err)
			}

			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
			service := NewImportService(nil, logger)

			reader := bytes.NewReader(buf.Bytes())
			draws, err := service.ParseXLSX(reader, sheetName)
			if err != nil {
				t.Fatalf("ParseXLSX failed: %v", err)
			}

			if len(draws) != 1 {
				t.Fatalf("Expected 1 draw, got %d", len(draws))
			}

			draw := draws[0]
			if draw.Contest != tc.expected.Contest {
				t.Errorf("Expected contest %d, got %d", tc.expected.Contest, draw.Contest)
			}

			if !draw.DrawDate.Equal(tc.expected.DrawDate) {
				t.Errorf("Expected date %v, got %v", tc.expected.DrawDate, draw.DrawDate)
			}

			if draw.Bola1 != tc.expected.Bola1 || draw.Bola2 != tc.expected.Bola2 ||
				draw.Bola3 != tc.expected.Bola3 || draw.Bola4 != tc.expected.Bola4 ||
				draw.Bola5 != tc.expected.Bola5 {
				t.Errorf("Expected balls [%d,%d,%d,%d,%d], got [%d,%d,%d,%d,%d]",
					tc.expected.Bola1, tc.expected.Bola2, tc.expected.Bola3, tc.expected.Bola4, tc.expected.Bola5,
					draw.Bola1, draw.Bola2, draw.Bola3, draw.Bola4, draw.Bola5)
			}
		})
	}
}

func TestImportService_ParseXLSX_InvalidData(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// Set headers
	headers := []string{"Concurso", "Data Sorteio", "Bola1", "Bola2", "Bola3", "Bola4", "Bola5"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Set invalid data
	invalidData := [][]interface{}{
		{"invalid", "02/01/2024", 5, 12, 23, 34, 45}, // Invalid contest
		{1, "invalid date", 5, 12, 23, 34, 45},       // Invalid date
		{2, "02/01/2024", 0, 12, 23, 34, 45},         // Ball out of range
		{3, "02/01/2024", 5, 5, 23, 34, 45},          // Duplicate balls
		{4, "02/01/2024", 10, 20, 30, 40, 50},        // Valid - should be parsed
	}

	for rowIdx, row := range invalidData {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("Failed to write XLSX: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(nil, logger)

	reader := bytes.NewReader(buf.Bytes())
	draws, err := service.ParseXLSX(reader, sheetName)
	if err != nil {
		t.Fatalf("ParseXLSX failed: %v", err)
	}

	// Should only have the valid draw (contest 4)
	if len(draws) != 1 {
		t.Fatalf("Expected 1 valid draw, got %d", len(draws))
	}

	draw := draws[0]
	if draw.Contest != 4 {
		t.Errorf("Expected contest 4, got %d", draw.Contest)
	}
}

// Mock Querier for testing ImportDraws
type mockQuerier struct {
	inserted        []results.InsertDrawParams
	getDrawResult   results.Draw
	getDrawError    error
	listDrawsResult []results.Draw
	listDrawsError  error
}

func (m *mockQuerier) InsertDraw(ctx context.Context, arg results.InsertDrawParams) error {
	m.inserted = append(m.inserted, arg)
	return nil
}

// Implement other interface methods (not used in tests)
func (m *mockQuerier) CountDraws(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockQuerier) CountDrawsBetweenDates(ctx context.Context, arg results.CountDrawsBetweenDatesParams) (int64, error) {
	return 0, nil
}
func (m *mockQuerier) DeleteDraw(ctx context.Context, contest int64) error { return nil }
func (m *mockQuerier) GetContestRange(ctx context.Context) (results.GetContestRangeRow, error) {
	return results.GetContestRangeRow{}, nil
}
func (m *mockQuerier) GetDraw(ctx context.Context, contest int64) (results.Draw, error) {
	return m.getDrawResult, m.getDrawError
}
func (m *mockQuerier) GetDrawByDate(ctx context.Context, drawDate string) ([]results.Draw, error) {
	return nil, nil
}
func (m *mockQuerier) GetImportHistory(ctx context.Context, arg results.GetImportHistoryParams) ([]results.ImportHistory, error) {
	return nil, nil
}
func (m *mockQuerier) InsertImportHistory(ctx context.Context, arg results.InsertImportHistoryParams) (results.ImportHistory, error) {
	return results.ImportHistory{}, nil
}
func (m *mockQuerier) ListDraws(ctx context.Context, arg results.ListDrawsParams) ([]results.Draw, error) {
	return m.listDrawsResult, m.listDrawsError
}
func (m *mockQuerier) ListDrawsByBall(ctx context.Context, arg results.ListDrawsByBallParams) ([]results.Draw, error) {
	return nil, nil
}
func (m *mockQuerier) ListDrawsByContestRange(ctx context.Context, arg results.ListDrawsByContestRangeParams) ([]results.Draw, error) {
	return nil, nil
}
func (m *mockQuerier) ListDrawsByDateRange(ctx context.Context, arg results.ListDrawsByDateRangeParams) ([]results.Draw, error) {
	return nil, nil
}
func (m *mockQuerier) UpsertDraw(ctx context.Context, arg results.UpsertDrawParams) error { return nil }

// Mock DB for testing
type mockDB struct {
	querier *mockQuerier
}

func (m *mockDB) WithResultsTx(ctx context.Context, fn func(results.Querier) error) error {
	return fn(m.querier)
}

func TestImportService_ImportDraws(t *testing.T) {
	mockQ := &mockQuerier{}
	mockDB := &mockDB{querier: mockQ}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(mockDB, logger)

	testDraws := []models.Draw{
		{
			Contest:  1,
			DrawDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Bola1:    5, Bola2: 10, Bola3: 15, Bola4: 20, Bola5: 25,
			Source: "test",
			RawRow: "test data",
		},
		{
			Contest:  2,
			DrawDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			Bola1:    1, Bola2: 2, Bola3: 3, Bola4: 4, Bola5: 5,
			Source: "test2",
			RawRow: "test data 2",
		},
	}

	err := service.ImportDraws(context.Background(), testDraws)
	if err != nil {
		t.Fatalf("ImportDraws failed: %v", err)
	}

	if len(mockQ.inserted) != 2 {
		t.Fatalf("Expected 2 inserts, got %d", len(mockQ.inserted))
	}

	// Check first insert
	params := mockQ.inserted[0]
	if params.Contest != 1 {
		t.Errorf("Expected contest 1, got %d", params.Contest)
	}
	if params.DrawDate != "2024-01-01" {
		t.Errorf("Expected date 2024-01-01, got %s", params.DrawDate)
	}
	if params.Bola1 != 5 || params.Bola2 != 10 || params.Bola3 != 15 || params.Bola4 != 20 || params.Bola5 != 25 {
		t.Errorf("Unexpected ball values: %d,%d,%d,%d,%d", params.Bola1, params.Bola2, params.Bola3, params.Bola4, params.Bola5)
	}
	if !params.Source.Valid || params.Source.String != "test" {
		t.Errorf("Expected source 'test', got %v", params.Source)
	}
	if !params.RawRow.Valid || params.RawRow.String != "test data" {
		t.Errorf("Expected raw row 'test data', got %v", params.RawRow)
	}
}

func TestImportService_SaveArtifact(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(nil, logger)

	testData := "test XLSX content"
	reader := bytes.NewReader([]byte(testData))

	artifactID, err := service.SaveArtifact(reader)
	if err != nil {
		t.Fatalf("SaveArtifact failed: %v", err)
	}

	if artifactID == "" {
		t.Error("Expected non-empty artifact ID")
	}

	// Check that artifact was stored
	if len(service.artifacts) != 1 {
		t.Errorf("Expected 1 artifact stored, got %d", len(service.artifacts))
	}

	if storedData, exists := service.artifacts[artifactID]; !exists {
		t.Error("Artifact not found in storage")
	} else if string(storedData) != testData {
		t.Errorf("Stored data mismatch: expected %s, got %s", testData, string(storedData))
	}
}

func TestImportService_SaveArtifact_Error(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(nil, logger)

	// Create a reader that fails
	failingReader := &failingReader{}

	_, err := service.SaveArtifact(failingReader)
	if err == nil {
		t.Error("Expected error from SaveArtifact")
	}
}

type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, os.ErrClosed
}

func TestImportService_ImportArtifact(t *testing.T) {
	mockQ := &mockQuerier{}
	mockDB := &mockDB{querier: mockQ}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(mockDB, logger)

	// Create test XLSX data
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	headers := []string{"Concurso", "Data Sorteio", "Bola1", "Bola2", "Bola3", "Bola4", "Bola5"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	testData := []interface{}{1, "02/01/2024", 5, 12, 23, 34, 45}
	for i, value := range testData {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheetName, cell, value)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("Failed to write XLSX: %v", err)
	}

	// Save artifact
	reader := bytes.NewReader(buf.Bytes())
	artifactID, err := service.SaveArtifact(reader)
	if err != nil {
		t.Fatalf("SaveArtifact failed: %v", err)
	}

	// Import artifact
	err = service.ImportArtifact(context.Background(), artifactID, sheetName)
	if err != nil {
		t.Fatalf("ImportArtifact failed: %v", err)
	}

	// Check that draw was inserted
	if len(mockQ.inserted) != 1 {
		t.Fatalf("Expected 1 insert, got %d", len(mockQ.inserted))
	}

	// Check that artifact was cleaned up
	if len(service.artifacts) != 0 {
		t.Errorf("Expected artifact to be cleaned up, got %d artifacts", len(service.artifacts))
	}
}

func TestImportService_ImportArtifact_NotFound(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(nil, logger)

	err := service.ImportArtifact(context.Background(), "nonexistent", "Sheet1")
	if err == nil {
		t.Error("Expected error for nonexistent artifact")
	}
}

func TestImportService_GetDraw(t *testing.T) {
	mockQ := &mockQuerier{
		getDrawResult: results.Draw{
			Contest:  100,
			DrawDate: "2024-01-01",
			Bola1:    5,
			Bola2:    10,
			Bola3:    15,
			Bola4:    20,
			Bola5:    25,
			Source:   sql.NullString{String: "test", Valid: true},
			RawRow:   sql.NullString{String: "raw data", Valid: true},
		},
	}
	mockDB := &mockDB{querier: mockQ}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(mockDB, logger)

	draw, err := service.GetDraw(context.Background(), 100)
	if err != nil {
		t.Fatalf("GetDraw failed: %v", err)
	}

	if draw == nil {
		t.Fatal("Expected draw, got nil")
	}

	if draw.Contest != 100 {
		t.Errorf("Expected contest 100, got %d", draw.Contest)
	}

	expectedDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !draw.DrawDate.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, draw.DrawDate)
	}

	if draw.Bola1 != 5 || draw.Bola2 != 10 || draw.Bola3 != 15 || draw.Bola4 != 20 || draw.Bola5 != 25 {
		t.Errorf("Unexpected ball values: %d,%d,%d,%d,%d", draw.Bola1, draw.Bola2, draw.Bola3, draw.Bola4, draw.Bola5)
	}

	if draw.Source != "test" {
		t.Errorf("Expected source 'test', got %s", draw.Source)
	}

	if draw.RawRow != "raw data" {
		t.Errorf("Expected raw row 'raw data', got %s", draw.RawRow)
	}
}

func TestImportService_GetDraw_Error(t *testing.T) {
	mockQ := &mockQuerier{
		getDrawError: sql.ErrNoRows,
	}
	mockDB := &mockDB{querier: mockQ}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(mockDB, logger)

	_, err := service.GetDraw(context.Background(), 999)
	if err == nil {
		t.Error("Expected error from GetDraw")
	}
}

func TestImportService_GetDraw_NoDB(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(nil, logger)

	_, err := service.GetDraw(context.Background(), 100)
	if err == nil {
		t.Error("Expected error when no DB available")
	}
}

func TestImportService_ListDraws(t *testing.T) {
	mockQ := &mockQuerier{
		listDrawsResult: []results.Draw{
			{
				Contest:  100,
				DrawDate: "2024-01-01",
				Bola1:    5, Bola2: 10, Bola3: 15, Bola4: 20, Bola5: 25,
				Source: sql.NullString{String: "test1", Valid: true},
				RawRow: sql.NullString{String: "raw1", Valid: true},
			},
			{
				Contest:  101,
				DrawDate: "2024-01-02",
				Bola1:    1, Bola2: 2, Bola3: 3, Bola4: 4, Bola5: 5,
				Source: sql.NullString{String: "test2", Valid: true},
				RawRow: sql.NullString{String: "raw2", Valid: true},
			},
		},
	}
	mockDB := &mockDB{querier: mockQ}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(mockDB, logger)

	draws, err := service.ListDraws(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("ListDraws failed: %v", err)
	}

	if len(draws) != 2 {
		t.Fatalf("Expected 2 draws, got %d", len(draws))
	}

	// Check first draw
	draw := draws[0]
	if draw.Contest != 100 {
		t.Errorf("Expected contest 100, got %d", draw.Contest)
	}

	expectedDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !draw.DrawDate.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, draw.DrawDate)
	}

	if draw.Source != "test1" {
		t.Errorf("Expected source 'test1', got %s", draw.Source)
	}
}

func TestImportService_ListDraws_Error(t *testing.T) {
	mockQ := &mockQuerier{
		listDrawsError: sql.ErrNoRows,
	}
	mockDB := &mockDB{querier: mockQ}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(mockDB, logger)

	_, err := service.ListDraws(context.Background(), 10, 0)
	if err == nil {
		t.Error("Expected error from ListDraws")
	}
}

func TestImportService_ListDraws_NoDB(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(nil, logger)

	_, err := service.ListDraws(context.Background(), 10, 0)
	if err == nil {
		t.Error("Expected error when no DB available")
	}
}

func TestImportService_ValidateDrawData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewImportService(nil, logger)

	validDraw := &models.Draw{
		Contest:  100,
		DrawDate: time.Now(),
		Bola1:    5, Bola2: 10, Bola3: 15, Bola4: 20, Bola5: 25,
	}

	err := service.ValidateDrawData(validDraw)
	if err != nil {
		t.Errorf("Expected valid draw to pass validation, got error: %v", err)
	}

	invalidDraw := &models.Draw{
		Contest:  0, // Invalid contest
		DrawDate: time.Now(),
		Bola1:    5, Bola2: 10, Bola3: 15, Bola4: 20, Bola5: 25,
	}

	err = service.ValidateDrawData(invalidDraw)
	if err == nil {
		t.Error("Expected invalid draw to fail validation")
	}
}
