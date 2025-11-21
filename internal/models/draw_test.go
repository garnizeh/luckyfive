package models

import (
	"strings"
	"testing"
	"time"
)

func TestDraw_Validate(t *testing.T) {
	tests := []struct {
		name    string
		draw    Draw
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid draw",
			draw: Draw{
				Contest:  1,
				DrawDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Bola1:    1,
				Bola2:    2,
				Bola3:    3,
				Bola4:    4,
				Bola5:    5,
			},
			wantErr: false,
		},
		{
			name: "valid draw with higher numbers",
			draw: Draw{
				Contest:  100,
				DrawDate: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
				Bola1:    75,
				Bola2:    76,
				Bola3:    77,
				Bola4:    78,
				Bola5:    80,
			},
			wantErr: false,
		},
		{
			name: "negative contest",
			draw: Draw{
				Contest:  -1,
				DrawDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Bola1:    1,
				Bola2:    2,
				Bola3:    3,
				Bola4:    4,
				Bola5:    5,
			},
			wantErr: true,
			errMsg:  "contest must be positive",
		},
		{
			name: "zero contest",
			draw: Draw{
				Contest:  0,
				DrawDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Bola1:    1,
				Bola2:    2,
				Bola3:    3,
				Bola4:    4,
				Bola5:    5,
			},
			wantErr: true,
			errMsg:  "contest must be positive",
		},
		{
			name: "bola1 too low",
			draw: Draw{
				Contest:  1,
				DrawDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Bola1:    0,
				Bola2:    2,
				Bola3:    3,
				Bola4:    4,
				Bola5:    5,
			},
			wantErr: true,
			errMsg:  "bola1 must be between 1 and 80, got 0",
		},
		{
			name: "bola3 too high",
			draw: Draw{
				Contest:  1,
				DrawDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Bola1:    1,
				Bola2:    2,
				Bola3:    81,
				Bola4:    4,
				Bola5:    5,
			},
			wantErr: true,
			errMsg:  "bola3 must be between 1 and 80, got 81",
		},
		{
			name: "balls not in ascending order",
			draw: Draw{
				Contest:  1,
				DrawDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Bola1:    1,
				Bola2:    3,
				Bola3:    2,
				Bola4:    4,
				Bola5:    5,
			},
			wantErr: true,
			errMsg:  "balls must be in ascending order",
		},
		{
			name: "duplicate balls",
			draw: Draw{
				Contest:  1,
				DrawDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Bola1:    1,
				Bola2:    2,
				Bola3:    3,
				Bola4:    3,
				Bola5:    5,
			},
			wantErr: true,
			errMsg:  "duplicate ball number: 3",
		},
		{
			name: "zero draw date",
			draw: Draw{
				Contest:  1,
				DrawDate: time.Time{},
				Bola1:    1,
				Bola2:    2,
				Bola3:    3,
				Bola4:    4,
				Bola5:    5,
			},
			wantErr: true,
			errMsg:  "draw date cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.draw.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestParseDrawDate(t *testing.T) {
	tests := []struct {
		name     string
		dateStr  string
		wantTime time.Time
		wantErr  bool
	}{
		{
			name:     "DD/MM/YYYY format",
			dateStr:  "13/03/1994",
			wantTime: time.Date(1994, 3, 13, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "DD-MM-YYYY format",
			dateStr:  "25-12-2023",
			wantTime: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "YYYY/MM/DD format",
			dateStr:  "2023/12/25",
			wantTime: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "YYYY-MM-DD format",
			dateStr:  "2023-12-25",
			wantTime: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "DD/MM/YY format",
			dateStr:  "25/12/23",
			wantTime: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "D/M/YYYY format",
			dateStr:  "5/3/2023",
			wantTime: time.Date(2023, 3, 5, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:    "empty string",
			dateStr: "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			dateStr: "not-a-date",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			dateStr: "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDrawDate(tt.dateStr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if !got.Equal(tt.wantTime) {
					t.Errorf("expected time %v, got %v", tt.wantTime, got)
				}
			}
		})
	}
}

func TestParseBallNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "valid number",
			input: "25",
			want:  25,
		},
		{
			name:  "number with leading zeros",
			input: "005",
			want:  5,
		},
		{
			name:  "just zero",
			input: "0",
			want:  0,
		},
		{
			name:  "multiple zeros",
			input: "000",
			want:  0,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "non-numeric",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "mixed alphanumeric",
			input:   "12abc",
			wantErr: true,
		},
		{
			name:  "number with whitespace",
			input: " 25 ",
			want:  25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBallNumber(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if got != tt.want {
					t.Errorf("expected %d, got %d", tt.want, got)
				}
			}
		})
	}
}

