package predictor

import (
	"testing"
)

func TestComputeFreqAndMarginal(t *testing.T) {
	prev := [][]int{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}, {1, 3, 5, 7, 9}}
	freq := ComputeFreq(prev, 10)
	if freq[1] != 2 {
		t.Fatalf("expected freq[1]=2 got=%d", freq[1])
	}
	if freq[2] != 2 {
		t.Fatalf("expected freq[2]=2 got=%d", freq[2])
	}

	probs := ComputeMarginalProbabilities(prev, 0.1, 10)
	// total normalized to 5.0 across 1..10
	total := 0.0
	for i := 1; i <= 10; i++ {
		total += probs[i]
	}
	if total < 4.9 || total > 5.1 {
		t.Fatalf("expected marginal total ~5.0 got=%f", total)
	}
}

func TestComputePairwiseConditional(t *testing.T) {
	prev := [][]int{{1, 2, 3}, {1, 3, 4}, {2, 3, 4}}
	cond, freq := ComputePairwiseConditional(prev, 4)
	if freq[3] != 3 {
		t.Fatalf("expected freq[3]=3 got=%d", freq[3])
	}
	// conditional probabilities exist and between 0 and 1
	if cond[1][2] <= 0 || cond[1][2] >= 1.0 {
		t.Fatalf("unexpected cond[1][2]=%f", cond[1][2])
	}
}

func TestComputePosFreq(t *testing.T) {
	prev := [][]int{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}, {1, 3, 5, 7, 9}}
	pos := ComputePosFreq(prev, 5)
	if pos[0][1] != 2 {
		t.Fatalf("expected position 0 number 1 count = 2 got=%d", pos[0][1])
	}
	if pos[0][2] != 1 {
		t.Fatalf("expected position 0 number 2 count = 1 got=%d", pos[0][2])
	}
	// check position 2 (third number) counts
	if pos[2][3] != 1 || pos[2][4] != 1 || pos[2][5] != 1 {
		t.Fatalf("unexpected counts at position 2: %#v", pos[2])
	}
}
