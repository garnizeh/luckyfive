package predictor

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestHillClimbRefine_NoIterationsReturnsSame(t *testing.T) {
	orig := []int{1, 2, 3, 4, 5}
	cond := map[int]map[int]float64{}
	freq := map[int]int{1: 1, 2: 1, 3: 1, 4: 1, 5: 1}
	cfg := map[int]float64{}
	pos := map[int]float64{}
	rng := rand.New(rand.NewSource(42))

	out := hillClimbRefine(append([]int(nil), orig...), cond, freq, cfg, pos, rng, 0)
	if !reflect.DeepEqual(out, orig) {
		t.Fatalf("expected same slice when iterations=0: got=%v want=%v", out, orig)
	}
}

func TestHillClimbRefine_ImprovesOrEquals(t *testing.T) {
	orig := []int{1, 2, 3, 4, 5}
	// create a cond that rewards number 80 strongly to drive improvement
	cond := map[int]map[int]float64{
		80: {1: 5.0, 2: 5.0, 3: 5.0, 4: 5.0, 5: 5.0},
	}
	freq := map[int]int{1: 10, 2: 20, 3: 30, 4: 40, 5: 0, 80: 1}
	cfg := map[int]float64{}
	pos := map[int]float64{}
	rng := rand.New(rand.NewSource(2025))

	before := scoreCandidateGeneric(append([]int(nil), orig...), cond, freq, cfg, pos)
	out := hillClimbRefine(append([]int(nil), orig...), cond, freq, cfg, pos, rng, 500)
	after := scoreCandidateGeneric(append([]int(nil), out...), cond, freq, cfg, pos)

	if after < before-1e-12 {
		t.Fatalf("refined score should be >= original: before=%v after=%v", before, after)
	}
}

func TestEvolvePopulation_ElitePreservedAndSorted(t *testing.T) {
	pop := [][]int{
		{10, 11, 12, 13, 14},
		{1, 2, 3, 4, 5},
		{20, 21, 22, 23, 24},
	}
	cond := map[int]map[int]float64{}
	freq := map[int]int{}
	cfg := map[int]float64{}
	pos := map[int]float64{}
	rng := rand.New(rand.NewSource(7))

	// compute best in initial population using copies so original pop is not mutated
	bestScore := scoreCandidateGeneric(append([]int(nil), pop[0]...), cond, freq, cfg, pos)
	for i := 1; i < len(pop); i++ {
		s := scoreCandidateGeneric(append([]int(nil), pop[i]...), cond, freq, cfg, pos)
		if s > bestScore {
			bestScore = s
		}
	}

	out := evolvePopulation(pop, cond, freq, cfg, pos, 5, 0.2, 1, rng)
	if len(out) == 0 {
		t.Fatalf("expected non-empty population")
	}

	// ensure elements are sorted and length 5
	for _, indiv := range out {
		if len(indiv) != 5 {
			t.Fatalf("expected individual length 5, got %d", len(indiv))
		}
		for i := 1; i < len(indiv); i++ {
			if indiv[i] < indiv[i-1] {
				t.Fatalf("expected sorted individual, got %v", indiv)
			}
		}
	}

	// ensure evolved population has at least one individual with fitness >= initial best (elitism)
	outBest := -1e308
	for _, p := range out {
		s := scoreCandidateGeneric(append([]int(nil), p...), cond, freq, cfg, pos)
		if s > outBest {
			outBest = s
		}
	}
	if outBest < bestScore-1e-12 {
		t.Fatalf("expected evolved best score >= initial best score: initial=%v evolved=%v (pop=%v)", bestScore, outBest, out)
	}
}

func TestEvolvePopulation_Empty(t *testing.T) {
	out := evolvePopulation([][]int{}, nil, nil, nil, nil, 3, 0.5, 1, rand.New(rand.NewSource(1)))
	if len(out) != 0 {
		t.Fatalf("expected empty output for empty input, got %v", out)
	}
}
