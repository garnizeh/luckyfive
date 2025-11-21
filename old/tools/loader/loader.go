package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	_ "modernc.org/sqlite"
)

// --- Estruturas de Configuração Unificadas ---

type GameConfig struct {
	Lambda          float64 `json:"Lambda"`
	HotColdBoost    float64 `json:"HotColdBoost"`
	Alpha           float64 `json:"Alpha"`
	Beta            float64 `json:"Beta"`
	Gamma           float64 `json:"Gamma"`
	ClusterPenalty  float64 `json:"ClusterPenalty"`
	CandsMult       int     `json:"CandsMult"`
	HillIter        int     `json:"HillIter"`
	CooccWindow     int     `json:"CooccWindow"`
	HotWindow       int     `json:"HotWindow"`
	UseSmartFilters bool    `json:"UseSmartFilters"`
}

// --- Funções Utilitárias ---

func sanitizeName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return re.ReplaceAllString(s, "")
}

func sanitizeFilename(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, ".", "p")
	s = strings.ReplaceAll(s, ",", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "=", "_")
	return s
}

func intsToString(nums []int) string {
	out := make([]string, len(nums))
	for i, n := range nums {
		out[i] = fmt.Sprintf("%d", n)
	}
	return strings.Join(out, ",")
}

// --- Carga de Dados e Banco ---

func loadSpreadsheet(path, sheet string) ([]string, [][]string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("arquivo não encontrado: %s", path)
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	var sheetName string
	if sheet != "" {
		sheetName = sheet
	} else {
		names := f.GetSheetList()
		if len(names) == 0 {
			return nil, nil, fmt.Errorf("planilha vazia")
		}
		sheetName = names[0]
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("nenhuma linha encontrada na planilha")
	}

	head := rows[0]
	cols := make([]string, len(head))
	for i, h := range head {
		if strings.TrimSpace(h) == "" {
			cols[i] = fmt.Sprintf("col_%d", i+1)
		} else {
			cols[i] = sanitizeName(h)
			if cols[i] == "" {
				cols[i] = fmt.Sprintf("col_%d", i+1)
			}
		}
	}

	dataRows := make([][]string, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		dataRows = append(dataRows, rows[i])
	}
	return cols, dataRows, nil
}

func buildTable(db *sql.DB, tableName string, cols []string) error {
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = fmt.Sprintf("\"%s\" TEXT", c)
	}
	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(parts, ", "))
	_, err := db.Exec(createSQL)
	return err
}

func insertRows(db *sql.DB, tableName string, cols []string, rows [][]string) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", tableName, strings.Join(cols, ", "), strings.Join(placeholders, ", "))

	batchSize := 500
	inserted := 0

	for i := 0; i < len(rows); i += batchSize {
		end := min(i+batchSize, len(rows))

		tx, err := db.Begin()
		if err != nil {
			return inserted, err
		}

		stmt, err := tx.Prepare(insertSQL)
		if err != nil {
			tx.Rollback()
			return inserted, err
		}

		for _, row := range rows[i:end] {
			vals := make([]any, len(cols))
			for k := range cols {
				if k < len(row) {
					vals[k] = row[k]
				} else {
					vals[k] = nil
				}
			}
			if _, err := stmt.Exec(vals...); err != nil {
				stmt.Close()
				tx.Rollback()
				return inserted, err
			}
			inserted++
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			return inserted, err
		}
	}

	return inserted, nil
}

func detectSchema(db *sql.DB, tableName string) (contestCol string, numCols []string, err error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s);", tableName))
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()
	var cid int
	var name, ctype string
	var notnull, dflt_value, pk any
	cols := []string{}
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			return "", nil, err
		}
		cols = append(cols, name)
	}
	// find contest column
	for _, c := range cols {
		lc := strings.ToLower(c)
		if strings.Contains(lc, "conc") || strings.Contains(lc, "concurso") || strings.Contains(lc, "id") {
			contestCol = c
			break
		}
	}
	if contestCol == "" && len(cols) > 0 {
		contestCol = cols[0]
	}

	lower := map[string]string{}
	for _, c := range cols {
		lower[strings.ToLower(c)] = c
	}
	want := []string{"bola1", "bola2", "bola3", "bola4", "bola5"}
	ok := true
	for _, w := range want {
		if _, found := lower[w]; !found {
			ok = false
			break
		}
	}
	if ok {
		for _, w := range want {
			numCols = append(numCols, lower[w])
		}
		return contestCol, numCols, nil
	}

	for _, c := range cols {
		lc := strings.ToLower(c)
		if strings.Contains(lc, "bola") || strings.HasPrefix(lc, "b") || strings.Contains(lc, "num") || strings.Contains(lc, " n ") || strings.HasPrefix(lc, "n") {
			if c != contestCol {
				numCols = append(numCols, c)
			}
		}
	}
	if len(numCols) == 0 {
		for _, c := range cols {
			if c != contestCol {
				numCols = append(numCols, c)
			}
		}
	}
	return contestCol, numCols, nil
}

