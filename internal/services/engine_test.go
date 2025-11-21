package services

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/luckyfive/pkg/predictor"
)

func TestEngineService_RunSimulation(t *testing.T) {
	// simple historical data: 10 draws
	historical := make([][]int, 0, 10)
	for i := 0; i < 10; i++ {
		historical = append(historical, []int{1 + i, 2 + i, 3 + i, 4 + i, 5 + i})
	}

	pred := predictor.NewAdvancedPredictor(99)
	eng := NewEngineService(pred)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cfg := SimulationConfig{StartContest: 1, EndContest: 5, SimPrevMax: 5, SimPreds: 3, Seed: 99}
	res, err := eng.RunSimulation(ctx, cfg, historical)
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
