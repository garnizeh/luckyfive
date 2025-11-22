package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/garnizeh/luckyfive/internal/store/finances"
)

type FinancialService struct {
	financesQueries finances.Querier
	financesDB      *sql.DB
	logger          *slog.Logger
}

func NewFinancialService(
	financesQueries finances.Querier,
	financesDB *sql.DB,
	logger *slog.Logger,
) *FinancialService {
	return &FinancialService{
		financesQueries: financesQueries,
		financesDB:      financesDB,
		logger:          logger,
	}
}

type FinancialSummary struct {
	SimulationID     int64
	TotalBets        int
	TotalCostCents   int64
	TotalPrizesCents int64
	NetProfitCents   int64
	ROIPercentage    float64
	BreakEvenContest int
	BestPrizeCents   int64
	BestPrizeContest int
	QuinaWins        int
	QuadraWins       int
	TernoWins        int
}

// GetBetCost returns the cost of a bet for a given date
func (s *FinancialService) GetBetCost(
	ctx context.Context,
	date time.Time,
	numbersCount int,
	region string,
) (int64, error) {
	dateStr := date.Format("2006-01-02")

	cost, err := s.financesQueries.GetActiveBetCost(ctx, finances.GetActiveBetCostParams{
		EffectiveFrom: dateStr,
		EffectiveTo:   sql.NullString{String: dateStr, Valid: true},
		Region:        sql.NullString{String: region, Valid: true},
		NumbersCount:  sql.NullInt64{Int64: int64(numbersCount), Valid: true},
	})
	if err != nil {
		// Only default for "not found" errors, propagate other database errors
		if err == sql.ErrNoRows {
			return 250, nil
		}
		return 0, fmt.Errorf("get bet cost: %w", err)
	}

	return cost.CostCents, nil
}

// GetPrize returns the prize amount for a given contest and hit type
func (s *FinancialService) GetPrize(
	ctx context.Context,
	contest int,
	prizeType string, // "quina", "quadra", "terno"
) (int64, error) {
	prize, err := s.financesQueries.GetPrizeRule(ctx, finances.GetPrizeRuleParams{
		Contest:   int64(contest),
		PrizeType: prizeType,
	})
	if err != nil {
		// Use default values if not found
		defaults := map[string]int64{
			"quina":  50000000, // 500k BRL
			"quadra": 500000,   // 5k BRL
			"terno":  10000,    // 100 BRL
		}
		return defaults[prizeType], nil
	}

	return prize.AmountCents, nil
}

