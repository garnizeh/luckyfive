package services

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/garnizeh/luckyfive/internal/models"
	"github.com/garnizeh/luckyfive/internal/store/results"
)

// DBInterface defines the database operations needed by ImportService
type DBInterface interface {
	WithResultsTx(ctx context.Context, fn func(results.Querier) error) error
}

// ImportService handles importing lottery draw data from various sources
type ImportService struct {
	db        DBInterface // Database access
	logger    *slog.Logger
	artifacts map[string][]byte // Temporary storage for uploaded artifacts
	nextID    int               // Simple ID generator
}

// NewImportService creates a new import service
func NewImportService(db DBInterface, logger *slog.Logger) *ImportService {
	return &ImportService{
		db:        db,
		logger:    logger,
		artifacts: make(map[string][]byte),
		nextID:    1,
	}
}

// ParseXLSX parses an XLSX file and returns a slice of Draw structs
func (s *ImportService) ParseXLSX(reader io.Reader, sheet string) ([]models.Draw, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open XLSX file: %w", err)
	}
	defer f.Close()

	if sheet == "" {
		sheet = f.GetSheetList()[0] // Use first sheet if not specified
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet %s: %w", sheet, err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("XLSX file must have at least a header row and one data row")
	}

	// Detect column mapping from headers
	headers := rows[0]
	mapping, err := models.DetectColumns(headers)
	if err != nil {
		return nil, fmt.Errorf("failed to detect columns: %w", err)
	}

	s.logger.Info("Detected column mapping",
		"contest_col", mapping.ContestCol,
		"date_col", mapping.DateCol,
		"bola_cols", mapping.BolaCols)

	var draws []models.Draw
	source := fmt.Sprintf("xlsx:%s", sheet)

	// Process data rows
	for i, row := range rows[1:] {
		if len(row) == 0 {
			continue // Skip empty rows
		}

		draw, err := s.parseRow(row, mapping)
		if err != nil {
			s.logger.Warn("Skipping invalid row", "row", i+2, "error", err)
			continue
		}

		draw.Source = source
		draw.ImportedAt = time.Now()
		draw.RawRow = strings.Join(row, "\t")

		// Validate the draw
		if err := draw.Validate(); err != nil {
			s.logger.Warn("Skipping invalid draw", "row", i+2, "error", err)
			continue
		}

		draws = append(draws, draw)
	}

	s.logger.Info("Parsed draws from XLSX", "count", len(draws))
	return draws, nil
}

// parseRow parses a single row into a Draw struct
func (s *ImportService) parseRow(row []string, mapping models.ColumnMapping) (models.Draw, error) {
	draw := models.Draw{}

	// Parse contest number
	if mapping.ContestCol >= len(row) {
		return draw, fmt.Errorf("contest column %d out of range for row with %d columns", mapping.ContestCol, len(row))
	}
	contestStr := strings.TrimSpace(row[mapping.ContestCol])
	contest, err := strconv.Atoi(contestStr)
	if err != nil {
		return draw, fmt.Errorf("invalid contest number '%s': %w", contestStr, err)
	}
	draw.Contest = contest

	// Parse draw date
	if mapping.DateCol >= len(row) {
		return draw, fmt.Errorf("date column %d out of range for row with %d columns", mapping.DateCol, len(row))
	}
	dateStr := strings.TrimSpace(row[mapping.DateCol])
	drawDate, err := models.ParseDrawDate(dateStr)
	if err != nil {
		return draw, fmt.Errorf("invalid draw date '%s': %w", dateStr, err)
	}
	draw.DrawDate = drawDate

	// Parse ball numbers
	var balls []int
	for i, col := range mapping.BolaCols {
		if col >= len(row) {
			return draw, fmt.Errorf("bola%d column %d out of range for row with %d columns", i+1, col, len(row))
		}
		ballStr := strings.TrimSpace(row[col])
		ball, err := models.ParseBallNumber(ballStr)
		if err != nil {
			return draw, fmt.Errorf("invalid bola%d '%s': %w", i+1, ballStr, err)
		}
		balls = append(balls, ball)
	}

	// Sort balls in ascending order
	sort.Ints(balls)
	if len(balls) != 5 {
		return draw, fmt.Errorf("expected 5 balls, got %d", len(balls))
	}

	draw.Bola1 = balls[0]
	draw.Bola2 = balls[1]
	draw.Bola3 = balls[2]
	draw.Bola4 = balls[3]
	draw.Bola5 = balls[4]

	return draw, nil
}

// ImportDraws imports the parsed draws into the database using batched transactions
func (s *ImportService) ImportDraws(ctx context.Context, draws []models.Draw) error {
	if len(draws) == 0 {
		return fmt.Errorf("no draws to import")
	}

	s.logger.Info("Starting import", "count", len(draws))

	const batchSize = 100
	totalImported := 0
	totalSkipped := 0

	// Process draws in batches
	for i := 0; i < len(draws); i += batchSize {
		end := min(i+batchSize, len(draws))
		batch := draws[i:end]

		s.logger.Info("Processing batch", "batch_start", i, "batch_end", end-1, "batch_size", len(batch))

		// Import batch in a transaction
		imported, skipped, err := s.importBatch(ctx, batch)
		if err != nil {
			return fmt.Errorf("failed to import batch %d-%d: %w", i, end-1, err)
		}

		totalImported += imported
		totalSkipped += skipped

		s.logger.Info("Batch completed", "imported", imported, "skipped", skipped)
	}

	s.logger.Info("Import completed", "total_imported", totalImported, "total_skipped", totalSkipped)
	if totalSkipped > 0 {
		return fmt.Errorf("import completed with %d errors out of %d draws", totalSkipped, len(draws))
	}

	return nil
}

