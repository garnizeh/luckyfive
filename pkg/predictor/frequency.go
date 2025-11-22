package predictor

import "math"

// ComputeFreq returns a frequency map for numbers across previous draws.
func ComputeFreq(prevDraws [][]int, maxNum int) map[int]int {
	freq := make(map[int]int)
	for _, draw := range prevDraws {
		for _, n := range draw {
			if n >= 1 && (maxNum == 0 || n <= maxNum) {
				freq[n]++
			}
		}
	}
	return freq
}

// ComputePosFreq returns counts per position (0..4) for numbers.
func ComputePosFreq(prevDraws [][]int, positions int) map[int]map[int]int {
	pos := make(map[int]map[int]int)
	for p := range positions {
		pos[p] = make(map[int]int)
	}
	for _, draw := range prevDraws {
		for p := 0; p < len(draw) && p < positions; p++ {
			n := draw[p]
			if n >= 1 {
				pos[p][n]++
			}
		}
	}
	return pos
}

// ComputePairwiseConditional computes pairwise conditional probabilities P(j|i)
// using simple frequency and co-occurrence counts with add-one smoothing.
func ComputePairwiseConditional(prevDraws [][]int, maxNum int) (map[int]map[int]float64, map[int]int) {
	if len(prevDraws) == 0 {
		return make(map[int]map[int]float64), make(map[int]int)
	}

	freq := make(map[int]int)
	coOcc := make(map[[2]int]int)
	for _, draw := range prevDraws {
		nums := make([]int, 0, len(draw))
		for _, n := range draw {
			if n >= 1 && (maxNum == 0 || n <= maxNum) {
				nums = append(nums, n)
				freq[n]++
			}
		}
		// sort not required here
		for i := range len(nums) {
			for j := i + 1; j < len(nums); j++ {
				a, b := nums[i], nums[j]
				if a < b {
					coOcc[[2]int{a, b}]++
				} else {
					coOcc[[2]int{b, a}]++
				}
			}
		}
	}

	cond := make(map[int]map[int]float64)
	smooth := 1.0
	nMax := maxNum
	if nMax == 0 {
		nMax = 80
	}
	for i := 1; i <= nMax; i++ {
		if _, ok := cond[i]; !ok {
			cond[i] = make(map[int]float64)
		}
		for j := 1; j <= nMax; j++ {
			if i == j {
				continue
			}
			var count int
			key := [2]int{i, j}
			if i < j {
				count = coOcc[key]
			} else {
				keyRev := [2]int{j, i}
				count = coOcc[keyRev]
			}
			denom := float64(freq[i]) + smooth*float64(nMax)
			if denom == 0 {
				denom = smooth * float64(nMax)
			}
			cond[i][j] = (float64(count) + smooth) / denom
		}
	}
	return cond, freq
}

// ComputeMarginalProbabilities computes marginal probabilities normalized to a total of 5.
func ComputeMarginalProbabilities(prevDraws [][]int, lambda float64, maxNum int) map[int]float64 {
	probs := make(map[int]float64)
	totalDraws := len(prevDraws)
	if totalDraws == 0 {
		return probs
	}
	for i, draw := range prevDraws {
		contestsAgo := float64(totalDraws - 1 - i)
		weight := math.Exp(-lambda * contestsAgo)
		for _, n := range draw {
			if n >= 1 && (maxNum == 0 || n <= maxNum) {
				probs[n] += weight
			}
		}
	}
	sum := 0.0
	nMax := maxNum
	if nMax == 0 {
		nMax = 80
	}
	for _, v := range probs {
		sum += v
	}
	if sum == 0 {
		sum = 1.0
	}
	for n := 1; n <= nMax; n++ {
		probs[n] = (probs[n] / sum) * 5.0
	}
	return probs
}