func TestDetectColumns(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		want    ColumnMapping
		wantErr bool
		errMsg  string
	}{
		{
			name:    "perfect headers",
			headers: []string{"Concurso", "Data Sorteio", "Bola1", "Bola2", "Bola3", "Bola4", "Bola5"},
			want: ColumnMapping{
				ContestCol: 0,
				DateCol:    1,
				BolaCols:   [5]int{2, 3, 4, 5, 6},
			},
			wantErr: false,
		},
		{
			name:    "case insensitive",
			headers: []string{"concurso", "data sorteio", "bola1", "bola2", "bola3", "bola4", "bola5"},
			want: ColumnMapping{
				ContestCol: 0,
				DateCol:    1,
				BolaCols:   [5]int{2, 3, 4, 5, 6},
			},
			wantErr: false,
		},
		{
			name:    "alternative names",
			headers: []string{"Contest", "Draw Date", "Ball1", "Ball2", "Ball3", "Ball4", "Ball5"},
			want: ColumnMapping{
				ContestCol: 0,
				DateCol:    1,
				BolaCols:   [5]int{2, 3, 4, 5, 6},
			},
			wantErr: false,
		},
		{
			name:    "missing contest column",
			headers: []string{"Data"}, // Only one column, can't have contest at position 0 and date at position 1
			wantErr: true,
			errMsg:  "could not detect bola1 column",
		},
		{
			name:    "missing date column",
			headers: []string{"Concurso"}, // Only one column, can't have date at position 1
			wantErr: true,
			errMsg:  "could not detect date column",
		},
		{
			name:    "missing ball columns",
			headers: []string{"Concurso", "Data"}, // Only 2 columns, not enough for 5 balls
			wantErr: true,
			errMsg:  "could not detect bola1 column",
		},
		{
			name:    "position-based fallback",
			headers: []string{"Concurso", "Data", "Num1", "Num2", "Num3", "Num4", "Num5"},
			want: ColumnMapping{
				ContestCol: 0,
				DateCol:    1,
				BolaCols:   [5]int{2, 3, 4, 5, 6},
			},
			wantErr: false,
		},
		{
			name:    "real Quina headers",
			headers: []string{"Concurso", "Data Sorteio", "Bola1", "Bola2", "Bola3", "Bola4", "Bola5", "Ganhadores 5 acertos"},
			want: ColumnMapping{
				ContestCol: 0,
				DateCol:    1,
				BolaCols:   [5]int{2, 3, 4, 5, 6},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectColumns(tt.headers)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if got != tt.want {
					t.Errorf("expected %+v, got %+v", tt.want, got)
				}
			}
		})
	}
}

func TestExtractNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "bola1",
			input: "bola1",
			want:  1,
		},
		{
			name:  "bola 2",
			input: "bola 2",
			want:  2,
		},
		{
			name:  "ball3",
			input: "ball3",
			want:  3,
		},
		{
			name:  "numero4",
			input: "numero4",
			want:  4,
		},
		{
			name:  "n5",
			input: "n5",
			want:  5,
		},
		{
			name:  "no number",
			input: "bola",
			want:  0,
		},
		{
			name:  "number at start",
			input: "1bola",
			want:  1,
		},
		{
			name:  "multiple numbers",
			input: "bola1 extra2",
			want:  2,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNumber(tt.input)
			if got != tt.want {
				t.Errorf("expected %d, got %d", tt.want, got)
			}
		})
	}
}
