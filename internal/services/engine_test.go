package services

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/garnizeh/luckyfive/internal/store/results"
	"github.com/garnizeh/luckyfive/internal/store/results/mock"
	"github.com/garnizeh/luckyfive/pkg/predictor"
	"github.com/golang/mock/gomock"
)

func TestEngineService_RunSimulation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Mock historical data: 10 draws
	mockDraws := make([]results.Draw, 10)
	for i := 0; i < 10; i++ {
		mockDraws[i] = results.Draw{
			Contest:  int64(i + 1),
			Bola1:    int64(1 + i),
			Bola2:    int64(2 + i),
			Bola3:    int64(3 + i),
			Bola4:    int64(4 + i),
			Bola5:    int64(5 + i),
			DrawDate: "2023-01-01",
		}
	}

	// Expect the call to ListDrawsByContestRange
	mockQuerier.EXPECT().ListDrawsByContestRange(
		gomock.Any(),
		results.ListDrawsByContestRangeParams{FromContest: -4, ToContest: 5},
	).Return(mockDraws, nil)

	eng := NewEngineService(mockQuerier, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cfg := SimulationConfig{
		StartContest: 1,
		EndContest:   5,
		SimPrevMax:   5,
		SimPreds:     3,
		Seed:         99,
		Weights:      predictor.Weights{Alpha: 1.0, Beta: 1.0, Gamma: 1.0, Delta: 1.0},
	}
	res, err := eng.RunSimulation(ctx, cfg)
	if err != nil {
		t.Fatalf("RunSimulation error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected result, got nil")
	}
	if len(res.ContestResults) != 5 {
		t.Fatalf("expected 5 contest results, got %d", len(res.ContestResults))
	}
	if res.Summary.TotalContests != 5 {
		t.Fatalf("expected 5 total contests, got %d", res.Summary.TotalContests)
	}
}

func TestEngineService_RunSimulation_ErrorFetchingDraws(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Expect error when fetching draws
	mockQuerier.EXPECT().ListDrawsByContestRange(
		gomock.Any(),
		gomock.Any(),
	).Return(nil, sql.ErrNoRows)

	eng := NewEngineService(mockQuerier, logger)
	ctx := context.Background()

	cfg := SimulationConfig{
		StartContest: 1,
		EndContest:   5,
		SimPrevMax:   5,
		SimPreds:     3,
		Seed:         99,
	}
	_, err := eng.RunSimulation(ctx, cfg)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestEngineService_RunSimulation_ContextCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Mock historical data
	mockDraws := []results.Draw{
		{
			Contest:  1,
			Bola1:    1,
			Bola2:    2,
			Bola3:    3,
			Bola4:    4,
			Bola5:    5,
			DrawDate: "2023-01-01",
		},
	}

	mockQuerier.EXPECT().ListDrawsByContestRange(
		gomock.Any(),
		gomock.Any(),
	).Return(mockDraws, nil)

	eng := NewEngineService(mockQuerier, logger)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cfg := SimulationConfig{
		StartContest: 1,
		EndContest:   1,
		SimPrevMax:   5,
		SimPreds:     3,
		Seed:         99,
	}
	_, err := eng.RunSimulation(ctx, cfg)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestEngineService_RunSimulation_NoContestResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Mock empty historical data
	mockDraws := []results.Draw{}

	mockQuerier.EXPECT().ListDrawsByContestRange(
		gomock.Any(),
		gomock.Any(),
	).Return(mockDraws, nil)

	eng := NewEngineService(mockQuerier, logger)
	ctx := context.Background()

	cfg := SimulationConfig{
		StartContest: 1,
		EndContest:   5,
		SimPrevMax:   5,
		SimPreds:     3,
		Seed:         99,
	}
	res, err := eng.RunSimulation(ctx, cfg)
	if err != nil {
		t.Fatalf("RunSimulation error: %v", err)
	}
	if len(res.ContestResults) != 0 {
		t.Fatalf("expected 0 contest results, got %d", len(res.ContestResults))
	}
	if res.Summary.TotalContests != 0 {
		t.Fatalf("expected 0 total contests, got %d", res.Summary.TotalContests)
	}
}

func TestEngineService_convertDraws(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	eng := NewEngineService(mockQuerier, logger)

	input := []results.Draw{
		{
			Contest:  1,
			Bola1:    1,
			Bola2:    2,
			Bola3:    3,
			Bola4:    4,
			Bola5:    5,
			DrawDate: "2023-01-01",
		},
		{
			Contest:  2,
			Bola1:    6,
			Bola2:    7,
			Bola3:    8,
			Bola4:    9,
			Bola5:    10,
			DrawDate: "2023-01-02",
		},
	}

	result := eng.convertDraws(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 draws, got %d", len(result))
	}

	if result[0].Contest != 1 {
		t.Errorf("expected contest 1, got %d", result[0].Contest)
	}
	if len(result[0].Numbers) != 5 {
		t.Errorf("expected 5 numbers, got %d", len(result[0].Numbers))
	}
	if result[0].Numbers[0] != 1 {
		t.Errorf("expected first number 1, got %d", result[0].Numbers[0])
	}

	if result[1].Contest != 2 {
		t.Errorf("expected contest 2, got %d", result[1].Contest)
	}
	if result[1].Numbers[4] != 10 {
		t.Errorf("expected last number 10, got %d", result[1].Numbers[4])
	}
}

func TestEngineService_getHistoryUpTo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	eng := NewEngineService(mockQuerier, logger)

	draws := []predictor.Draw{
		{Contest: 1, Numbers: []int{1, 2, 3, 4, 5}},
		{Contest: 2, Numbers: []int{6, 7, 8, 9, 10}},
		{Contest: 3, Numbers: []int{11, 12, 13, 14, 15}},
		{Contest: 4, Numbers: []int{16, 17, 18, 19, 20}},
		{Contest: 5, Numbers: []int{21, 22, 23, 24, 25}},
	}

	// Test getting history up to contest 3 with max history 2
	result := eng.getHistoryUpTo(draws, 3, 2)
	if len(result) != 2 {
		t.Fatalf("expected 2 draws, got %d", len(result))
	}
	if result[0].Contest != 1 {
		t.Errorf("expected first contest 1, got %d", result[0].Contest)
	}
	if result[1].Contest != 2 {
		t.Errorf("expected second contest 2, got %d", result[1].Contest)
	}

	// Test getting history up to contest 5 with max history 10 (should get all previous)
	result = eng.getHistoryUpTo(draws, 5, 10)
	if len(result) != 4 {
		t.Fatalf("expected 4 draws, got %d", len(result))
	}
	if result[3].Contest != 4 {
		t.Errorf("expected last contest 4, got %d", result[3].Contest)
	}
}

func TestEngineService_findContestInHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mock.NewMockQuerier(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	eng := NewEngineService(mockQuerier, logger)

	draws := []predictor.Draw{
		{Contest: 1, Numbers: []int{1, 2, 3, 4, 5}},
		{Contest: 2, Numbers: []int{6, 7, 8, 9, 10}},
		{Contest: 3, Numbers: []int{11, 12, 13, 14, 15}},
	}

	// Test finding existing contest
	result := eng.findContestInHistory(draws, 2)
	if result == nil {
		t.Fatal("expected to find contest 2, got nil")
	}
	if result.Contest != 2 {
		t.Errorf("expected contest 2, got %d", result.Contest)
	}

	// Test finding non-existing contest
	result = eng.findContestInHistory(draws, 99)
	if result != nil {
		t.Errorf("expected nil for non-existing contest, got %v", result)
	}
}