func loadDraw(db *sql.DB, tableName, contestCol string, numCols []string, contest any) ([]int, error) {
	cols := strings.Join(numCols, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? LIMIT 1;", cols, tableName, contestCol)
	row := db.QueryRow(query, contest)
	vals := make([]any, len(numCols))
	ptrs := make([]any, len(numCols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := row.Scan(ptrs...); err != nil {
		return nil, err
	}
	res := []int{}
	for _, v := range vals {
		if v == nil {
			continue
		}
		switch t := v.(type) {
		case int64:
			res = append(res, int(t))
		case int:
			res = append(res, t)
		case []uint8:
			s := string(t)
			var n int
			fmt.Sscanf(s, "%d", &n)
			if n > 0 {
				res = append(res, n)
			}
		case string:
			var n int
			fmt.Sscanf(t, "%d", &n)
			if n > 0 {
				res = append(res, n)
			}
		default:
			var n int
			fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &n)
			if n > 0 {
				res = append(res, n)
			}
		}
	}
	return res, nil
}

func loadPreviousDrawsBefore(db *sql.DB, tableName, contestCol string, numCols []string, target int) ([][]int, error) {
	cols := strings.Join(append([]string{contestCol}, numCols...), ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s < ? ORDER BY %s ASC;", cols, tableName, contestCol, contestCol)
	rows, err := db.Query(query, target)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := [][]int{}
	for rows.Next() {
		colsCount := 1 + len(numCols)
		vals := make([]any, colsCount)
		ptrs := make([]any, colsCount)
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		nums := []int{}
		for i := 1; i < colsCount; i++ {
			v := vals[i]
			if v == nil {
				continue
			}
			switch t := v.(type) {
			case int64:
				nums = append(nums, int(t))
			case int:
				nums = append(nums, t)
			case []uint8:
				s := string(t)
				var n int
				fmt.Sscanf(s, "%d", &n)
				if n > 0 {
					nums = append(nums, n)
				}
			case string:
				var n int
				fmt.Sscanf(t, "%d", &n)
				if n > 0 {
					nums = append(nums, n)
				}
			default:
				var n int
				fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &n)
				if n > 0 {
					nums = append(nums, n)
				}
			}
		}
		if len(nums) > 0 {
			out = append(out, nums)
		}
	}
	return out, nil
}

// --- Lógica Estatística Avançada ---

func computePairwiseConditional(prevDraws [][]int) (map[int]map[int]float64, map[int]int) {
	freq := make(map[int]int)
	coOcc := make(map[[2]int]int)
	for _, draw := range prevDraws {
		nums := make([]int, 0, len(draw))
		for _, n := range draw {
			if n >= 1 && n <= 80 {
				nums = append(nums, n)
				freq[n]++
			}
		}
		sort.Ints(nums)
		for i := 0; i < len(nums); i++ {
			for j := i + 1; j < len(nums); j++ {
				key := [2]int{nums[i], nums[j]}
				coOcc[key]++
			}
		}
	}
	cond := make(map[int]map[int]float64)
	smooth := 1.0
	for i := 1; i <= 80; i++ {
		if _, ok := cond[i]; !ok {
			cond[i] = make(map[int]float64)
		}
		for j := 1; j <= 80; j++ {
			if i == j {
				continue
			}
			key := [2]int{i, j}
			keyRev := [2]int{j, i}
			count := 0
			if k := key; k[0] < k[1] {
				count = coOcc[k]
			} else {
				count = coOcc[keyRev]
			}
			cond[i][j] = (float64(count) + smooth) / (float64(freq[i]) + smooth*80.0)
		}
	}
	return cond, freq
}

func computeFreq(prevDraws [][]int) map[int]int {
	freq := make(map[int]int)
	for _, draw := range prevDraws {
		for _, n := range draw {
			if n >= 1 && n <= 80 {
				freq[n]++
			}
		}
	}
	return freq
}

func computeMarginalProbabilities(prevDraws [][]int, lambda float64) map[int]float64 {
	probs := make(map[int]float64)
	totalDraws := len(prevDraws)
	if totalDraws == 0 {
		return probs
	}
	for i, draw := range prevDraws {
		contestsAgo := float64(totalDraws - 1 - i)
		weight := math.Exp(-lambda * contestsAgo)
		for _, n := range draw {
			if n >= 1 && n <= 80 {
				probs[n] += weight
			}
		}
	}
	sum := 0.0
	for _, v := range probs {
		sum += v
	}
	if sum == 0 {
		sum = 1.0
	}
	for n := 1; n <= 80; n++ {
		probs[n] = (probs[n] / sum) * 5.0
	}
	return probs
}

func computePosFreq(prevDraws [][]int) map[int]map[int]int {
	pos := make(map[int]map[int]int)
	for p := 0; p < 5; p++ {
		pos[p] = make(map[int]int)
	}
	for _, draw := range prevDraws {
		for p := 0; p < len(draw) && p < 5; p++ {
			n := draw[p]
			if n >= 1 && n <= 80 {
				pos[p][n]++
			}
		}
	}
	return pos
}

func scoreCandidate(candidate []int, cond map[int]map[int]float64, freq map[int]int, cfg GameConfig, posFreqSum map[int]float64) float64 {
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
	return cfg.Alpha*coScore + cfg.Beta*marg + cfg.Gamma*posScore - cfg.ClusterPenalty*cluster
}

func hillClimbRefine(candidate []int, cond map[int]map[int]float64, freq map[int]int, cfg GameConfig, weightList []struct {
	num    int
	weight float64
}, posFreqSum map[int]float64, rng *rand.Rand) []int {
	best := make([]int, len(candidate))
	copy(best, candidate)
	bestScore := scoreCandidate(best, cond, freq, cfg, posFreqSum)

	exists := func(slice []int, v int) bool {
		for _, x := range slice {
			if x == v {
				return true
			}
		}
		return false
	}
	for it := 0; it < cfg.HillIter; it++ {
		pos := rng.Intn(len(best))
		r := rng.Float64() * func() float64 {
			s := 0.0
			for _, w := range weightList {
				s += w.weight
			}
			return s
		}()
		cum := 0.0
		var chosen int
		for _, w := range weightList {
			cum += w.weight
			if r <= cum {
				chosen = w.num
				break
			}
		}
		if chosen == 0 {
			chosen = rng.Intn(80) + 1
		}
		if exists(best, chosen) {
			continue
		}
		old := best[pos]
		best[pos] = chosen
		sort.Ints(best)
		sc := scoreCandidate(best, cond, freq, cfg, posFreqSum)
		if sc > bestScore {
			bestScore = sc
		} else {
			best[pos] = old // reverte
			sort.Ints(best)
		}
	}
	return best
}

func evolvePopulation(pop [][]int, cond map[int]map[int]float64, freq map[int]int, cfg GameConfig, weightList []struct {
	num    int
	weight float64
}, iterations int, mutateProb float64, eliteCount int, posFreqSum map[int]float64, rng *rand.Rand) [][]int {

	score := func(c []int) float64 {
		return scoreCandidate(c, cond, freq, cfg, posFreqSum)
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
		var inds []ind
		for i := 0; i < popSize; i++ {
			inds = append(inds, ind{i, score(pop[i])})
		}
		sort.Slice(inds, func(i, j int) bool { return inds[i].fitness > inds[j].fitness })

		newPop := make([][]int, 0, popSize)
		for i := 0; i < eliteCount && i < popSize; i++ {
			newPop = append(newPop, pop[inds[i].idx])
		}

		for len(newPop) < popSize {
			p1 := pop[inds[rng.Intn(popSize)].idx]
			p2 := pop[inds[rng.Intn(popSize)].idx]

			childSet := make(map[int]bool)
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
			for len(childSet) < 5 {
				r := rng.Float64() * func() float64 {
					s := 0.0
					for _, w := range weightList {
						s += w.weight
					}
					return s
				}()
				cum := 0.0
				var chosen int
				for _, w := range weightList {
					cum += w.weight
					if r <= cum {
						chosen = w.num
						break
					}
				}
				if chosen == 0 {
					chosen = rng.Intn(80) + 1
				}
				childSet[chosen] = true
			}
			if rng.Float64() < mutateProb {
				arr := make([]int, 0, 5)
				for k := range childSet {
					arr = append(arr, k)
				}
				sort.Ints(arr)
				if len(arr) > 0 {
					i := rng.Intn(len(arr))
					delete(childSet, arr[i])
					chosen := rng.Intn(80) + 1
					childSet[chosen] = true
				}
			}
			child := make([]int, 0, 5)
			for k := range childSet {
				child = append(child, k)
			}
			sort.Ints(child)
			newPop = append(newPop, child)
		}
		pop = newPop
	}
	uniq := make([][]int, 0, len(pop))
	seen := make(map[string]bool)
	for _, p := range pop {
		k := fmt.Sprintf("%v", p)
		if seen[k] {
			continue
		}
		seen[k] = true
		uniq = append(uniq, p)
	}
	return uniq
}

func passaFiltrosTopologicos(jogo []int) bool {
	if len(jogo) != 5 {
		return false
	}

	soma := 0
	impares := 0
	for _, n := range jogo {
		soma += n
		if n%2 != 0 {
			impares++
		}
	}

	if soma < 120 || soma > 280 {
		return false
	}
	if impares == 0 || impares == 5 {
		return false
	}
	seqCount := 0
	for i := 0; i < len(jogo)-1; i++ {
		if jogo[i+1] == jogo[i]+1 {
			seqCount++
		}
	}
	if seqCount > 2 {
		return false
	}
	return true
}

func generateAdvancedPredictions(prevDraws [][]int, numPredictions int, cfg GameConfig, rng *rand.Rand) [][]int {
	if len(prevDraws) == 0 {
		return [][]int{}
	}

	coOccDraws := prevDraws
	if cfg.CooccWindow > 0 && cfg.CooccWindow < len(prevDraws) {
		coOccDraws = prevDraws[len(prevDraws)-cfg.CooccWindow:]
	}
	cond, _ := computePairwiseConditional(coOccDraws)
	freq := computeFreq(prevDraws)

	posFreq := computePosFreq(prevDraws)
	posFreqSum := make(map[int]float64)
	totalPos := 0.0
	for p := 0; p < 5; p++ {
		for _, v := range posFreq[p] {
			totalPos += float64(v)
		}
	}
	if totalPos == 0 {
		totalPos = 1.0
	}
	for n := 1; n <= 80; n++ {
		sum := 0.0
		for p := 0; p < 5; p++ {
			sum += float64(posFreq[p][n])
		}
		posFreqSum[n] = sum / totalPos
	}

	totalDraws := len(prevDraws)
	recency := make(map[int]int)
	for i, draw := range prevDraws {
		for _, n := range draw {
			recency[n] = totalDraws - 1 - i
		}
	}
	appearedRecently := make(map[int]bool)
	hw := cfg.HotWindow
	if hw > totalDraws {
		hw = totalDraws
	}
	for _, draw := range prevDraws[len(prevDraws)-hw:] {
		for _, n := range draw {
			appearedRecently[n] = true
		}
	}

	type numW struct {
		num    int
		weight float64
	}
	var weightList []numW
	totalWeight := 0.0
	for n := 1; n <= 80; n++ {
		f := float64(freq[n])
		contestsAgo := float64(totalDraws)
		if r, ok := recency[n]; ok {
			contestsAgo = float64(r)
		}
		recW := math.Exp(-cfg.Lambda * contestsAgo)
		boost := 1.0
		if !appearedRecently[n] {
			boost = cfg.HotColdBoost
		}
		w := f * recW * boost
		if w <= 0 {
			w = 0.0001
		}
		weightList = append(weightList, numW{n, w})
		totalWeight += w
	}
	sort.Slice(weightList, func(i, j int) bool { return weightList[i].weight > weightList[j].weight })

	if rng == nil {
		rng = rand.New(rand.NewSource(1))
	}

	factor := 1
	if cfg.UseSmartFilters {
		factor = 5
	}
	numToGenerate := numPredictions * cfg.CandsMult * factor

	candidates := make([][]int, 0, numToGenerate)

	wListGeneric := make([]struct {
		num    int
		weight float64
	}, len(weightList))
	for i, w := range weightList {
		wListGeneric[i] = struct {
			num    int
			weight float64
		}{w.num, w.weight}
	}

	for i := 0; i < numToGenerate; i++ {
		r := rng.Float64() * totalWeight
		cum := 0.0
		var seed int
		for _, w := range weightList {
			cum += w.weight
			if r <= cum {
				seed = w.num
				break
			}
		}
		if seed == 0 {
			seed = rng.Intn(80) + 1
		}
		selected := []int{seed}

		for len(selected) < 5 {
			bestCand := 0
			bestProb := -1.0
			for n := 1; n <= 80; n++ {
				if slices.Contains(selected, n) {
					continue
				}
				prod := 1.0
				for _, s := range selected {
					prod *= cond[s][n]
				}
				prod *= (float64(freq[n]) + 1.0) / float64(len(prevDraws)*5)
				if prod > bestProb {
					bestProb = prod
					bestCand = n
				}
			}
			if bestCand == 0 || rng.Float64() < 0.2 {
				rr := rng.Float64() * totalWeight
				cum2 := 0.0
				for _, w := range weightList {
					cum2 += w.weight
					if rr <= cum2 {
						bestCand = w.num
						break
					}
				}
			}
			if bestCand == 0 {
				bestCand = rng.Intn(80) + 1
			}

			if !slices.Contains(selected, bestCand) {
				selected = append(selected, bestCand)
			}
		}
		sort.Ints(selected)

		refined := hillClimbRefine(selected, cond, freq, cfg, wListGeneric, posFreqSum, rng)
		candidates = append(candidates, refined)
	}

	type scored struct {
		pred  []int
		score float64
	}
	var scoredList []scored
	for _, c := range candidates {
		sc := scoreCandidate(c, cond, freq, cfg, posFreqSum)
		scoredList = append(scoredList, scored{c, sc})
	}
	sort.Slice(scoredList, func(i, j int) bool { return scoredList[i].score > scoredList[j].score })

	seedsCount := cfg.CandsMult
	if seedsCount > len(scoredList) {
		seedsCount = len(scoredList)
	}
	pop := make([][]int, 0, seedsCount)
	for i := 0; i < seedsCount; i++ {
		pop = append(pop, scoredList[i].pred)
	}

	elites := 10
	if elites > len(pop) {
		elites = len(pop)
	}
	evolved := evolvePopulation(pop, cond, freq, cfg, wListGeneric, 40, 0.15, elites, posFreqSum, rng)

	for _, e := range evolved {
		sc := scoreCandidate(e, cond, freq, cfg, posFreqSum)
		scoredList = append(scoredList, scored{e, sc})
	}
	sort.Slice(scoredList, func(i, j int) bool { return scoredList[i].score > scoredList[j].score })

	probs := computeMarginalProbabilities(prevDraws, cfg.Lambda)
	taken := make(map[int]bool)
	preds := make([][]int, 0, numPredictions)

	for len(preds) < numPredictions {
		bestIdx := -1
		bestGain := 0.0
		bestScore := math.Inf(-1)

		for i, s := range scoredList {
			if cfg.UseSmartFilters {
				if !passaFiltrosTopologicos(s.pred) {
					continue
				}
			}

			key := fmt.Sprintf("%v", s.pred)
			skip := false
			for _, p := range preds {
				if fmt.Sprintf("%v", p) == key {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			gain := 0.0
			for _, n := range s.pred {
				if !taken[n] {
					gain += probs[n]
				}
			}

			if gain > bestGain || (gain == bestGain && s.score > bestScore) {
				bestGain = gain
				bestIdx = i
				bestScore = s.score
			}
		}

		if bestIdx < 0 {
			break
		}

		chosen := scoredList[bestIdx].pred
		preds = append(preds, chosen)
		for _, n := range chosen {
			taken[n] = true
		}
		scoredList = append(scoredList[:bestIdx], scoredList[bestIdx+1:]...)
	}

	return preds
}

// --- Funções de Gerenciamento ---

type PredictionEntry struct {
	Index   int   `json:"index"`
	Numbers []int `json:"numbers"`
	Hits    int   `json:"hits"`
}

type ContestResult struct {
	Contest             int               `json:"contest"`
	Actual              []int             `json:"actual"`
	Predictions         []PredictionEntry `json:"predictions"`
	BestHits            int               `json:"best_hits"`
	BestPredictionIndex int               `json:"best_prediction_index"`
}

type SimulationResult struct {
	Params   GameConfig         `json:"params"`
	Start    int                `json:"start_contest"`
	End      int                `json:"end_contest"`
	Contests []ContestResult    `json:"contests"`
	Summary  map[string]float64 `json:"summary"`
}

func writeSimulation(outDir, base string, res *SimulationResult) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	jsonData, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	jsonFile := filepath.Join(outDir, base+".json")
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return err
	}

	csvFile := filepath.Join(outDir, base+"_details.csv")
	f, err := os.Create(csvFile)
	if err != nil {
		return err
	}
	defer f.Close()
	csvw := csv.NewWriter(f)
	defer csvw.Flush()
	csvw.Write([]string{"contest", "pred_index", "prediction", "hits", "actual"})
	for _, c := range res.Contests {
		actualStr := intsToString(c.Actual)
		for _, p := range c.Predictions {
			csvw.Write([]string{fmt.Sprintf("%d", c.Contest), fmt.Sprintf("%d", p.Index), intsToString(p.Numbers), fmt.Sprintf("%d", p.Hits), actualStr})
		}
	}

	csvSum := filepath.Join(outDir, base+"_summary.csv")
	f2, err := os.Create(csvSum)
	if err != nil {
		return err
	}
	defer f2.Close()
	csvw2 := csv.NewWriter(f2)
	defer csvw2.Flush()
	csvw2.Write([]string{"contest", "best_hits", "best_prediction_index", "actual"})
	for _, c := range res.Contests {
		csvw2.Write([]string{fmt.Sprintf("%d", c.Contest), fmt.Sprintf("%d", c.BestHits), fmt.Sprintf("%d", c.BestPredictionIndex), intsToString(c.Actual)})
	}
	return nil
}

func compareGame(actual, pred []int) (int, []int) {
	set := map[int]bool{}
	for _, a := range actual {
		set[a] = true
	}
	hits := []int{}
	for _, p := range pred {
		if set[p] {
			hits = append(hits, p)
		}
	}
	return len(hits), hits
}

type Jogo []int

func GerarJogosEstatisticos(db *sql.DB, tableName string, limiteConcurso int, qtdJogos int, cfg GameConfig, rng *rand.Rand) ([]Jogo, error) {
	contestCol, numCols, err := detectSchema(db, tableName)
	if err != nil {
		return nil, err
	}

	draws, err := loadPreviousDrawsBefore(db, tableName, contestCol, numCols, 999999)
	if err != nil {
		return nil, err
	}

	if limiteConcurso > 0 {
		draws, err = loadPreviousDrawsBefore(db, tableName, contestCol, numCols, limiteConcurso+1)
		if err != nil {
			return nil, err
		}
	}

	if len(draws) < 10 {
		return nil, fmt.Errorf("poucos dados para gerar previsão (min 10)")
	}

	rawPreds := generateAdvancedPredictions(draws, qtdJogos, cfg, rng)

	var jogos []Jogo
	for _, p := range rawPreds {
		jogos = append(jogos, Jogo(p))
	}

	return jogos, nil
}

func runSimulation(db *sql.DB, tableName string, start, end, prevMax int, numPredsSim int, cfg GameConfig, simSave bool, simOut string, rng *rand.Rand) error {
	contestCol, numCols, err := detectSchema(db, tableName)
	if err != nil {
		return fmt.Errorf("falha ao detectar esquema: %v", err)
	}
	fmt.Printf("Usando coluna concurso: %s, colunas de números: %v\n", contestCol, numCols)

	res := &SimulationResult{
		Params:   cfg,
		Start:    start,
		End:      end,
		Contests: []ContestResult{},
		Summary:  map[string]float64{},
	}

	for c := start; c <= end; c++ {
		fmt.Printf("\n--- Concurso %d ---\n", c)
		actual, err := loadDraw(db, tableName, contestCol, numCols, c)
		if err != nil {
			fmt.Printf("Concurso %d não encontrado: %v\n", c, err)
			continue
		}
		sort.Ints(actual)

		prev, err := loadPreviousDrawsBefore(db, tableName, contestCol, numCols, c)
		if err != nil {
			fmt.Printf("Erro ao carregar concursos anteriores para %d: %v\n", c, err)
			continue
		}
		if prevMax > 0 && len(prev) > prevMax {
			prev = prev[len(prev)-prevMax:]
		}

		preds := generateAdvancedPredictions(prev, numPredsSim, cfg, rng)

		bestHits := 0
		bestPredIdx := -1
		contestRes := ContestResult{Contest: c, Actual: actual, Predictions: []PredictionEntry{}, BestHits: 0, BestPredictionIndex: -1}

		for i, p := range preds {
			sort.Ints(p)
			hits, hitNums := compareGame(actual, p)
			if hits >= 3 { // Destaque visual para acertos relevantes
				fmt.Printf(">> Pred #%d: %v -> ACERTOU %d (%v)\n", i+1, p, hits, hitNums)
			} else if hits >= 2 {
				fmt.Printf("   Pred #%d: %v -> acertou %d\n", i+1, p, hits)
			}
			if hits > bestHits {
				bestHits = hits
				bestPredIdx = i
			}
			contestRes.Predictions = append(contestRes.Predictions, PredictionEntry{Index: i + 1, Numbers: p, Hits: hits})
		}
		contestRes.BestHits = bestHits
		if bestPredIdx >= 0 {
			contestRes.BestPredictionIndex = bestPredIdx + 1
		}
		res.Contests = append(res.Contests, contestRes)
		fmt.Printf("Melhor previsão do concurso %d: %d acertos\n", c, bestHits)
	}

	total := len(res.Contests)
	if total > 0 {
		sumBest := 0.0
		cnt2 := 0
		cnt3 := 0
		cnt4 := 0
		cnt5 := 0
		for _, c := range res.Contests {
			sumBest += float64(c.BestHits)
			if c.BestHits >= 2 {
				cnt2++
			}
			if c.BestHits >= 3 {
				cnt3++
			}
			if c.BestHits >= 4 {
				cnt4++
			}
			if c.BestHits >= 5 {
				cnt5++
			}
		}
		res.Summary = map[string]float64{
			"contests":      float64(total),
			"avg_best_hits": sumBest / float64(total),
			"cnt_2_plus":    float64(cnt2),
			"cnt_3_plus":    float64(cnt3),
			"cnt_4_plus":    float64(cnt4),
			"cnt_5_quina":   float64(cnt5),
		}
	}

	if simSave {
		now := time.Now().Format("20060102T150405")
		base := fmt.Sprintf("sim_%d-%d_alpha%.2f_beta%.2f_lambda%.3f_%s", start, end, cfg.Alpha, cfg.Beta, cfg.Lambda, now)
		base = sanitizeFilename(base)
		if err := writeSimulation(simOut, base, res); err != nil {
			fmt.Printf("Erro ao salvar simulação: %v\n", err)
		} else {
			fmt.Printf("Resultados salvos em %s\n", filepath.Join(simOut, base+".json"))
		}
	}
	return nil
}

func printPreview(db *sql.DB, tableName string, limit int) error {
	var count int
	if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", tableName)).Scan(&count); err != nil {
		return err
	}
	fmt.Printf("Tabela '%s' encontrada, total de linhas: %d\n", tableName, count)

	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d;", tableName, limit)
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(strings.Join(columns, " | "))
	fmt.Println(strings.Repeat("-", 50))

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		strValues := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				strValues[i] = "NULL"
			} else {
				switch v := val.(type) {
				case []uint8:
					strValues[i] = string(v)
				default:
					strValues[i] = fmt.Sprintf("%v", v)
				}
			}
		}
		fmt.Println(strings.Join(strValues, " | "))
	}
	fmt.Println(strings.Repeat("-", 50))
	return nil
}