// CalculateSimulationFinances calculates complete financial summary
func (s *FinancialService) CalculateSimulationFinances(
	ctx context.Context,
	simulationID int64,
	contestResults []ContestResult,
) (*FinancialSummary, error) {
	summary := &FinancialSummary{
		SimulationID: simulationID,
	}

	tx, err := s.financesDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txQueries := finances.New(tx)

	var cumulativeCost int64
	var cumulativePrizes int64

	for _, result := range contestResults {
		// Get bet cost for this contest
		// TODO: get actual contest date from results DB
		contestDate := time.Now() // Placeholder
		betCost, err := s.GetBetCost(ctx, contestDate, 5, "BR")
		if err != nil {
			return nil, fmt.Errorf("get bet cost: %w", err)
		}

		summary.TotalBets++
		summary.TotalCostCents += betCost
		cumulativeCost += betCost

		// Determine prize based on hits
		var prizeType string
		var prizeCents int64

		switch result.BestHits {
		case 5:
			prizeType = "quina"
			summary.QuinaWins++
			prizeCents, _ = s.GetPrize(ctx, result.Contest, "quina")
		case 4:
			prizeType = "quadra"
			summary.QuadraWins++
			prizeCents, _ = s.GetPrize(ctx, result.Contest, "quadra")
		case 3:
			prizeType = "terno"
			summary.TernoWins++
			prizeCents, _ = s.GetPrize(ctx, result.Contest, "terno")
		}

		summary.TotalPrizesCents += prizeCents
		cumulativePrizes += prizeCents

		// Track best prize
		if prizeCents > summary.BestPrizeCents {
			summary.BestPrizeCents = prizeCents
			summary.BestPrizeContest = result.Contest
		}

		// Track break-even
		if summary.BreakEvenContest == 0 && cumulativePrizes >= cumulativeCost {
			summary.BreakEvenContest = result.Contest
		}

		// Create contest bet record
		betNumbersJSON, _ := json.Marshal(result.BestPrediction)

		_, err = txQueries.CreateContestBet(ctx, finances.CreateContestBetParams{
			SimulationID: simulationID,
			Contest:      int64(result.Contest),
			BetNumbers:   string(betNumbersJSON),
			CostCents:    betCost,
			Hits:         sql.NullInt64{Int64: int64(result.BestHits), Valid: true},
			PrizeType:    sql.NullString{String: prizeType, Valid: prizeType != ""},
			PrizeCents:   sql.NullInt64{Int64: prizeCents, Valid: prizeCents > 0},
		})
		if err != nil {
			return nil, fmt.Errorf("create contest bet: %w", err)
		}

		// Create ledger entries
		// Cost entry
		txQueries.CreateLedgerEntry(ctx, finances.CreateLedgerEntryParams{
			TransactionDate: contestDate.Format(time.RFC3339),
			TransactionType: "bet",
			AmountCents:     -betCost,
			SimulationID:    sql.NullInt64{Int64: simulationID, Valid: true},
			Contest:         sql.NullInt64{Int64: int64(result.Contest), Valid: true},
			Description:     sql.NullString{String: fmt.Sprintf("Bet for contest %d", result.Contest), Valid: true},
		})

		// Prize entry (if any)
		if prizeCents > 0 {
			txQueries.CreateLedgerEntry(ctx, finances.CreateLedgerEntryParams{
				TransactionDate: contestDate.Format(time.RFC3339),
				TransactionType: "prize",
				AmountCents:     prizeCents,
				SimulationID:    sql.NullInt64{Int64: simulationID, Valid: true},
				Contest:         sql.NullInt64{Int64: int64(result.Contest), Valid: true},
				Description:     sql.NullString{String: fmt.Sprintf("%s win for contest %d", prizeType, result.Contest), Valid: true},
			})
		}
	}

	// Calculate net profit and ROI
	summary.NetProfitCents = summary.TotalPrizesCents - summary.TotalCostCents

	if summary.TotalCostCents > 0 {
		summary.ROIPercentage = float64(summary.NetProfitCents) / float64(summary.TotalCostCents) * 100
	}

	// Upsert simulation finances
	_, err = txQueries.CreateSimulationFinances(ctx, finances.CreateSimulationFinancesParams{
		SimulationID:     simulationID,
		TotalBets:        int64(summary.TotalBets),
		TotalCostCents:   summary.TotalCostCents,
		TotalPrizesCents: summary.TotalPrizesCents,
		NetProfitCents:   summary.NetProfitCents,
		RoiPercentage:    summary.ROIPercentage,
	})
	if err != nil {
		// Try update instead
		txQueries.UpdateSimulationFinances(ctx, finances.UpdateSimulationFinancesParams{
			SimulationID:     simulationID,
			TotalBets:        int64(summary.TotalBets),
			TotalCostCents:   summary.TotalCostCents,
			TotalPrizesCents: summary.TotalPrizesCents,
			NetProfitCents:   summary.NetProfitCents,
			RoiPercentage:    summary.ROIPercentage,
			BreakEvenContest: sql.NullInt64{Int64: int64(summary.BreakEvenContest), Valid: summary.BreakEvenContest > 0},
			BestPrizeCents:   sql.NullInt64{Int64: summary.BestPrizeCents, Valid: true},
			BestPrizeContest: sql.NullInt64{Int64: int64(summary.BestPrizeContest), Valid: summary.BestPrizeContest > 0},
			QuinaWins:        sql.NullInt64{Int64: int64(summary.QuinaWins), Valid: true},
			QuadraWins:       sql.NullInt64{Int64: int64(summary.QuadraWins), Valid: true},
			TernoWins:        sql.NullInt64{Int64: int64(summary.TernoWins), Valid: true},
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return summary, nil
}

// GetFinancialSummary retrieves existing financial summary
func (s *FinancialService) GetFinancialSummary(
	ctx context.Context,
	simulationID int64,
) (*FinancialSummary, error) {
	fin, err := s.financesQueries.GetSimulationFinances(ctx, simulationID)
	if err != nil {
		return nil, fmt.Errorf("get finances: %w", err)
	}

	return &FinancialSummary{
		SimulationID:     fin.SimulationID,
		TotalBets:        int(fin.TotalBets),
		TotalCostCents:   fin.TotalCostCents,
		TotalPrizesCents: fin.TotalPrizesCents,
		NetProfitCents:   fin.NetProfitCents,
		ROIPercentage:    fin.RoiPercentage,
		BreakEvenContest: int(fin.BreakEvenContest.Int64),
		BestPrizeCents:   fin.BestPrizeCents.Int64,
		BestPrizeContest: int(fin.BestPrizeContest.Int64),
		QuinaWins:        int(fin.QuinaWins.Int64),
		QuadraWins:       int(fin.QuadraWins.Int64),
		TernoWins:        int(fin.TernoWins.Int64),
	}, nil
}

// CreateBudget creates a new budget
func (s *FinancialService) CreateBudget(
	ctx context.Context,
	name string,
	description string,
	totalCents int64,
	startDate, endDate time.Time,
) (*finances.Budget, error) {
	budget, err := s.financesQueries.CreateBudget(ctx, finances.CreateBudgetParams{
		Name:             name,
		Description:      sql.NullString{String: description, Valid: description != ""},
		TotalAmountCents: totalCents,
		StartDate:        startDate.Format("2006-01-02"),
		EndDate:          sql.NullString{String: endDate.Format("2006-01-02"), Valid: !endDate.IsZero()},
	})
	if err != nil {
		return nil, fmt.Errorf("create budget: %w", err)
	}

	return &budget, nil
}

// AllocateBudget allocates budget to a simulation or sweep
func (s *FinancialService) AllocateBudget(
	ctx context.Context,
	budgetID int64,
	simulationID *int64,
	sweepID *int64,
	amountCents int64,
) error {
	// Check budget availability
	available, err := s.CheckBudgetAvailability(ctx, budgetID, amountCents)
	if err != nil {
		return fmt.Errorf("check budget availability: %w", err)
	}
	if !available {
		return fmt.Errorf("insufficient budget funds")
	}

	// Get current budget to calculate new spent amount
	budget, err := s.financesQueries.GetBudget(ctx, budgetID)
	if err != nil {
		return fmt.Errorf("get budget: %w", err)
	}

	currentSpent := int64(0)
	if budget.SpentCents.Valid {
		currentSpent = budget.SpentCents.Int64
	}
	newSpent := currentSpent + amountCents

	// Create allocation
	_, err = s.financesQueries.AllocateBudget(ctx, finances.AllocateBudgetParams{
		BudgetID:       budgetID,
		SimulationID:   sql.NullInt64{Int64: *simulationID, Valid: simulationID != nil},
		AllocatedCents: amountCents,
	})
	if err != nil {
		return fmt.Errorf("create allocation: %w", err)
	}

	// Update budget spent
	err = s.financesQueries.UpdateBudgetSpent(ctx, finances.UpdateBudgetSpentParams{
		SpentCents:         sql.NullInt64{Int64: newSpent, Valid: true},
		TotalAmountCents:   budget.TotalAmountCents,
		TotalAmountCents_2: budget.TotalAmountCents,
		EndDate:            budget.EndDate,
		ID:                 budgetID,
	})
	if err != nil {
		return fmt.Errorf("update budget spent: %w", err)
	}

	return nil
}

// CheckBudgetAvailability checks if budget has enough remaining funds
func (s *FinancialService) CheckBudgetAvailability(
	ctx context.Context,
	budgetID int64,
	requiredCents int64,
) (bool, error) {
	budget, err := s.financesQueries.GetBudget(ctx, budgetID)
	if err != nil {
		return false, fmt.Errorf("get budget: %w", err)
	}

	remaining := budget.TotalAmountCents
	if budget.SpentCents.Valid {
		remaining -= budget.SpentCents.Int64
	}

	return remaining >= requiredCents, nil
}

// GetLedgerEntries retrieves ledger entries with pagination
func (s *FinancialService) GetLedgerEntries(
	ctx context.Context,
	simulationID *int64,
	limit, offset int,
) ([]finances.LedgerEntry, error) {
	if simulationID != nil {
		return s.financesQueries.GetLedgerEntries(ctx, finances.GetLedgerEntriesParams{
			SimulationID: sql.NullInt64{Int64: *simulationID, Valid: true},
			Limit:        int64(limit),
			Offset:       int64(offset),
		})
	}

	// If no simulation ID, get all entries (this might need a different query)
	return s.financesQueries.GetLedgerEntries(ctx, finances.GetLedgerEntriesParams{
		SimulationID: sql.NullInt64{Valid: false},
		Limit:        int64(limit),
		Offset:       int64(offset),
	})
}

// GetLedgerBalance gets the current balance for a simulation
func (s *FinancialService) GetLedgerBalance(
	ctx context.Context,
	simulationID int64,
) (int64, error) {
	balance, err := s.financesQueries.GetLedgerBalance(ctx, sql.NullInt64{Int64: simulationID, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("get ledger balance: %w", err)
	}

	// The query returns interface{}, need to cast to int64
	if bal, ok := balance.(int64); ok {
		return bal, nil
	}
	return 0, fmt.Errorf("unexpected balance type")
}
