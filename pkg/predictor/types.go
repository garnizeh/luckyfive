package predictor

import "context"

// Prediction represents a single prediction (a set of numbers) and optional score/meta.
type Prediction struct {
	Numbers []int
	Score   float64
	Method  string
}

// PredictionParams provides inputs for GeneratePredictions.
type PredictionParams struct {
	HistoricalDraws [][]int
	MaxHistory      int   // sim_prev_max
	NumPredictions  int   // sim_preds
	Seed            int64 // deterministic seed
	// Weights and other algorithm knobs can be added here
}

// Weights contains optional algorithm weights that can be tuned or evolved.
type Weights struct {
	Alpha float64
	Beta  float64
	Gamma float64
	Delta float64
}

// ScoreResult represents aggregated scoring information for a set of predictions
type ScoreResult struct {
	BestHits          int
	BestPredictionIdx int
	BestPrediction    []int
	HitDistribution   map[int]int
	QuinaCount        int
	QuadraCount       int
	TernoCount        int
}

// Predictor is the interface used by services to generate predictions.
type Predictor interface {
	GeneratePredictions(ctx context.Context, params PredictionParams) ([]Prediction, error)
}

// Scorer is the interface used to score predictions against an actual draw.
type Scorer interface {
	ScorePredictions(predictions []Prediction, actual []int) *ScoreResult
}
