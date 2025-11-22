package services

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/garnizeh/luckyfive/internal/store/finances"
)

func TestFinancialService_GetBetCost(t *testing.T) {
	// Mock querier
	mockQueries := &mockFinancesQuerier{}

	// Create service
	logger := slog.New(slog.NewTextHandler(nil, nil))
	service := NewFinancialService(mockQueries, nil, logger)

	ctx := context.Background()
	date := time.Now()

	// Test successful case
	mockQueries.getActiveBetCost = func(ctx context.Context, params finances.GetActiveBetCostParams) (finances.BetCost, error) {
		return finances.BetCost{CostCents: 250}, nil
	}

	cost, err := service.GetBetCost(ctx, date, 5, "BR")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cost != 250 {
		t.Errorf("expected cost 250, got %d", cost)
	}

	// Test default case when not found
	mockQueries.getActiveBetCost = func(ctx context.Context, params finances.GetActiveBetCostParams) (finances.BetCost, error) {
		return finances.BetCost{}, sql.ErrNoRows
	}

	cost, err = service.GetBetCost(ctx, date, 5, "BR")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cost != 250 {
		t.Errorf("expected default cost 250, got %d", cost)
	}

	// Test error case - this should actually test a real database error
	mockQueries.getActiveBetCost = func(ctx context.Context, params finances.GetActiveBetCostParams) (finances.BetCost, error) {
		return finances.BetCost{}, errors.New("database connection error")
	}

	_, err = service.GetBetCost(ctx, date, 5, "BR")
	if err == nil {
		t.Error("expected error for database connection failure, got nil")
	}
}

func TestFinancialService_GetPrize(t *testing.T) {
	mockQueries := &mockFinancesQuerier{}
	logger := slog.New(slog.NewTextHandler(nil, nil))
	service := NewFinancialService(mockQueries, nil, logger)

	ctx := context.Background()

	// Test successful case
	mockQueries.getPrizeRule = func(ctx context.Context, params finances.GetPrizeRuleParams) (finances.PrizeRule, error) {
		return finances.PrizeRule{AmountCents: 50000000}, nil
	}

	prize, err := service.GetPrize(ctx, 1, "quina")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if prize != 50000000 {
		t.Errorf("expected prize 50000000, got %d", prize)
	}

	// Test default case
	mockQueries.getPrizeRule = func(ctx context.Context, params finances.GetPrizeRuleParams) (finances.PrizeRule, error) {
		return finances.PrizeRule{}, sql.ErrNoRows
	}

	prize, err = service.GetPrize(ctx, 1, "quina")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if prize != 50000000 {
		t.Errorf("expected default prize 50000000, got %d", prize)
	}
}

