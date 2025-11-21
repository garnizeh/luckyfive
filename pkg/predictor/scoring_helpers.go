package predictor

import "sort"

// scoreCandidateGeneric computes a heuristic score for a candidate combination.
// It uses pairwise conditional probabilities (cond), frequency counts (freq),
// optional per-number weights (cfgWeights) and positional frequency sums (posFreqSum).
func scoreCandidateGeneric(candidate []int, cond map[int]map[int]float64, freq map[int]int, cfgWeights map[int]float64, posFreqSum map[int]float64) float64 {
	// default weights
	alpha := 1.5
	beta := 0.5
	gamma := 1.0
	clusterPenalty := 2.0
	if cfgWeights != nil {
		if v, ok := cfgWeights[1]; ok {
			alpha = v
		}
		if v, ok := cfgWeights[2]; ok {
			beta = v
		}
		if v, ok := cfgWeights[3]; ok {
			gamma = v
		}
		if v, ok := cfgWeights[4]; ok {
			clusterPenalty = v
		}
	}

	coScore := 0.0
	for i := 0; i < len(candidate); i++ {
		for j := i + 1; j < len(candidate); j++ {
			a := candidate[i]
			b := candidate[j]
			if cond[a] != nil {
				coScore += cond[a][b]
			}
			if cond[b] != nil {
				coScore += cond[b][a]
			}
		}
	}

	marg := 0.0
	totalFreq := 0.0
	for _, v := range freq {
		totalFreq += float64(v)
	}
	if totalFreq == 0 {
		totalFreq = 1.0
	}
	for _, n := range candidate {
		marg += float64(freq[n]) / totalFreq
	}

	posScore := 0.0
	if posFreqSum != nil {
		for _, n := range candidate {
			posScore += posFreqSum[n]
		}
	}

	decadeCount := make(map[int]int)
	for _, n := range candidate {
		d := n / 10
		decadeCount[d]++
	}
	cluster := 0.0
	for _, c := range decadeCount {
		if c > 1 {
			cluster += float64(c - 1)
		}
	}

	// final score
	score := alpha*coScore + beta*marg + gamma*posScore - clusterPenalty*cluster

	// normalization: prefer larger marginal sums
	// small adjustment to make score more discriminative
	sort.Ints(candidate)
	return score
}
