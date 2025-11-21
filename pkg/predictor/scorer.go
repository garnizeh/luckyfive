package predictor

// Simple Scorer implementation ported from loader logic.
type ScorerImpl struct{}

func NewScorer() *ScorerImpl { return &ScorerImpl{} }

func (s *ScorerImpl) ScorePredictions(predictions []Prediction, actual []int) *ScoreResult {
	result := &ScoreResult{HitDistribution: make(map[int]int)}
	for idx, pred := range predictions {
		hits := countHits(pred.Numbers, actual)
		result.HitDistribution[hits]++
		if hits > result.BestHits {
			result.BestHits = hits
			result.BestPredictionIdx = idx
			result.BestPrediction = pred.Numbers
		}
		switch hits {
		case 5:
			result.QuinaCount++
		case 4:
			result.QuadraCount++
		case 3:
			result.TernoCount++
		}
	}
	return result
}

func countHits(prediction, actual []int) int {
	actualSet := make(map[int]bool)
	for _, a := range actual {
		actualSet[a] = true
	}
	hits := 0
	for _, p := range prediction {
		if actualSet[p] {
			hits++
		}
	}
	return hits
}