func TestFinancialService_GetFinancialSummary(t *testing.T) {
	mockQueries := &mockFinancesQuerier{}
	logger := slog.New(slog.NewTextHandler(nil, nil))
	service := NewFinancialService(mockQueries, nil, logger)

	ctx := context.Background()

	// Mock query
	mockQueries.getSimulationFinances = func(ctx context.Context, simulationID int64) (finances.SimulationFinance, error) {
		return finances.SimulationFinance{
			SimulationID:     1,
			TotalBets:        10,
			TotalCostCents:   2500,
			TotalPrizesCents: 50000000,
			NetProfitCents:   49750000,
			RoiPercentage:    1990000.0,
			BreakEvenContest: sql.NullInt64{Int64: 5, Valid: true},
			BestPrizeCents:   sql.NullInt64{Int64: 50000000, Valid: true},
			BestPrizeContest: sql.NullInt64{Int64: 1, Valid: true},
			QuinaWins:        sql.NullInt64{Int64: 1, Valid: true},
			QuadraWins:       sql.NullInt64{Int64: 2, Valid: true},
			TernoWins:        sql.NullInt64{Int64: 3, Valid: true},
		}, nil
	}

	summary, err := service.GetFinancialSummary(ctx, 1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if summary.SimulationID != 1 {
		t.Errorf("expected simulation ID 1, got %d", summary.SimulationID)
	}
	if summary.TotalBets != 10 {
		t.Errorf("expected 10 bets, got %d", summary.TotalBets)
	}
	if summary.NetProfitCents != 49750000 {
		t.Errorf("expected net profit 49750000, got %d", summary.NetProfitCents)
	}
	if summary.QuinaWins != 1 {
		t.Errorf("expected 1 quina win, got %d", summary.QuinaWins)
	}

	// Test error case
	mockQueries.getSimulationFinances = func(ctx context.Context, simulationID int64) (finances.SimulationFinance, error) {
		return finances.SimulationFinance{}, errors.New("database error")
	}

	_, err = service.GetFinancialSummary(ctx, 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestFinancialService_CreateBudget(t *testing.T) {
	mockQueries := &mockFinancesQuerier{}
	logger := slog.New(slog.NewTextHandler(nil, nil))
	service := NewFinancialService(mockQueries, nil, logger)

	ctx := context.Background()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0) // 1 month later

	// Mock successful creation
	mockQueries.createBudget = func(ctx context.Context, params finances.CreateBudgetParams) (finances.Budget, error) {
		return finances.Budget{
			ID:               1,
			Name:             "Test Budget",
			Description:      sql.NullString{String: "Test Description", Valid: true},
			TotalAmountCents: 100000,
			SpentCents:       sql.NullInt64{Int64: 0, Valid: true},
			StartDate:        startDate.Format("2006-01-02"),
			EndDate:          sql.NullString{String: endDate.Format("2006-01-02"), Valid: true},
		}, nil
	}

	budget, err := service.CreateBudget(ctx, "Test Budget", "Test Description", 100000, startDate, endDate)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if budget.ID != 1 {
		t.Errorf("expected budget ID 1, got %d", budget.ID)
	}
	if budget.Name != "Test Budget" {
		t.Errorf("expected name 'Test Budget', got %s", budget.Name)
	}
	if budget.TotalAmountCents != 100000 {
		t.Errorf("expected total amount 100000, got %d", budget.TotalAmountCents)
	}

	// Test error case
	mockQueries.createBudget = func(ctx context.Context, params finances.CreateBudgetParams) (finances.Budget, error) {
		return finances.Budget{}, errors.New("database error")
	}

	_, err = service.CreateBudget(ctx, "Test Budget", "Test Description", 100000, startDate, endDate)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestFinancialService_AllocateBudget(t *testing.T) {
	mockQueries := &mockFinancesQuerier{}
	logger := slog.New(slog.NewTextHandler(nil, nil))
	service := NewFinancialService(mockQueries, nil, logger)

	ctx := context.Background()
	simulationID := int64(1)

	// Mock budget check
	mockQueries.getBudget = func(ctx context.Context, id int64) (finances.Budget, error) {
		return finances.Budget{
			ID:               1,
			Name:             "Test Budget",
			TotalAmountCents: 100000,
			SpentCents:       sql.NullInt64{Int64: 30000, Valid: true},
			StartDate:        time.Now().Format("2006-01-02"),
			EndDate:          sql.NullString{String: time.Now().AddDate(0, 1, 0).Format("2006-01-02"), Valid: true},
		}, nil
	}

	// Mock allocation creation
	mockQueries.allocateBudget = func(ctx context.Context, params finances.AllocateBudgetParams) (finances.BudgetAllocation, error) {
		return finances.BudgetAllocation{
			ID:             1,
			BudgetID:       1,
			SimulationID:   sql.NullInt64{Int64: 1, Valid: true},
			AllocatedCents: 50000,
			SpentCents:     sql.NullInt64{Int64: 0, Valid: true},
		}, nil
	}

	// Mock budget update
	mockQueries.updateBudgetSpent = func(ctx context.Context, params finances.UpdateBudgetSpentParams) error {
		return nil
	}

	err := service.AllocateBudget(ctx, 1, &simulationID, nil, 50000)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test insufficient budget
	err = service.AllocateBudget(ctx, 1, &simulationID, nil, 80000)
	if err == nil {
		t.Error("expected error for insufficient budget, got nil")
	}

	// Test error case
	mockQueries.getBudget = func(ctx context.Context, id int64) (finances.Budget, error) {
		return finances.Budget{}, errors.New("database error")
	}

	err = service.AllocateBudget(ctx, 1, &simulationID, nil, 50000)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestFinancialService_CheckBudgetAvailability(t *testing.T) {
	mockQueries := &mockFinancesQuerier{}
	logger := slog.New(slog.NewTextHandler(nil, nil))
	service := NewFinancialService(mockQueries, nil, logger)

	ctx := context.Background()

	// Mock budget retrieval
	mockQueries.getBudget = func(ctx context.Context, id int64) (finances.Budget, error) {
		return finances.Budget{
			ID:               1,
			Name:             "Test Budget",
			TotalAmountCents: 100000,
			SpentCents:       sql.NullInt64{Int64: 30000, Valid: true},
		}, nil
	}

	// Test sufficient budget
	available, err := service.CheckBudgetAvailability(ctx, 1, 50000)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !available {
		t.Error("expected budget to be available")
	}

	// Test insufficient budget
	available, err = service.CheckBudgetAvailability(ctx, 1, 80000)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if available {
		t.Error("expected budget to be unavailable")
	}

	// Test error case
	mockQueries.getBudget = func(ctx context.Context, id int64) (finances.Budget, error) {
		return finances.Budget{}, errors.New("database error")
	}

	_, err = service.CheckBudgetAvailability(ctx, 1, 50000)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestFinancialService_GetLedgerEntries(t *testing.T) {
	mockQueries := &mockFinancesQuerier{}
	logger := slog.New(slog.NewTextHandler(nil, nil))
	service := NewFinancialService(mockQueries, nil, logger)

	ctx := context.Background()
	simulationID := int64(1)

	// Mock ledger entries
	mockQueries.getLedgerEntries = func(ctx context.Context, params finances.GetLedgerEntriesParams) ([]finances.LedgerEntry, error) {
		return []finances.LedgerEntry{
			{
				ID:              1,
				SimulationID:    sql.NullInt64{Int64: 1, Valid: true},
				TransactionType: "bet",
				AmountCents:     -250,
				Description:     sql.NullString{String: "Bet placed", Valid: true},
				CreatedAt:       sql.NullString{String: time.Now().Format("2006-01-02 15:04:05"), Valid: true},
			},
			{
				ID:              2,
				SimulationID:    sql.NullInt64{Int64: 1, Valid: true},
				TransactionType: "prize",
				AmountCents:     50000000,
				Description:     sql.NullString{String: "Quina prize", Valid: true},
				CreatedAt:       sql.NullString{String: time.Now().Format("2006-01-02 15:04:05"), Valid: true},
			},
		}, nil
	}

	entries, err := service.GetLedgerEntries(ctx, &simulationID, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].TransactionType != "bet" {
		t.Errorf("expected first entry type 'bet', got %s", entries[0].TransactionType)
	}
	if entries[1].AmountCents != 50000000 {
		t.Errorf("expected second entry amount 50000000, got %d", entries[1].AmountCents)
	}

	// Test error case
	mockQueries.getLedgerEntries = func(ctx context.Context, params finances.GetLedgerEntriesParams) ([]finances.LedgerEntry, error) {
		return nil, errors.New("database error")
	}

	_, err = service.GetLedgerEntries(ctx, &simulationID, 10, 0)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestFinancialService_GetLedgerBalance(t *testing.T) {
	mockQueries := &mockFinancesQuerier{}
	logger := slog.New(slog.NewTextHandler(nil, nil))
	service := NewFinancialService(mockQueries, nil, logger)

	ctx := context.Background()

	// Test successful balance retrieval
	mockQueries.getLedgerBalance = func(ctx context.Context, simulationID sql.NullInt64) (interface{}, error) {
		return int64(100000), nil
	}

	balance, err := service.GetLedgerBalance(ctx, 1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if balance != 100000 {
		t.Errorf("expected balance 100000, got %d", balance)
	}

	// Test error case
	mockQueries.getLedgerBalance = func(ctx context.Context, simulationID sql.NullInt64) (interface{}, error) {
		return nil, errors.New("database error")
	}

	_, err = service.GetLedgerBalance(ctx, 1)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Test unexpected type
	mockQueries.getLedgerBalance = func(ctx context.Context, simulationID sql.NullInt64) (interface{}, error) {
		return "not an int64", nil
	}

	_, err = service.GetLedgerBalance(ctx, 1)
	if err == nil {
		t.Error("expected error for unexpected type, got nil")
	}
}

func TestFinancialService_GetLedgerEntries_WithNilSimulationID(t *testing.T) {
	mockQueries := &mockFinancesQuerier{}
	logger := slog.New(slog.NewTextHandler(nil, nil))
	service := NewFinancialService(mockQueries, nil, logger)

	ctx := context.Background()

	// Mock ledger entries for all simulations (simulationID = nil)
	mockQueries.getLedgerEntries = func(ctx context.Context, params finances.GetLedgerEntriesParams) ([]finances.LedgerEntry, error) {
		// Verify that SimulationID is not valid (nil case)
		if params.SimulationID.Valid {
			t.Errorf("expected SimulationID to be invalid (nil), but it was valid with value %d", params.SimulationID.Int64)
		}

		return []finances.LedgerEntry{
			{
				ID:              1,
				SimulationID:    sql.NullInt64{Int64: 1, Valid: true},
				TransactionType: "bet",
				AmountCents:     -250,
				Description:     sql.NullString{String: "Bet for contest 1", Valid: true},
				CreatedAt:       sql.NullString{String: time.Now().Format("2006-01-02 15:04:05"), Valid: true},
			},
			{
				ID:              2,
				SimulationID:    sql.NullInt64{Int64: 2, Valid: true},
				TransactionType: "prize",
				AmountCents:     50000000,
				Description:     sql.NullString{String: "Quina prize", Valid: true},
				CreatedAt:       sql.NullString{String: time.Now().Format("2006-01-02 15:04:05"), Valid: true},
			},
		}, nil
	}

	entries, err := service.GetLedgerEntries(ctx, nil, 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	// Verify entries from different simulations are included
	simulationIDs := make(map[int64]bool)
	for _, entry := range entries {
		if entry.SimulationID.Valid {
			simulationIDs[entry.SimulationID.Int64] = true
		}
	}
	if len(simulationIDs) != 2 {
		t.Errorf("expected entries from 2 different simulations, got %d", len(simulationIDs))
	}

	// Test error case
	mockQueries.getLedgerEntries = func(ctx context.Context, params finances.GetLedgerEntriesParams) ([]finances.LedgerEntry, error) {
		return nil, errors.New("database error")
	}

	_, err = service.GetLedgerEntries(ctx, nil, 10, 0)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

type mockFinancesQuerier struct {
	getActiveBetCost         func(ctx context.Context, params finances.GetActiveBetCostParams) (finances.BetCost, error)
	getPrizeRule             func(ctx context.Context, params finances.GetPrizeRuleParams) (finances.PrizeRule, error)
	createContestBet         func(ctx context.Context, params finances.CreateContestBetParams) (finances.ContestBet, error)
	createLedgerEntry        func(ctx context.Context, params finances.CreateLedgerEntryParams) (finances.LedgerEntry, error)
	createSimulationFinances func(ctx context.Context, params finances.CreateSimulationFinancesParams) (finances.SimulationFinance, error)
	getSimulationFinances    func(ctx context.Context, simulationID int64) (finances.SimulationFinance, error)
	createBudget             func(ctx context.Context, params finances.CreateBudgetParams) (finances.Budget, error)
	allocateBudget           func(ctx context.Context, params finances.AllocateBudgetParams) (finances.BudgetAllocation, error)
	getBudget                func(ctx context.Context, id int64) (finances.Budget, error)
	updateBudgetSpent        func(ctx context.Context, params finances.UpdateBudgetSpentParams) error
	getLedgerEntries         func(ctx context.Context, params finances.GetLedgerEntriesParams) ([]finances.LedgerEntry, error)
	getLedgerBalance         func(ctx context.Context, simulationID sql.NullInt64) (interface{}, error)
}

func (m *mockFinancesQuerier) GetActiveBetCost(ctx context.Context, params finances.GetActiveBetCostParams) (finances.BetCost, error) {
	if m.getActiveBetCost != nil {
		return m.getActiveBetCost(ctx, params)
	}
	return finances.BetCost{}, nil
}

func (m *mockFinancesQuerier) GetPrizeRule(ctx context.Context, params finances.GetPrizeRuleParams) (finances.PrizeRule, error) {
	if m.getPrizeRule != nil {
		return m.getPrizeRule(ctx, params)
	}
	return finances.PrizeRule{}, nil
}

func (m *mockFinancesQuerier) CreateContestBet(ctx context.Context, params finances.CreateContestBetParams) (finances.ContestBet, error) {
	if m.createContestBet != nil {
		return m.createContestBet(ctx, params)
	}
	return finances.ContestBet{}, nil
}

func (m *mockFinancesQuerier) CreateLedgerEntry(ctx context.Context, params finances.CreateLedgerEntryParams) (finances.LedgerEntry, error) {
	if m.createLedgerEntry != nil {
		return m.createLedgerEntry(ctx, params)
	}
	return finances.LedgerEntry{}, nil
}

func (m *mockFinancesQuerier) CreateSimulationFinances(ctx context.Context, params finances.CreateSimulationFinancesParams) (finances.SimulationFinance, error) {
	if m.createSimulationFinances != nil {
		return m.createSimulationFinances(ctx, params)
	}
	return finances.SimulationFinance{}, nil
}

func (m *mockFinancesQuerier) GetSimulationFinances(ctx context.Context, simulationID int64) (finances.SimulationFinance, error) {
	if m.getSimulationFinances != nil {
		return m.getSimulationFinances(ctx, simulationID)
	}
	return finances.SimulationFinance{}, nil
}

// Stub implementations for other methods (not used in tests)
func (m *mockFinancesQuerier) CreateBetCost(ctx context.Context, arg finances.CreateBetCostParams) (finances.BetCost, error) {
	return finances.BetCost{}, nil
}

func (m *mockFinancesQuerier) CreateBudget(ctx context.Context, params finances.CreateBudgetParams) (finances.Budget, error) {
	if m.createBudget != nil {
		return m.createBudget(ctx, params)
	}
	return finances.Budget{}, nil
}

func (m *mockFinancesQuerier) AllocateBudget(ctx context.Context, params finances.AllocateBudgetParams) (finances.BudgetAllocation, error) {
	if m.allocateBudget != nil {
		return m.allocateBudget(ctx, params)
	}
	return finances.BudgetAllocation{}, nil
}

func (m *mockFinancesQuerier) GetBudget(ctx context.Context, id int64) (finances.Budget, error) {
	if m.getBudget != nil {
		return m.getBudget(ctx, id)
	}
	return finances.Budget{}, nil
}

func (m *mockFinancesQuerier) UpdateBudgetSpent(ctx context.Context, params finances.UpdateBudgetSpentParams) error {
	if m.updateBudgetSpent != nil {
		return m.updateBudgetSpent(ctx, params)
	}
	return nil
}

func (m *mockFinancesQuerier) GetLedgerEntries(ctx context.Context, params finances.GetLedgerEntriesParams) ([]finances.LedgerEntry, error) {
	if m.getLedgerEntries != nil {
		return m.getLedgerEntries(ctx, params)
	}
	return nil, nil
}

func (m *mockFinancesQuerier) GetLedgerBalance(ctx context.Context, simulationID sql.NullInt64) (interface{}, error) {
	if m.getLedgerBalance != nil {
		return m.getLedgerBalance(ctx, simulationID)
	}
	return int64(0), nil
}
func (m *mockFinancesQuerier) GetContestBetByContest(ctx context.Context, arg finances.GetContestBetByContestParams) (finances.ContestBet, error) {
	return finances.ContestBet{}, nil
}
func (m *mockFinancesQuerier) GetContestBets(ctx context.Context, simulationID int64) ([]finances.ContestBet, error) {
	return nil, nil
}
func (m *mockFinancesQuerier) GetLedgerByDateRange(ctx context.Context, arg finances.GetLedgerByDateRangeParams) ([]finances.LedgerEntry, error) {
	return nil, nil
}
func (m *mockFinancesQuerier) GetBudgetAllocations(ctx context.Context, budgetID int64) ([]finances.BudgetAllocation, error) {
	return nil, nil
}
func (m *mockFinancesQuerier) ListPrizeRules(ctx context.Context, arg finances.ListPrizeRulesParams) ([]finances.PrizeRule, error) {
	return nil, nil
}
func (m *mockFinancesQuerier) ListTopSimulationsByROI(ctx context.Context, arg finances.ListTopSimulationsByROIParams) ([]finances.ListTopSimulationsByROIRow, error) {
	return nil, nil
}
func (m *mockFinancesQuerier) UpdateAllocationSpent(ctx context.Context, arg finances.UpdateAllocationSpentParams) error {
	return nil
}
func (m *mockFinancesQuerier) UpdateContestBetPrize(ctx context.Context, arg finances.UpdateContestBetPrizeParams) error {
	return nil
}
func (m *mockFinancesQuerier) UpdateSimulationFinances(ctx context.Context, arg finances.UpdateSimulationFinancesParams) error {
	return nil
}
func (m *mockFinancesQuerier) UpsertPrizeRule(ctx context.Context, arg finances.UpsertPrizeRuleParams) error {
	return nil
}