// importBatch imports a batch of draws within a single transaction
func (s *ImportService) importBatch(ctx context.Context, draws []models.Draw) (imported int, skipped int, err error) {
	if s.db == nil {
		// For testing without DB
		return len(draws), 0, nil
	}

	err = s.db.WithResultsTx(ctx, func(q results.Querier) error {
		for _, draw := range draws {
			params := results.InsertDrawParams{
				Contest:  int64(draw.Contest),
				DrawDate: draw.DrawDate.Format("2006-01-02"), // Convert to YYYY-MM-DD format
				Bola1:    int64(draw.Bola1),
				Bola2:    int64(draw.Bola2),
				Bola3:    int64(draw.Bola3),
				Bola4:    int64(draw.Bola4),
				Bola5:    int64(draw.Bola5),
				Source:   sql.NullString{String: draw.Source, Valid: draw.Source != ""},
				RawRow:   sql.NullString{String: draw.RawRow, Valid: draw.RawRow != ""},
			}

			err := q.InsertDraw(ctx, params)
			if err != nil {
				s.logger.Warn("Failed to insert draw", "contest", draw.Contest, "error", err)
				skipped++
				continue
			}
			imported++
		}
		return nil
	})
	return imported, skipped, err
}

// SaveArtifact saves an uploaded XLSX file temporarily and returns an artifact ID
func (s *ImportService) SaveArtifact(reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read artifact data: %w", err)
	}

	// Generate simple ID
	id := fmt.Sprintf("artifact_%d", s.nextID)
	s.nextID++

	s.artifacts[id] = data
	s.logger.Info("Saved artifact", "id", id, "size", len(data))

	return id, nil
}

// ImportArtifact imports data from a previously saved artifact
func (s *ImportService) ImportArtifact(ctx context.Context, artifactID string, sheet string) error {
	data, exists := s.artifacts[artifactID]
	if !exists {
		return fmt.Errorf("artifact %s not found", artifactID)
	}

	reader := strings.NewReader(string(data))
	draws, err := s.ParseXLSX(reader, sheet)
	if err != nil {
		return fmt.Errorf("failed to parse XLSX from artifact: %w", err)
	}

	err = s.ImportDraws(ctx, draws)
	if err != nil {
		return fmt.Errorf("failed to import draws: %w", err)
	}

	// Clean up artifact after successful import
	delete(s.artifacts, artifactID)
	s.logger.Info("Imported artifact", "id", artifactID, "draws_count", len(draws))

	return nil
}

// GetDraw retrieves a single draw by contest number
func (s *ImportService) GetDraw(ctx context.Context, contest int) (*models.Draw, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var draw *models.Draw
	err := s.db.WithResultsTx(ctx, func(q results.Querier) error {
		result, err := q.GetDraw(ctx, int64(contest))
		if err != nil {
			return err
		}

		// Convert sqlc result to models.Draw
		drawDate, err := time.Parse("2006-01-02", result.DrawDate)
		if err != nil {
			return fmt.Errorf("invalid draw date format: %w", err)
		}

		draw = &models.Draw{
			Contest:    int(result.Contest),
			DrawDate:   drawDate,
			Bola1:      int(result.Bola1),
			Bola2:      int(result.Bola2),
			Bola3:      int(result.Bola3),
			Bola4:      int(result.Bola4),
			Bola5:      int(result.Bola5),
			Source:     result.Source.String,
			ImportedAt: time.Now(), // Not stored in DB, use current time
			RawRow:     result.RawRow.String,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return draw, nil
}

// ListDraws retrieves draws with pagination
func (s *ImportService) ListDraws(ctx context.Context, limit, offset int) ([]models.Draw, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var draws []models.Draw
	err := s.db.WithResultsTx(ctx, func(q results.Querier) error {
		params := results.ListDrawsParams{
			Limit:  int64(limit),
			Offset: int64(offset),
		}
		results, err := q.ListDraws(ctx, params)
		if err != nil {
			return err
		}

		for _, r := range results {
			drawDate, err := time.Parse("2006-01-02", r.DrawDate)
			if err != nil {
				s.logger.Warn("Invalid draw date format", "contest", r.Contest, "error", err)
				continue
			}

			draw := models.Draw{
				Contest:    int(r.Contest),
				DrawDate:   drawDate,
				Bola1:      int(r.Bola1),
				Bola2:      int(r.Bola2),
				Bola3:      int(r.Bola3),
				Bola4:      int(r.Bola4),
				Bola5:      int(r.Bola5),
				Source:     r.Source.String,
				ImportedAt: time.Now(), // Not stored, use current
				RawRow:     r.RawRow.String,
			}
			draws = append(draws, draw)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return draws, nil
}

// ValidateDrawData validates draw data (wrapper around models.Draw.Validate)
func (s *ImportService) ValidateDrawData(draw *models.Draw) error {
	return draw.Validate()
}
