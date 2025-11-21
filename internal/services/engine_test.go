package services

import (
	"context"
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
}
