package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/garnizeh/luckyfive/internal/store/results"
	"github.com/garnizeh/luckyfive/pkg/predictor"
)

type EngineServicer interface {
	RunSimulation(ctx context.Context, cfg SimulationConfig) (*SimulationResult, error)
}

type EngineService struct {
	resultsQueries results.Querier
	predictor      predictor.Predictor
	scorer         predictor.Scorer
	logger         *slog.Logger
}

func NewEngineService(
	resultsQueries results.Querier,
	logger *slog.Logger,
) *EngineService {
	return &EngineService{
		resultsQueries: resultsQueries,
		scorer:         predictor.NewScorer(),
		logger:         logger,
	}
}

type SimulationConfig struct {
	StartContest    int
	EndContest      int
	SimPrevMax      int
	SimPreds        int
	Weights         predictor.Weights
	Seed            int64
	EnableEvolution bool
	Generations     int
	MutationRate    float64
}

type SimulationResult struct {
	ContestResults []ContestResult
	Summary        Summary
	Config         SimulationConfig
	DurationMs     int64
}

type ContestResult struct {
	Contest             int
	ActualNumbers       []int
	BestHits            int
	BestPrediction      []int
	BestPredictionIndex int
	AllPredictions      []predictor.Prediction
}

type Summary struct {
	TotalContests int
	QuinaHits     int
	QuadraHits    int
	TernoHits     int
	AverageHits   float64
	HitRateQuina  float64
	HitRateQuadra float64
	HitRateTerno  float64
	TotalHits     int // Add this field to track total hits across all contests
}

func (s *EngineService) RunSimulation(
	ctx context.Context,
	cfg SimulationConfig,
) (*SimulationResult, error) {
	start := time.Now()

	// Fetch historical draws
	draws, err := s.resultsQueries.ListDrawsByContestRange(
		ctx,
		results.ListDrawsByContestRangeParams{
			FromContest: int64(cfg.StartContest - cfg.SimPrevMax),
			ToContest:   int64(cfg.EndContest),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("fetch draws: %w", err)
	}

	// Convert to predictor format
	historicalDraws := s.convertDraws(draws)

	// Initialize predictor with seed
	pred := predictor.NewAdvancedPredictor(cfg.Seed)

	// Run simulation for each contest
	var contestResults []ContestResult
	var summary Summary

	for contest := cfg.StartContest; contest <= cfg.EndContest; contest++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Get historical data up to this contest
		history := s.getHistoryUpTo(historicalDraws, contest, cfg.SimPrevMax)

		// Generate predictions
		predictions, err := pred.GeneratePredictions(ctx, predictor.PredictionParams{
			HistoricalDraws: history,
			MaxHistory:      cfg.SimPrevMax,
			NumPredictions:  cfg.SimPreds,
			Weights:         cfg.Weights,
			Seed:            cfg.Seed + int64(contest),
		})
		if err != nil {
			return nil, fmt.Errorf("generate predictions: %w", err)
		}

		// Get actual result
		actual := s.findContestInHistory(historicalDraws, contest)
		if actual == nil {
			continue
		}

		// Score predictions
		score := s.scorer.ScorePredictions(predictions, actual.Numbers)

		// Record result
		contestResults = append(contestResults, ContestResult{
			Contest:             contest,
			ActualNumbers:       actual.Numbers,
			BestHits:            score.BestHits,
			BestPrediction:      score.BestPrediction,
			BestPredictionIndex: score.BestPredictionIdx,
			AllPredictions:      predictions,
		})

		// Update summary
		summary.TotalContests++
		summary.QuinaHits += score.QuinaCount
		summary.QuadraHits += score.QuadraCount
		summary.TernoHits += score.TernoCount
		summary.TotalHits += score.BestHits
	}

	// Calculate rates
	if summary.TotalContests > 0 {
		summary.HitRateQuina = float64(summary.QuinaHits) / float64(summary.TotalContests)
		summary.HitRateQuadra = float64(summary.QuadraHits) / float64(summary.TotalContests)
		summary.HitRateTerno = float64(summary.TernoHits) / float64(summary.TotalContests)

		summary.AverageHits = float64(summary.TotalHits) / float64(summary.TotalContests)
	}

	return &SimulationResult{
		ContestResults: contestResults,
		Summary:        summary,
		Config:         cfg,
		DurationMs:     time.Since(start).Milliseconds(),
	}, nil
}

// Helper methods
func (s *EngineService) convertDraws(draws []results.Draw) []predictor.Draw {
	result := make([]predictor.Draw, len(draws))
	for i, d := range draws {
		result[i] = predictor.Draw{
			Contest: int(d.Contest),
			Numbers: []int{int(d.Bola1), int(d.Bola2), int(d.Bola3), int(d.Bola4), int(d.Bola5)},
			Date:    time.Time{}, // TODO: parse DrawDate if needed
		}
	}
	return result
}

func (s *EngineService) getHistoryUpTo(draws []predictor.Draw, upToContest, maxHistory int) []predictor.Draw {
	var result []predictor.Draw
	for _, d := range draws {
		if d.Contest < upToContest {
			result = append(result, d)
		}
	}
	if len(result) > maxHistory {
		result = result[len(result)-maxHistory:]
	}
	return result
}

func (s *EngineService) findContestInHistory(draws []predictor.Draw, contest int) *predictor.Draw {
	for _, d := range draws {
		if d.Contest == contest {
			return &d
		}
	}
	return nil
}
