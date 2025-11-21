package predictor

import "testing"

func TestScorerBasic(t *testing.T) {
	preds := []Prediction{
		{Numbers: []int{1, 2, 3, 4, 5}},
		{Numbers: []int{6, 7, 8, 9, 10}},
		{Numbers: []int{1, 3, 5, 7, 9}},
	}
	actual := []int{1, 3, 5, 11, 12}
	scorer := NewScorer()
	res := scorer.ScorePredictions(preds, actual)
	if res.BestHits != 3 {
		t.Fatalf("expected BestHits=3 got=%d", res.BestHits)
	}
	if res.QuinaCount != 0 {
		t.Fatalf("expected QuinaCount=0 got=%d", res.QuinaCount)
	}
	if res.TernoCount != 2 {
		t.Fatalf("expected TernoCount=2 got=%d", res.TernoCount)
	}
}

func TestScorePredictions_MultipleQuina(t *testing.T) {
	actual := []int{1, 2, 3, 4, 5}
	preds := []Prediction{
		{Numbers: []int{1, 2, 3, 4, 5}},
		{Numbers: []int{1, 2, 3, 4, 5}},
		{Numbers: []int{6, 7, 8, 9, 10}},
	}
	res := NewScorer().ScorePredictions(preds, actual)
	if res.QuinaCount != 2 {
		t.Fatalf("expected QuinaCount=2 got=%d", res.QuinaCount)
	}
	if res.BestHits != 5 {
		t.Fatalf("expected BestHits=5 got=%d", res.BestHits)
	}
}

func TestScorePredictions_MultipleQuadra(t *testing.T) {
	actual := []int{1, 2, 3, 4, 5}
	preds := []Prediction{
		{Numbers: []int{1, 2, 3, 4, 6}},
		{Numbers: []int{1, 2, 3, 4, 7}},
		{Numbers: []int{5, 6, 7, 8, 9}},
	}
	res := NewScorer().ScorePredictions(preds, actual)
	if res.QuadraCount != 2 {
		t.Fatalf("expected QuadraCount=2 got=%d", res.QuadraCount)
	}
	if res.BestHits != 4 {
		t.Fatalf("expected BestHits=4 got=%d", res.BestHits)
	}
}
