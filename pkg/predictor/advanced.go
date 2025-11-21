package predictor

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// AdvancedPredictor implements a simplified but deterministic candidate generator
// derived from the legacy loader. It focuses on candidate generation using
// marginal probabilities and pairwise conditionals.
type AdvancedPredictor struct {
	rng *rand.Rand
}

// NewAdvancedPredictor creates a new predictor with the provided seed. Use seed=0 for time-based seed.
func NewAdvancedPredictor(seed int64) *AdvancedPredictor {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &AdvancedPredictor{rng: rand.New(rand.NewSource(seed))}
}

// GeneratePredictions produces `params.NumPredictions` predictions. The implementation
// is a simplified port: it samples a seed number from marginal probabilities and
// fills remaining numbers by maximizing pairwise conditional probabilities.
func (p *AdvancedPredictor) GeneratePredictions(ctx context.Context, params PredictionParams) ([]Prediction, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if params.NumPredictions <= 0 {
		return []Prediction{}, nil
	}

	// Compute statistics from history
	maxNum := 80
	freq := ComputeFreq(params.HistoricalDraws, maxNum)
	cond, _ := ComputePairwiseConditional(params.HistoricalDraws, maxNum)
	probs := ComputeMarginalProbabilities(params.HistoricalDraws, 0.08, maxNum)

	// positional frequency sums
	posFreq := ComputePosFreq(params.HistoricalDraws, 5)
	posFreqSum := make(map[int]float64)
	totalPos := 0.0
	for ppos := 0; ppos < 5; ppos++ {
		for _, v := range posFreq[ppos] {
			totalPos += float64(v)
		}
	}
	if totalPos == 0 {
		totalPos = 1.0
	}
	for n := 1; n <= maxNum; n++ {
		sum := 0.0
		for ppos := 0; ppos < 5; ppos++ {
			sum += float64(posFreq[ppos][n])
		}
		posFreqSum[n] = sum / totalPos
	}

	// Number of candidates to generate (heuristic)
	mult := 10
	numToGenerate := params.NumPredictions * mult
	if numToGenerate < 50 {
		numToGenerate = 50
	}

	type scored struct {
		nums  []int
		score float64
	}

	var candidates []scored
	for i := 0; i < numToGenerate; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// sample first seed by marginal probs
		first := sampleByWeight(p.rng, probs, maxNum)
		if first == 0 {
			first = p.rng.Intn(maxNum) + 1
		}
		selected := []int{first}

		// fill until 5 numbers using conditional probabilities
		for len(selected) < 5 {
			best := 0
			bestScore := -1.0
			for n := 1; n <= maxNum; n++ {
				if contains(selected, n) {
					continue
				}
				// compute combined conditional score relative to selected
				s := 0.0
				for _, a := range selected {
					s += cond[a][n]
				}
				// break ties using marginal probs
				s = s + probs[n]*0.1
				if s > bestScore {
					bestScore = s
					best = n
				}
			}
			if best == 0 {
				best = p.rng.Intn(maxNum) + 1
			}
			selected = append(selected, best)
			sort.Ints(selected)
		}

		// refine candidate via hill-climb
		refined := hillClimbRefine(selected, cond, freq, nil, posFreqSum, p.rng, 40)

		// compute a lightweight score using generic scorer
		sc := scoreCandidateGeneric(refined, cond, freq, nil, posFreqSum)
		candidates = append(candidates, scored{nums: refined, score: sc})
	}

	// evolve top seeds
	seedsCount := params.NumPredictions
	if seedsCount < 1 {
		seedsCount = 1
	}
	if seedsCount > len(candidates) {
		seedsCount = len(candidates)
	}
	pop := make([][]int, 0, seedsCount)
	for i := 0; i < seedsCount; i++ {
		pop = append(pop, candidates[i].nums)
	}
	evolved := evolvePopulation(pop, cond, freq, nil, posFreqSum, 40, 0.15, 10, p.rng)
	for _, e := range evolved {
		refined := hillClimbRefine(e, cond, freq, nil, posFreqSum, p.rng, 60)
		sc := scoreCandidateGeneric(refined, cond, freq, nil, posFreqSum)
		candidates = append(candidates, scored{nums: refined, score: sc})
	}

	// sort candidates by score desc and pick top unique
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	out := make([]Prediction, 0, params.NumPredictions)
	seen := make(map[string]bool)
	for _, c := range candidates {
		key := keyFromSlice(c.nums)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, Prediction{Numbers: c.nums, Score: c.score})
		if len(out) >= params.NumPredictions {
			break
		}
	}

	return out, nil
}

// helpers
func sampleByWeight(r *rand.Rand, weights map[int]float64, maxNum int) int {
	total := 0.0
	for i := 1; i <= maxNum; i++ {
		total += weights[i]
	}
	if total <= 0 {
		return 0
	}
	r0 := r.Float64() * total
	cum := 0.0
	for i := 1; i <= maxNum; i++ {
		cum += weights[i]
		if r0 <= cum {
			return i
		}
	}
	return 0
}

func contains(a []int, v int) bool {
	for _, x := range a {
		if x == v {
			return true
		}
	}
	return false
}

func keyFromSlice(a []int) string {
	return fmt.Sprintf("%v", a)
}
