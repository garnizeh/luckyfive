package predictor

import (
	"math/rand"
)

// hillClimbRefine performs a simple local search to improve a candidate based on
// conditional probabilities and other heuristics.
func hillClimbRefine(candidate []int, cond map[int]map[int]float64, freq map[int]int, cfgWeights map[int]float64, posFreqSum map[int]float64, rng *rand.Rand, iterations int) []int {
	best := make([]int, len(candidate))
	copy(best, candidate)
	bestScore := scoreCandidateGeneric(best, cond, freq, cfgWeights, posFreqSum)

	exists := func(slice []int, v int) bool {
		for _, x := range slice {
			if x == v {
				return true
			}
		}
		return false
	}

	for it := 0; it < iterations; it++ {
		pos := rng.Intn(len(best))
		// choose replacement by marginal-like weights
		candidatePool := rng.Intn(80) + 1
		if exists(best, candidatePool) {
			continue
		}
		old := best[pos]
		best[pos] = candidatePool
		// ensure sorted
		// simple insertion sort for small slice
		for i := pos; i > 0 && best[i] < best[i-1]; i-- {
			best[i], best[i-1] = best[i-1], best[i]
		}
		sc := scoreCandidateGeneric(best, cond, freq, cfgWeights, posFreqSum)
		if sc > bestScore {
			bestScore = sc
		} else {
			// revert
			for i := 0; i < len(best); i++ {
				if best[i] == candidatePool && i == pos {
					best[i] = old
				}
			}
			// re-sort
			for i := 1; i < len(best); i++ {
				for j := i; j > 0 && best[j] < best[j-1]; j-- {
					best[j], best[j-1] = best[j-1], best[j]
				}
			}
		}
	}
	return best
}

// evolvePopulation applies a simple evolutionary loop: elitism + crossover + mutation.
func evolvePopulation(pop [][]int, cond map[int]map[int]float64, freq map[int]int, cfgWeights map[int]float64, posFreqSum map[int]float64, iterations int, mutateProb float64, eliteCount int, rng *rand.Rand) [][]int {
	score := func(c []int) float64 {
		return scoreCandidateGeneric(c, cond, freq, cfgWeights, posFreqSum)
	}
	popSize := len(pop)
	if popSize == 0 {
		return pop
	}
	if eliteCount < 1 {
		eliteCount = 1
	}

	for it := 0; it < iterations; it++ {
		type ind struct {
			idx     int
			fitness float64
		}
		inds := make([]ind, 0, popSize)
		for i := 0; i < popSize; i++ {
			inds = append(inds, ind{i, score(pop[i])})
		}
		// sort by fitness desc
		for i := 0; i < len(inds)-1; i++ {
			for j := i + 1; j < len(inds); j++ {
				if inds[j].fitness > inds[i].fitness {
					inds[i], inds[j] = inds[j], inds[i]
				}
			}
		}

		newPop := make([][]int, 0, popSize)
		for i := 0; i < eliteCount && i < popSize; i++ {
			cp := make([]int, len(pop[inds[i].idx]))
			copy(cp, pop[inds[i].idx])
			newPop = append(newPop, cp)
		}

		for len(newPop) < popSize {
			p1 := pop[inds[rng.Intn(popSize)].idx]
			p2 := pop[inds[rng.Intn(popSize)].idx]

			childSet := make(map[int]bool)
			// take alternating elements from parents
			for i := 0; i < 2 && len(childSet) < 5; i++ {
				source := p1
				if i%2 == 1 {
					source = p2
				}
				for _, v := range source {
					if len(childSet) >= 5 {
						break
					}
					childSet[v] = true
				}
			}
			// fill randomly if needed
			for len(childSet) < 5 {
				candidate := rng.Intn(80) + 1
				childSet[candidate] = true
			}
			// mutation
			if rng.Float64() < mutateProb {
				// replace one element
				arr := make([]int, 0, 5)
				for k := range childSet {
					arr = append(arr, k)
				}
				idx := rng.Intn(len(arr))
				delete(childSet, arr[idx])
				childSet[rng.Intn(80)+1] = true
			}
			child := make([]int, 0, 5)
			for k := range childSet {
				child = append(child, k)
			}
			// sort
			for i := 1; i < len(child); i++ {
				for j := i; j > 0 && child[j] < child[j-1]; j-- {
					child[j], child[j-1] = child[j-1], child[j]
				}
			}
			newPop = append(newPop, child)
		}

		pop = newPop
	}

	// remove duplicates
	uniq := make([][]int, 0, len(pop))
	seen := make(map[string]bool)
	for _, p := range pop {
		k := keyFromSlice(p)
		if seen[k] {
			continue
		}
		seen[k] = true
		uniq = append(uniq, p)
	}
	return uniq
}
