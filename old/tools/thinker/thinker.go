package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
	_ "modernc.org/sqlite"
)

type SimResult struct {
	Index    int
	BestHits int
	HitsDist map[int]int
	Config   map[string]float64
	Seed     int64
	BestPred []int
}

func main() {
	dbPath := flag.String("db", "data/dados.db", "caminho para sqlite db")
	limit := flag.Int("limit", 6881, "ultimo concurso usado como histórico (valida contra limit+1)")
	sims := flag.Int("sims", 50, "quantidade de simulacoes a executar")
	preds := flag.Int("preds", 25, "quantidade de previsoes por simulacao (gen-games)")
	outDir := flag.String("out", "data/thinker", "diretorio para salvar saidas")

	// parameter ranges
	alphaMin := flag.Float64("alpha-min", 0.8, "alpha min")
	alphaMax := flag.Float64("alpha-max", 2.5, "alpha max")
	betaMin := flag.Float64("beta-min", 0.1, "beta min")
	betaMax := flag.Float64("beta-max", 1.0, "beta max")
	lambdaMin := flag.Float64("lambda-min", 0.05, "lambda min")
	lambdaMax := flag.Float64("lambda-max", 0.15, "lambda max")

	candsMult := flag.Int("cands-mult", 100, "cands mult")
	hillIter := flag.Int("hill-iter", 100, "hill iter")

	smartFilters := flag.Bool("smart-filters", true, "usar smart filters")
	seedBase := flag.Int64("seed-base", 0, "seed base (0 = time.Now)")

	flag.Parse()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "erro criar out dir: %v\n", err)
		os.Exit(1)
	}

	// open DB to fetch actual draw
	connStr := fmt.Sprintf("file:%s?cache=shared&mode=rwc", *dbPath)
	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro abrir db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	tableName := "quina"
	contestCol, numCols, err := detectSchema(db, tableName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro detect schema: %v\n", err)
		os.Exit(1)
	}

	targetContest := *limit + 1
	actual, err := loadDraw(db, tableName, contestCol, numCols, targetContest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro carregar concurso %d: %v\n", targetContest, err)
		os.Exit(1)
	}
	sort.Ints(actual)
	fmt.Printf("Validando resultados do concurso %d: %v\n", targetContest, actual)

	// seed
	var rng *rand.Rand
	if *seedBase != 0 {
		rng = rand.New(rand.NewSource(*seedBase))
	} else {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	results := make([]SimResult, 0, *sims)
	var mu sync.Mutex

	// concurrency control using errgroup with limit similar to concurrent_sweep
	defaultThreads := max(runtime.NumCPU() - 1, 1)
	// allow environment control via GOMAXPROCS or flag? use defaultThreads
	var g errgroup.Group
	g.SetLimit(defaultThreads)

	var runIdx int32

	for i := 0; i < *sims; i++ {
		g.Go(func() error {
			idx := int(atomic.AddInt32(&runIdx, 1)) // 1-based

			// stratified sampling across ranges
			t := 0.0
			if *sims > 1 {
				t = float64(i) / float64(*sims-1)
			}
			// small jitter
			jitter := func() float64 { return (rng.Float64() - 0.5) * 0.1 }

			alpha := clamp(*alphaMin+(*alphaMax-*alphaMin)*t+jitter(), *alphaMin, *alphaMax)
			beta := clamp(*betaMin+(*betaMax-*betaMin)*t+jitter(), *betaMin, *betaMax)
			lambda := clamp(*lambdaMin+(*lambdaMax-*lambdaMin)*t+jitter(), *lambdaMin, *lambdaMax)

			seed := int64(0)
			if *seedBase != 0 {
				seed = *seedBase + int64(i)
			} else {
				seed = time.Now().UnixNano() + int64(i)
			}

			cfgMap := map[string]float64{"alpha": alpha, "beta": beta, "lambda": lambda}

			simOut := filepath.Join(*outDir, fmt.Sprintf("sim_%04d_%d.txt", idx, targetContest))

			args := []string{"run", "main.go", "--db", *dbPath, "--load=false", "--gen-games", strconv.Itoa(*preds), "--limit-contest", strconv.Itoa(*limit),
				"--alpha", fmt.Sprintf("%g", alpha), "--beta", fmt.Sprintf("%g", beta), "--lambda", fmt.Sprintf("%g", lambda),
				"--cands-mult", strconv.Itoa(*candsMult), "--hill-iter", strconv.Itoa(*hillIter),
				"--smart-filters", fmt.Sprintf("%v", *smartFilters), "--seed", fmt.Sprintf("%d", seed)}

			cmd := exec.Command("go", args...)
			outF, err := os.Create(simOut)
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro criar arquivo de saida sim %d: %v\n", idx, err)
				return nil
			}
			defer outF.Close()
			cmd.Stdout = outF
			cmd.Stderr = outF
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "erro executar sim %d: %v\n", idx, err)
				return nil
			}

			// parse predictions from simOut
			cards, err := parseGeneratedCards(simOut)
			if err != nil {
				fmt.Fprintf(os.Stderr, "erro parse file %s: %v\n", simOut, err)
				return nil
			}

			// compare to actual
			best := 0
			hitsDist := map[int]int{}
			var bestPred []int
			for _, c := range cards {
				hits, matched := compareGame(actual, c)
				hitsDist[hits]++
				if hits > best {
					best = hits
					bestPred = matched
				}
			}

			mu.Lock()
			results = append(results, SimResult{Index: idx, BestHits: best, HitsDist: hitsDist, Config: cfgMap, Seed: seed, BestPred: bestPred})
			mu.Unlock()

			fmt.Printf("Sim %3d: best=%d dist=%v cfg a=%.3f b=%.3f l=%.4f seed=%d\n", idx, best, hitsDist, alpha, beta, lambda, seed)
			return nil
		})
	}

	// wait for all goroutines
	if err := g.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Um erro ocorreu durante as simulações: %v\n", err)
	}

	// sort by BestHits desc
	sort.Slice(results, func(i, j int) bool { return results[i].BestHits > results[j].BestHits })

	top := min(len(results), 10)
	fmt.Println("\nTop results:")
	for i := range top {
		r := results[i]
		fmt.Printf("#%d sim=%d best=%d seed=%d cfg=%v bestPred=%v dist=%v\n", i+1, r.Index, r.BestHits, r.Seed, r.Config, r.BestPred, r.HitsDist)
	}

	// aggregate counts for duque/terno/quadra/quina across top results
	agg := map[int]int{}
	for i := range top {
		for hits, cnt := range results[i].HitsDist {
			if hits >= 2 && hits <= 5 {
				agg[hits] += cnt
			}
		}
	}

	fmt.Println("\nAggregate hits in top", top, "simulations:")
	fmt.Printf("Duque (2 acertos): %d\n", agg[2])
	fmt.Printf("Terno (3 acertos): %d\n", agg[3])
	fmt.Printf("Quadra (4 acertos): %d\n", agg[4])
	fmt.Printf("Quina  (5 acertos): %d\n", agg[5])
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func parseGeneratedCards(path string) ([][]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	re := regexp.MustCompile(`Jogo #\\d+: \\[(.*?)\\]`)
	cards := [][]int{}
	for scanner.Scan() {
		line := scanner.Text()
		m := re.FindStringSubmatch(line)
		if len(m) == 2 {
			parts := strings.Fields(m[1])
			nums := []int{}
			for _, p := range parts {
				if p == "" {
					continue
				}
				n, err := strconv.Atoi(strings.TrimSpace(p))
				if err != nil {
					continue
				}
				nums = append(nums, n)
			}
			if len(nums) > 0 {
				cards = append(cards, nums)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return cards, nil
}

// --- small helpers copied from other tools ---
func detectSchema(db *sql.DB, tableName string) (contestCol string, numCols []string, err error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s);", tableName))
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()
	var cid int
	var name, ctype string
	var notnull, dflt_value, pk interface{}
	cols := []string{}
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			return "", nil, err
		}
		cols = append(cols, name)
	}
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

func loadDraw(db *sql.DB, tableName, contestCol string, numCols []string, contest interface{}) ([]int, error) {
	cols := strings.Join(numCols, ", ")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? LIMIT 1;", cols, tableName, contestCol)
	row := db.QueryRow(query, contest)
	vals := make([]interface{}, len(numCols))
	ptrs := make([]interface{}, len(numCols))
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
