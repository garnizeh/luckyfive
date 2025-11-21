package predictor

import (
	"math"
	"testing"
)

func TestScoreCandidateGeneric_Basic(t *testing.T) {
	candidate := []int{1, 2, 3, 4, 5}

	cond := map[int]map[int]float64{
		1: {2: 0.1},
		2: {1: 0.2, 3: 0.05},
		3: {2: 0.05},
	}

	freq := map[int]int{
		1: 10,
		2: 20,
		3: 30,
		4: 40,
		5: 0,
	}

	cfgWeights := map[int]float64{
		1: 2.0, // alpha
		2: 1.0, // beta
		3: 0.5, // gamma
		4: 1.0, // clusterPenalty
	}

	posFreqSum := map[int]float64{
		1: 0.5,
		2: 0.4,
		3: 0.3,
		4: 0.2,
		5: 0.1,
	}

	// compute score
	// expected values computed by hand in test planning
	got := scoreCandidateGeneric(append([]int(nil), candidate...), cond, freq, cfgWeights, posFreqSum)

	want := -1.45

	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("unexpected score: got=%v want=%v", got, want)
	}
}

func TestScoreCandidateGeneric_ZeroFreq(t *testing.T) {
	candidate := []int{7, 8, 9, 10, 11}
	// empty freq map should not panic and should treat totalFreq as 1
	freq := map[int]int{}
	got := scoreCandidateGeneric(append([]int(nil), candidate...), nil, freq, nil, nil)
	if math.IsNaN(got) {
		t.Fatalf("score should not be NaN")
	}
}
