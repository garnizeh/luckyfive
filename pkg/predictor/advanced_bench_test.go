package predictor

import (
	"context"
	"math/rand"
	"testing"
)

func BenchmarkGeneratePredictions(b *testing.B) {
	seed := int64(42)
	p := NewAdvancedPredictor(seed)
	// build synthetic history: 200 draws
	history := make([][]int, 0, 200)
	r := rand.New(rand.NewSource(seed))
	for range 200 {
		draw := make([]int, 5)
		for j := range 5 {
			draw[j] = r.Intn(80) + 1
		}
		history = append(history, draw)
	}

	params := PredictionParams{HistoricalDraws: history, NumPredictions: 20, Seed: seed}
	ctx := context.Background()

	for b.Loop() {
		_, err := p.GeneratePredictions(ctx, params)
		if err != nil {
			b.Fatalf("generate error: %v", err)
		}
	}
	b.ReportAllocs()
}
