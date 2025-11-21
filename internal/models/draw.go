package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Draw represents a lottery draw result
type Draw struct {
	Contest    int       `json:"contest"`
	DrawDate   time.Time `json:"draw_date"`
	Bola1      int       `json:"bola1"`
	Bola2      int       `json:"bola2"`
	Bola3      int       `json:"bola3"`
	Bola4      int       `json:"bola4"`
	Bola5      int       `json:"bola5"`
	Source     string    `json:"source"`
	ImportedAt time.Time `json:"imported_at"`
	RawRow     string    `json:"raw_row"`
}

// Validate checks if the draw data is valid
func (d *Draw) Validate() error {
	if d.Contest <= 0 {
		return errors.New("contest must be positive")
	}

	// Check balls are between 1 and 80
	balls := []int{d.Bola1, d.Bola2, d.Bola3, d.Bola4, d.Bola5}
	for i, ball := range balls {
		if ball < 1 || ball > 80 {
			return fmt.Errorf("bola%d must be between 1 and 80, got %d", i+1, ball)
		}
	}

	// Check for duplicates first
	seen := make(map[int]bool)
	for _, ball := range balls {
		if seen[ball] {
			return fmt.Errorf("duplicate ball number: %d", ball)
		}
		seen[ball] = true
	}

	// Check balls are in ascending order
	if d.Bola1 >= d.Bola2 || d.Bola2 >= d.Bola3 || d.Bola3 >= d.Bola4 || d.Bola4 >= d.Bola5 {
		return errors.New("balls must be in ascending order")
	}

	if d.DrawDate.IsZero() {
		return errors.New("draw date cannot be zero")
	}

	return nil
}

// ParseDrawDate attempts to parse a date string in various formats
func ParseDrawDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, errors.New("empty date string")
	}

	// Common Brazilian date formats
	formats := []string{
		"02/01/2006", // DD/MM/YYYY
		"02-01-2006", // DD-MM-YYYY
		"2006/01/02", // YYYY/MM/DD
		"2006-01-02", // YYYY-MM-DD
		"02/01/06",   // DD/MM/YY
		"02-01-06",   // DD-MM-YY
		"2/1/2006",   // D/M/YYYY
		"2-1-2006",   // D-M-YYYY
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ParseBallNumber attempts to parse a ball number from string
func ParseBallNumber(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty ball number")
	}

	// Remove leading zeros
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return 0, nil // was just "0" or "00", etc.
	}

	num, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid ball number: %s", s)
	}

	return num, nil
}

// ColumnMapping represents detected column positions
type ColumnMapping struct {
	ContestCol int
	DateCol    int
	BolaCols   [5]int // indices for bola1, bola2, bola3, bola4, bola5
}

// DetectColumns attempts to detect column positions based on header names
func DetectColumns(headers []string) (ColumnMapping, error) {
	mapping := ColumnMapping{
		ContestCol: -1,
		DateCol:    -1,
		BolaCols:   [5]int{-1, -1, -1, -1, -1},
	}

	// Convert headers to lowercase for case-insensitive matching
	lowerHeaders := make([]string, len(headers))
	for i, h := range headers {
		lowerHeaders[i] = strings.ToLower(strings.TrimSpace(h))
	}

	// Find contest column
	contestNames := []string{"concurso", "contest", "num", "numero", "nº", "no", "id"}
	for _, name := range contestNames {
		for i, header := range lowerHeaders {
			if strings.Contains(header, name) {
				mapping.ContestCol = i
				break
			}
		}
		if mapping.ContestCol != -1 {
			break
		}
	}

	// Find date column
	dateNames := []string{"data", "date", "sorteio", "draw"}
	for _, name := range dateNames {
		for i, header := range lowerHeaders {
			if strings.Contains(header, name) {
				mapping.DateCol = i
				break
			}
		}
		if mapping.DateCol != -1 {
			break
		}
	}

	// If contest/date columns not found, assume position-based (first column = contest, second = date)
	if mapping.ContestCol == -1 && len(headers) > 0 {
		mapping.ContestCol = 0
	}
	if mapping.DateCol == -1 && len(headers) > 1 {
		mapping.DateCol = 1
	}

	// Find ball columns
	ballNames := []string{"bola", "ball", "dezena", "numero"}
	for i, header := range lowerHeaders {
		for _, name := range ballNames {
			if strings.Contains(header, name) {
				// Try to extract number from header (bola1, bola 1, n1, etc.)
				if num := extractNumber(header); num >= 1 && num <= 5 {
					mapping.BolaCols[num-1] = i
				}
			}
		}
	}

	// If no specific ball columns found, try position-based detection
	if mapping.BolaCols[0] == -1 {
		// Assume next columns after contest and date are balls
		usedCols := make(map[int]bool)
		if mapping.ContestCol != -1 {
			usedCols[mapping.ContestCol] = true
		}
		if mapping.DateCol != -1 {
			usedCols[mapping.DateCol] = true
		}

		ballIndex := 0
		for i := 0; i < len(headers) && ballIndex < 5; i++ {
			if !usedCols[i] {
				mapping.BolaCols[ballIndex] = i
				ballIndex++
			}
		}
	}

	// Validate mapping
	if mapping.ContestCol == -1 {
		return mapping, errors.New("could not detect contest column")
	}
	if mapping.DateCol == -1 {
		return mapping, errors.New("could not detect date column")
	}
	for i, col := range mapping.BolaCols {
		if col == -1 {
			return mapping, fmt.Errorf("could not detect bola%d column", i+1)
		}
	}

	return mapping, nil
}

// extractNumber extracts a number from a string like "bola1", "bola 1", etc.
func extractNumber(s string) int {
	// Find the last digit sequence
	var numStr strings.Builder
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] >= '0' && s[i] <= '9' {
			numStr.WriteByte(s[i])
		} else if numStr.Len() > 0 {
			break
		}
	}

	if numStr.Len() == 0 {
		return 0
	}

	// Reverse the string since we built it backwards
	runes := []rune(numStr.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	num, _ := strconv.Atoi(string(runes))
	return num
}