func main() {
	// CLI flags
	file := flag.String("file", "data/Quina.xlsx", "Caminho para arquivo .xlsx")
	sheet := flag.String("sheet", "", "Nome da planilha (opcional)")
	dbPath := flag.String("db", "", "Caminho SQLite (opcional, vazio = in-memory)")
	load := flag.Bool("load", true, "Carregar dados da planilha para o banco")

	// Simulação
	simStart := flag.Int("sim-start", 0, "Concurso inicial simulação")
	simEnd := flag.Int("sim-end", 0, "Concurso final simulação")
	simPrevMax := flag.Int("sim-prev-max", 500, "Histórico máximo para simulação")

	// --- FLAGS RESTAURADAS ---
	simPreds := flag.Int("sim-preds", 10, "Quantidade de previsões por concurso na simulação")
	simClean := flag.Bool("sim-clean", false, "Limpar diretório de saída antes de começar")
	// -------------------------

	simSave := flag.Bool("sim-save", true, "Salvar JSON/CSV da simulação")
	simOut := flag.String("sim-out", "data/simulations", "Diretório de saída simulação")

	// Geração de Jogos (Produção)
	genGames := flag.Int("gen-games", 10, "Quantidade de jogos a gerar para apostar")

	// Quando gerando jogos, permite especificar até qual concurso usar como histórico
	limitContest := flag.Int("limit-contest", 0, "Limite de concurso a usar como último histórico (ex: 6881). 0 = usa todos os concursos disponíveis")

	// Parâmetros do Algoritmo (Config)
	lambda := flag.Float64("lambda", 0.08, "Peso da recência (decai com o tempo)")
	hotColdBoost := flag.Float64("hot-cold-boost", 1.5, "Boost para números frios")
	alpha := flag.Float64("alpha", 1.5, "Peso Co-ocorrência (Pares)")
	beta := flag.Float64("beta", 0.5, "Peso Frequência Marginal")
	gamma := flag.Float64("gamma", 1.0, "Peso Afinidade Posicional")
	clusterPenalty := flag.Float64("cluster-penalty", 2.0, "Penalidade para números na mesma dezena")
	candsMult := flag.Int("cands-mult", 100, "Multiplicador de candidatos internos")
	hillIter := flag.Int("hill-iter", 100, "Iterações de otimização local")
	cooccWindow := flag.Int("coocc-window", 100, "Janela de concursos para calcular pares")
	hotWindow := flag.Int("hot-window", 15, "Janela para definir números quentes")
	smartFilters := flag.Bool("smart-filters", true, "Ativar filtros de topologia (Soma, Par/Impar)")

	seed := flag.Int64("seed", 0, "Seed aleatória (0 = tempo atual)")
	flag.Parse()

	// Configuração Unificada
	cfg := GameConfig{
		Lambda: *lambda, HotColdBoost: *hotColdBoost, Alpha: *alpha, Beta: *beta, Gamma: *gamma,
		ClusterPenalty: *clusterPenalty, CandsMult: *candsMult, HillIter: *hillIter,
		CooccWindow: *cooccWindow, HotWindow: *hotWindow, UseSmartFilters: *smartFilters,
	}

	var connStr string
	inMemory := strings.TrimSpace(*dbPath) == ""
	if inMemory {
		connStr = "file:quinadb?mode=memory&cache=shared"
	} else {
		connStr = fmt.Sprintf("file:%s?cache=shared&mode=rwc", *dbPath)
	}

	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		log.Fatalf("Falha ao abrir sqlite: %v", err)
	}
	defer db.Close()

	var rng *rand.Rand
	if *seed != 0 {
		rng = rand.New(rand.NewSource(*seed))
		fmt.Printf("Seed fixa: %d\n", *seed)
	} else {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	tableName := "quina"
	shouldLoad := *load || inMemory

	if shouldLoad {
		cols, dataRows, err := loadSpreadsheet(*file, *sheet)
		if err != nil {
			log.Fatalf("Erro carga planilha: %v", err)
		}
		if err := buildTable(db, tableName, cols); err != nil {
			log.Fatalf("Erro criar tabela: %v", err)
		}
		inserted, err := insertRows(db, tableName, cols, dataRows)
		if err != nil {
			log.Fatalf("Erro inserir linhas: %v", err)
		}
		fmt.Printf("Dados carregados: %d linhas\n", inserted)
	}

	printPreview(db, tableName, 1)

	// 1. Modo Simulação
	if *simStart > 0 && *simEnd >= *simStart {
		fmt.Println("--- Iniciando Simulação (Backtesting) ---")

		// Lógica de Clean
		if *simClean {
			absOut, _ := filepath.Abs(*simOut)
			if strings.Contains(strings.ToLower(absOut), "simul") { // Safety check
				os.RemoveAll(absOut)
				os.MkdirAll(absOut, 0755)
				fmt.Printf("Diretório limpo: %s\n", absOut)
			}
		}

		// Passamos agora o simPreds recuperado das flags
		if err := runSimulation(db, tableName, *simStart, *simEnd, *simPrevMax, *simPreds, cfg, *simSave, *simOut, rng); err != nil {
			log.Fatalf("Erro na simulação: %v", err)
		}
	} else {
		// 2. Modo Geração de Jogos (Default)
		fmt.Println("\n--- Gerando Jogos Otimizados para Próximo Concurso ---")
		fmt.Printf("Config: Alpha=%.2f, Beta=%.2f, SmartFilters=%v\n", cfg.Alpha, cfg.Beta, cfg.UseSmartFilters)

		jogos, err := GerarJogosEstatisticos(db, tableName, *limitContest, *genGames, cfg, rng)
		if err != nil {
			log.Fatalf("Erro gerar jogos: %v", err)
		}

		fmt.Printf("\nSugestões Geradas (%d jogos):\n", len(jogos))
		fmt.Println("------------------------------------------------")
		for i, jogo := range jogos {
			fmt.Printf("Jogo #%02d: %v\n", i+1, jogo)
		}
		fmt.Println("------------------------------------------------")
	}
}
