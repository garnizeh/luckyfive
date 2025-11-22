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
		params := PredictionParams{
			HistoricalDraws: []Draw{{Contest: 1, Numbers: []int{1, 2, 3, 4, 5}}},
			NumPredictions:  5,
			Weights:         Weights{Alpha: 1.0, Beta: 1.0, Gamma: 1.0, Delta: 1.0},
			Seed:            1,
		}
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
	params := PredictionParams{
		HistoricalDraws: []Draw{
			{Contest: 1, Numbers: []int{1, 2, 3, 4, 5}},
			{Contest: 2, Numbers: []int{2, 3, 4, 5, 6}},
		},
		NumPredictions: 3,
		Weights:        Weights{Alpha: 1.0, Beta: 1.0, Gamma: 1.0, Delta: 1.0},
		Seed:           42,
	}
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
	params := PredictionParams{
		HistoricalDraws: []Draw{
			{Contest: 1, Numbers: []int{1, 2, 3, 4, 5}},
			{Contest: 2, Numbers: []int{2, 3, 4, 5, 6}},
			{Contest: 3, Numbers: []int{3, 4, 5, 6, 7}},
		},
		NumPredictions: 5,
		Weights:        Weights{Alpha: 1.0, Beta: 1.0, Gamma: 1.0, Delta: 1.0},
		Seed:           seed,
	}
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

func TestAdvancedPredictor_EmptyHistory(t *testing.T) {
	p := NewAdvancedPredictor(42)
	ctx := context.Background()
	params := PredictionParams{
		HistoricalDraws: []Draw{}, // Empty history
		NumPredictions:  3,
		Weights:         Weights{Alpha: 1.0, Beta: 1.0, Gamma: 1.0, Delta: 1.0},
		Seed:            42,
	}
	res, err := p.GeneratePredictions(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error with empty history: %v", err)
	}
	if len(res) != 3 {
		t.Fatalf("expected 3 predictions, got %d", len(res))
	}
	// Should still generate valid predictions even with no history
	for _, pred := range res {
		if len(pred.Numbers) != 5 {
			t.Fatalf("expected 5 numbers per prediction, got %d", len(pred.Numbers))
		}
	}
}

func TestAdvancedPredictor_ZeroPredictions(t *testing.T) {
	p := NewAdvancedPredictor(42)
	ctx := context.Background()
	params := PredictionParams{
		HistoricalDraws: []Draw{
			{Contest: 1, Numbers: []int{1, 2, 3, 4, 5}},
		},
		NumPredictions: 0, // Zero predictions
		Weights:        Weights{Alpha: 1.0, Beta: 1.0, Gamma: 1.0, Delta: 1.0},
		Seed:           42,
	}
	res, err := p.GeneratePredictions(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("expected 0 predictions, got %d", len(res))
	}
}

func TestAdvancedPredictor_LargeHistory(t *testing.T) {
	p := NewAdvancedPredictor(42)
	ctx := context.Background()

	// Create a large history (100 draws)
	history := make([]Draw, 100)
	for i := 0; i < 100; i++ {
		history[i] = Draw{
			Contest: i + 1,
			Numbers: []int{(i % 80) + 1, ((i + 1) % 80) + 1, ((i + 2) % 80) + 1, ((i + 3) % 80) + 1, ((i + 4) % 80) + 1},
		}
	}

	params := PredictionParams{
		HistoricalDraws: history,
		NumPredictions:  5,
		Weights:         Weights{Alpha: 1.0, Beta: 1.0, Gamma: 1.0, Delta: 1.0},
		Seed:            42,
	}
	res, err := p.GeneratePredictions(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error with large history: %v", err)
	}
	if len(res) != 5 {
		t.Fatalf("expected 5 predictions, got %d", len(res))
	}
}
