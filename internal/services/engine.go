package services

import (
	"context"
	"fmt"
	"time"

	"github.com/garnizeh/luckyfive/pkg/predictor"
)

type EngineService struct {
	predictor predictor.Predictor
	scorer    predictor.Scorer
}

func NewEngineService(pred predictor.Predictor) *EngineService {
	return &EngineService{predictor: pred, scorer: predictor.NewScorer()}
}

type SimulationConfig struct {
	StartContest int
	EndContest   int
	SimPrevMax   int
	SimPreds     int
	Seed         int64
}

type ContestResult struct {
	Contest        int
	ActualNumbers  []int
	BestHits       int
	BestPrediction []int
}

type SimulationResult struct {
	ContestResults []ContestResult
	Summary        struct {
		TotalContests int
		QuinaHits     int
		QuadraHits    int
		TernoHits     int
		AverageHits   float64
	}
	DurationMs int64
}

// RunSimulation runs a simple simulation using the provided historical draws.
// historical should be ordered by contest ascending; the 'contest' numbers are
// inferred from index offsets.
func (s *EngineService) RunSimulation(ctx context.Context, cfg SimulationConfig, historical [][]int) (*SimulationResult, error) {
	start := time.Now()
	if cfg.EndContest < cfg.StartContest {
		return nil, fmt.Errorf("invalid contest range")
	}

	var results []ContestResult
	var summary SimulationResult

	for contest := cfg.StartContest; contest <= cfg.EndContest; contest++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// build history up to this contest (use last SimPrevMax draws)
		// assume contest indexes map directly to history indexes (1-based)
		idx := contest - 1
		if idx < 0 || idx >= len(historical) {
			continue
		}
		end := idx
		startIdx := 0
		if cfg.SimPrevMax > 0 && end-cfg.SimPrevMax+1 > 0 {
			startIdx = end - cfg.SimPrevMax + 1
		}
		historySlice := make([][]int, 0, end-startIdx+1)
		for i := startIdx; i <= end; i++ {
			historySlice = append(historySlice, historical[i])
		}

		preds, err := s.predictor.GeneratePredictions(ctx, predictor.PredictionParams{
			HistoricalDraws: historySlice,
			NumPredictions:  cfg.SimPreds,
			Seed:            cfg.Seed + int64(contest),
		})
		if err != nil {
			return nil, fmt.Errorf("generate predictions: %w", err)
		}
		actual := historical[idx]
		score := s.scorer.ScorePredictions(preds, actual)

		cr := ContestResult{Contest: contest, ActualNumbers: actual, BestHits: score.BestHits, BestPrediction: score.BestPrediction}
		results = append(results, cr)

		summary.Summary.TotalContests++
		summary.Summary.QuinaHits += score.QuinaCount
		summary.Summary.QuadraHits += score.QuadraCount
		summary.Summary.TernoHits += score.TernoCount
	}

	if summary.Summary.TotalContests > 0 {
		totalHits := summary.Summary.QuinaHits*5 + summary.Summary.QuadraHits*4 + summary.Summary.TernoHits*3
		summary.Summary.AverageHits = float64(totalHits) / float64(summary.Summary.TotalContests)
	}

	summary.ContestResults = results
	summary.DurationMs = time.Since(start).Milliseconds()
	return &summary, nil
}
