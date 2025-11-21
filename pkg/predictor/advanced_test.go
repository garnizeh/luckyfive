package predictor

import (
	"context"
	"testing"
	"time"
)

func TestAdvancedPredictor_ContextCancelled(t *testing.T) {
	for i := range 2 {
		p := NewAdvancedPredictor(int64(i))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		params := PredictionParams{HistoricalDraws: [][]int{{1, 2, 3, 4, 5}}, NumPredictions: 5, Seed: 1}
		_, err := p.GeneratePredictions(ctx, params)
		if err == nil {
			t.Fatalf("expected error due to cancelled context")
		}
	}
}

func TestAdvancedPredictor_BasicReturn(t *testing.T) {
	p := NewAdvancedPredictor(42)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	params := PredictionParams{HistoricalDraws: [][]int{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}}, NumPredictions: 3, Seed: 42}
	res, err := p.GeneratePredictions(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil slice (may be empty) but got nil")
	}
}

func TestAdvancedPredictor_Deterministic(t *testing.T) {
	seed := int64(12345)
	p1 := NewAdvancedPredictor(seed)
	p2 := NewAdvancedPredictor(seed)
	ctx := context.Background()
	params := PredictionParams{HistoricalDraws: [][]int{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}, {3, 4, 5, 6, 7}}, NumPredictions: 5, Seed: seed}
	r1, err1 := p1.GeneratePredictions(ctx, params)
	if err1 != nil {
		t.Fatalf("unexpected err1: %v", err1)
	}
	r2, err2 := p2.GeneratePredictions(ctx, params)
	if err2 != nil {
		t.Fatalf("unexpected err2: %v", err2)
	}
	if len(r1) != len(r2) {
		t.Fatalf("expected same length, got %d vs %d", len(r1), len(r2))
	}
	for i := range r1 {
		if keyFromSlice(r1[i].Numbers) != keyFromSlice(r2[i].Numbers) {
			t.Fatalf("predictions differ at index %d: %v vs %v", i, r1[i].Numbers, r2[i].Numbers)
		}
	}
}
