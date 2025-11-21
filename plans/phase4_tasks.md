# Phase 4: Financial Tracking — Detailed Tasks

**Duration:** 1 week (Week 7)  
**Estimated Effort:** 40 hours  
**Team:** 1 developer  
**Status:** Not Started

---

## Overview

Phase 4 implements comprehensive financial tracking to validate simulation experiments with real monetary data. This phase adds bet tracking, ledger management, cost analysis, and ROI calculations to determine which strategies are actually profitable.

**Success Criteria:**
- ✅ Can track bets and costs for each simulation
- ✅ Ledger maintains accurate financial history
- ✅ ROI calculated correctly based on prize rules
- ✅ Budget constraints enforced
- ✅ Financial reports generated
- ✅ Test coverage > 80%

---

## Dependencies

**From Previous Phases:**
- Simulation results available
- Contest results imported
- Comparison engine working

**External:**
- Quina prize table and rules
- Bet cost structure (varies by region)

---

## Task Breakdown

### Sprint 4.1: Financial Database & Ledger (Days 1-3)

#### Task 4.1.1: Finances Database Schema
**Effort:** 4 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Define comprehensive financial tracking schema in finances.db.

**Acceptance Criteria:**
- [ ] Ledger table created
- [ ] Bet tracking tables created
- [ ] Prize rules table created
- [ ] Budget management tables created

**Subtasks:**
1. Create `data/migrations/finances_001_initial.sql`:
   ```sql
   -- Prize table (Quina official prize structure)
   CREATE TABLE prize_rules (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       contest INTEGER NOT NULL,
       prize_type TEXT NOT NULL,  -- 'quina', 'quadra', 'terno'
       amount_cents INTEGER NOT NULL,
       winners INTEGER,
       total_collected_cents INTEGER,
       notes TEXT,
       UNIQUE(contest, prize_type)
   );
   
   -- Bet costs (per region/year)
   CREATE TABLE bet_costs (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       effective_from TEXT NOT NULL,
       effective_to TEXT,
       cost_cents INTEGER NOT NULL,
       numbers_count INTEGER DEFAULT 5,
       region TEXT DEFAULT 'BR',
       notes TEXT
   );
   
   -- Ledger entries
   CREATE TABLE ledger (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       transaction_date TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
       transaction_type TEXT NOT NULL,  -- 'bet', 'prize', 'budget', 'adjustment'
       amount_cents INTEGER NOT NULL,  -- negative for costs, positive for prizes
       simulation_id INTEGER,
       contest INTEGER,
       description TEXT,
       metadata_json TEXT,
       created_at TEXT DEFAULT CURRENT_TIMESTAMP,
       FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE SET NULL
   );
   
   -- Simulation financial summary
   CREATE TABLE simulation_finances (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       simulation_id INTEGER NOT NULL UNIQUE,
       total_bets INTEGER NOT NULL DEFAULT 0,
       total_cost_cents INTEGER NOT NULL DEFAULT 0,
       total_prizes_cents INTEGER NOT NULL DEFAULT 0,
       net_profit_cents INTEGER NOT NULL DEFAULT 0,
       roi_percentage REAL NOT NULL DEFAULT 0.0,
       break_even_contest INTEGER,
       best_prize_cents INTEGER DEFAULT 0,
       best_prize_contest INTEGER,
       quina_wins INTEGER DEFAULT 0,
       quadra_wins INTEGER DEFAULT 0,
       terno_wins INTEGER DEFAULT 0,
       updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
       FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
   );
   
   -- Contest bets (what was actually bet)
   CREATE TABLE contest_bets (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       simulation_id INTEGER NOT NULL,
       contest INTEGER NOT NULL,
       bet_numbers TEXT NOT NULL,  -- JSON array
       cost_cents INTEGER NOT NULL,
       hits INTEGER,
       prize_type TEXT,  -- 'quina', 'quadra', 'terno', null
       prize_cents INTEGER DEFAULT 0,
       placed_at TEXT DEFAULT CURRENT_TIMESTAMP,
       FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE,
       UNIQUE(simulation_id, contest, bet_numbers)
   );
   
   -- Budget management
   CREATE TABLE budgets (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       name TEXT NOT NULL,
       description TEXT,
       total_amount_cents INTEGER NOT NULL,
       spent_cents INTEGER DEFAULT 0,
       remaining_cents INTEGER,
       start_date TEXT NOT NULL,
       end_date TEXT,
       status TEXT DEFAULT 'active',  -- 'active', 'exhausted', 'expired'
       created_at TEXT DEFAULT CURRENT_TIMESTAMP,
       updated_at TEXT DEFAULT CURRENT_TIMESTAMP
   );
   
   CREATE TABLE budget_allocations (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       budget_id INTEGER NOT NULL,
       simulation_id INTEGER,
       sweep_id INTEGER,
       allocated_cents INTEGER NOT NULL,
       spent_cents INTEGER DEFAULT 0,
       created_at TEXT DEFAULT CURRENT_TIMESTAMP,
       FOREIGN KEY (budget_id) REFERENCES budgets(id) ON DELETE CASCADE,
       FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
   );
   
   -- Indexes
   CREATE INDEX idx_ledger_simulation_id ON ledger(simulation_id);
   CREATE INDEX idx_ledger_transaction_date ON ledger(transaction_date);
   CREATE INDEX idx_contest_bets_simulation_id ON contest_bets(simulation_id);
   CREATE INDEX idx_contest_bets_contest ON contest_bets(contest);
   CREATE INDEX idx_prize_rules_contest ON prize_rules(contest);
   ```

2. Create migration runner (if not exists)
3. Apply migration

**Testing:**
- Verify all tables created
- Test foreign key constraints
- Test unique constraints

---

#### Task 4.1.2: Financial Queries (sqlc)
**Effort:** 5 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Define sqlc queries for financial operations.

**Acceptance Criteria:**
- [ ] Ledger CRUD queries
- [ ] Financial summary queries
- [ ] Prize rule queries
- [ ] Budget queries
- [ ] Querier interface generated

**Subtasks:**
1. Create `internal/store/queries/finances.sql`:
   ```sql
   -- Ledger operations
   -- name: CreateLedgerEntry :one
   INSERT INTO ledger (
       transaction_date, transaction_type, amount_cents,
       simulation_id, contest, description, metadata_json
   ) VALUES (?, ?, ?, ?, ?, ?, ?)
   RETURNING *;
   
   -- name: GetLedgerEntries :many
   SELECT * FROM ledger
   WHERE simulation_id = ?
   ORDER BY transaction_date DESC
   LIMIT ? OFFSET ?;
   
   -- name: GetLedgerBalance :one
   SELECT COALESCE(SUM(amount_cents), 0) as balance
   FROM ledger
   WHERE simulation_id = ?;
   
   -- name: GetLedgerByDateRange :many
   SELECT * FROM ledger
   WHERE transaction_date >= ? AND transaction_date <= ?
   ORDER BY transaction_date DESC;
   
   -- Prize rules
   -- name: UpsertPrizeRule :exec
   INSERT INTO prize_rules (
       contest, prize_type, amount_cents, winners, total_collected_cents, notes
   ) VALUES (?, ?, ?, ?, ?, ?)
   ON CONFLICT(contest, prize_type) DO UPDATE SET
       amount_cents = excluded.amount_cents,
       winners = excluded.winners,
       total_collected_cents = excluded.total_collected_cents,
       notes = excluded.notes;
   
   -- name: GetPrizeRule :one
   SELECT * FROM prize_rules
   WHERE contest = ? AND prize_type = ?
   LIMIT 1;
   
   -- name: ListPrizeRules :many
   SELECT * FROM prize_rules
   WHERE contest >= ? AND contest <= ?
   ORDER BY contest DESC, prize_type ASC;
   
   -- Bet costs
   -- name: CreateBetCost :one
   INSERT INTO bet_costs (
       effective_from, effective_to, cost_cents, numbers_count, region, notes
   ) VALUES (?, ?, ?, ?, ?, ?)
   RETURNING *;
   
   -- name: GetActiveBetCost :one
   SELECT * FROM bet_costs
   WHERE effective_from <= ? AND (effective_to IS NULL OR effective_to >= ?)
       AND region = ? AND numbers_count = ?
   ORDER BY effective_from DESC
   LIMIT 1;
   
   -- Contest bets
   -- name: CreateContestBet :one
   INSERT INTO contest_bets (
       simulation_id, contest, bet_numbers, cost_cents, hits, prize_type, prize_cents
   ) VALUES (?, ?, ?, ?, ?, ?, ?)
   RETURNING *;
   
   -- name: GetContestBets :many
   SELECT * FROM contest_bets
   WHERE simulation_id = ?
   ORDER BY contest ASC;
   
   -- name: GetContestBetByContest :one
   SELECT * FROM contest_bets
   WHERE simulation_id = ? AND contest = ?
   LIMIT 1;
   
   -- name: UpdateContestBetPrize :exec
   UPDATE contest_bets
   SET hits = ?, prize_type = ?, prize_cents = ?
   WHERE id = ?;
   
   -- Simulation finances
   -- name: CreateSimulationFinances :one
   INSERT INTO simulation_finances (
       simulation_id, total_bets, total_cost_cents,
       total_prizes_cents, net_profit_cents, roi_percentage
   ) VALUES (?, ?, ?, ?, ?, ?)
   RETURNING *;
   
   -- name: GetSimulationFinances :one
   SELECT * FROM simulation_finances
   WHERE simulation_id = ?
   LIMIT 1;
   
   -- name: UpdateSimulationFinances :exec
   UPDATE simulation_finances
   SET total_bets = ?,
       total_cost_cents = ?,
       total_prizes_cents = ?,
       net_profit_cents = ?,
       roi_percentage = ?,
       break_even_contest = ?,
       best_prize_cents = ?,
       best_prize_contest = ?,
       quina_wins = ?,
       quadra_wins = ?,
       terno_wins = ?,
       updated_at = CURRENT_TIMESTAMP
   WHERE simulation_id = ?;
   
   -- name: ListTopSimulationsByROI :many
   SELECT sf.*, s.recipe_name, s.mode
   FROM simulation_finances sf
   JOIN simulations s ON sf.simulation_id = s.id
   WHERE s.status = 'completed'
   ORDER BY sf.roi_percentage DESC
   LIMIT ? OFFSET ?;
   
   -- Budgets
   -- name: CreateBudget :one
   INSERT INTO budgets (
       name, description, total_amount_cents, start_date, end_date
   ) VALUES (?, ?, ?, ?, ?)
   RETURNING *;
   
   -- name: GetBudget :one
   SELECT * FROM budgets WHERE id = ? LIMIT 1;
   
   -- name: UpdateBudgetSpent :exec
   UPDATE budgets
   SET spent_cents = ?,
       remaining_cents = total_amount_cents - ?,
       status = CASE
           WHEN total_amount_cents - ? <= 0 THEN 'exhausted'
           WHEN ? > end_date THEN 'expired'
           ELSE 'active'
       END,
       updated_at = CURRENT_TIMESTAMP
   WHERE id = ?;
   
   -- name: AllocateBudget :one
   INSERT INTO budget_allocations (
       budget_id, simulation_id, sweep_id, allocated_cents
   ) VALUES (?, ?, ?, ?)
   RETURNING *;
   
   -- name: UpdateAllocationSpent :exec
   UPDATE budget_allocations
   SET spent_cents = ?
   WHERE id = ?;
   
   -- name: GetBudgetAllocations :many
   SELECT * FROM budget_allocations
   WHERE budget_id = ?;
   ```

2. Run `make generate` to create Querier interface

**Testing:**
- Test all queries with sample data
- Verify aggregations
- Test edge cases

---

#### Task 4.1.3: Financial Service Implementation
**Effort:** 8 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Implement FinancialService to manage all financial operations.

**Acceptance Criteria:**
- [ ] Ledger management methods
- [ ] Cost calculation methods
- [ ] Prize calculation methods
- [ ] ROI calculation methods
- [ ] Uses Querier interface (mockable)

**Subtasks:**
1. Create `internal/services/financial.go`:
   ```go
   package services
   
   import (
       "context"
       "database/sql"
       "encoding/json"
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
       SimulationID      int64
       TotalBets         int
       TotalCostCents    int64
       TotalPrizesCents  int64
       NetProfitCents    int64
       ROIPercentage     float64
       BreakEvenContest  int
       BestPrizeCents    int64
       BestPrizeContest  int
       QuinaWins         int
       QuadraWins        int
       TernoWins         int
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
           Region:        region,
           NumbersCount:  int64(numbersCount),
       })
       if err != nil {
           // Default to 2.50 BRL (250 cents) if not found
           return 250, nil
       }
       
       return cost.CostCents, nil
   }
   
   // GetPrize returns the prize amount for a given contest and hit type
   func (s *FinancialService) GetPrize(
       ctx context.Context,
       contest int,
       prizeType string,  // "quina", "quadra", "terno"
   ) (int64, error) {
       prize, err := s.financesQueries.GetPrizeRule(ctx, finances.GetPrizeRuleParams{
           Contest:   int64(contest),
           PrizeType: prizeType,
       })
       if err != nil {
           // Use default values if not found
           defaults := map[string]int64{
               "quina":  50000000,  // 500k BRL
               "quadra": 500000,    // 5k BRL
               "terno":  10000,     // 100 BRL
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
           contestDate := time.Now()  // Placeholder
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
   ```

2. Add budget management methods
3. Add ledger query methods

**Testing:**
- Mock financesQueries
- Test cost calculation
- Test prize lookup
- Test ROI calculation
- Test ledger entries

---

### Sprint 4.2: Integration with Simulations (Days 4-5)

#### Task 4.2.1: Enhance SimulationService with Finance Tracking
**Effort:** 6 hours  
**Priority:** High  
**Assignee:** Dev 1

**Description:**
Integrate financial tracking into simulation execution flow.

**Acceptance Criteria:**
- [ ] Simulations automatically calculate finances
- [ ] Financial summary saved with simulation
- [ ] Optional: skip finance tracking flag

**Subtasks:**
1. Update `SimulationService.ExecuteSimulation`:
   ```go
   func (s *SimulationService) ExecuteSimulation(ctx context.Context, simID int64) error {
       // ... existing simulation logic ...
       
       // Calculate finances
       if s.financialService != nil {
           finSummary, err := s.financialService.CalculateSimulationFinances(
               ctx,
               simID,
               result.ContestResults,
           )
           if err != nil {
               s.logger.Warn("finance calculation failed", "error", err)
               // Don't fail simulation if finance tracking fails
           } else {
               s.logger.Info("finances calculated",
                   "simulation_id", simID,
                   "roi", finSummary.ROIPercentage,
                   "net_profit_cents", finSummary.NetProfitCents,
               )
           }
       }
       
       // ... rest of completion logic ...
   }
   ```

2. Add FinancialService dependency to SimulationService constructor
3. Update all service initialization code

**Testing:**
- Test simulation with finance tracking
- Test simulation without finance service
- Verify ledger entries created

---

#### Task 4.2.2: Financial Endpoints
**Effort:** 5 hours  
**Priority:** High  
**Assignee:** Dev 1

**Description:**
Create API endpoints for financial data access.

**Acceptance Criteria:**
- [ ] GET /api/v1/simulations/:id/finances
- [ ] GET /api/v1/finances/ledger
- [ ] GET /api/v1/finances/top-roi
- [ ] POST /api/v1/finances/prize-rules (admin)
- [ ] POST /api/v1/finances/bet-costs (admin)

**Subtasks:**
1. Create `internal/handlers/finances.go`:
   ```go
   func GetSimulationFinances(finSvc *services.FinancialService) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           simID := chi.URLParam(r, "id")
           id, _ := strconv.ParseInt(simID, 10, 64)
           
           summary, err := finSvc.GetFinancialSummary(r.Context(), id)
           if err != nil {
               WriteError(w, 404, err)
               return
           }
           
           WriteJSON(w, 200, summary)
       }
   }
   
   func GetLedger(finSvc *services.FinancialService) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           // Parse query params (sim_id, date_from, date_to, limit, offset)
           // Query ledger
           // Return entries
       }
   }
   
   func GetTopROI(finSvc *services.FinancialService) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           // Query top simulations by ROI
           // Return list
       }
   }
   ```

2. Wire into router
3. Add admin authentication for write endpoints

**Testing:**
- Test all endpoints
- Test query parameters
- Test error cases

---

### Sprint 4.3: Budget Management (Days 6-7)

#### Task 4.3.1: Budget Service
**Effort:** 6 hours  
**Priority:** Medium  
**Assignee:** Dev 1

**Description:**
Implement budget creation and tracking.

**Acceptance Criteria:**
- [ ] Create budget
- [ ] Allocate budget to simulations/sweeps
- [ ] Track spending against budget
- [ ] Enforce budget limits

**Subtasks:**
1. Extend FinancialService with budget methods:
   ```go
   func (s *FinancialService) CreateBudget(
       ctx context.Context,
       name string,
       description string,
       totalCents int64,
       startDate, endDate time.Time,
   ) (*finances.Budget, error) {
       // Create budget
   }
   
   func (s *FinancialService) AllocateBudget(
       ctx context.Context,
       budgetID int64,
       simulationID *int64,
       sweepID *int64,
       amountCents int64,
   ) error {
       // Check budget availability
       // Create allocation
       // Update budget spent
   }
   
   func (s *FinancialService) CheckBudgetAvailability(
       ctx context.Context,
       budgetID int64,
       requiredCents int64,
   ) (bool, error) {
       // Check if budget has enough remaining
   }
   ```

2. Add budget enforcement to simulation creation

**Testing:**
- Test budget creation
- Test allocation
- Test budget exhaustion
- Test budget expiration

---

#### Task 4.3.2: Budget Endpoints
**Effort:** 4 hours  
**Priority:** Medium  
**Assignee:** Dev 1

**Description:**
Create budget management endpoints.

**Acceptance Criteria:**
- [ ] POST /api/v1/budgets
- [ ] GET /api/v1/budgets/:id
- [ ] GET /api/v1/budgets
- [ ] POST /api/v1/budgets/:id/allocate
- [ ] GET /api/v1/budgets/:id/allocations

**Subtasks:**
1. Create `internal/handlers/budgets.go`
2. Implement all CRUD endpoints
3. Wire into router

**Testing:**
- Test CRUD operations
- Test allocation tracking

---

### Sprint 4.4: Reports & Analytics (Remaining Time)

#### Task 4.4.1: Financial Reports Service
**Effort:** 5 hours  
**Priority:** Low  
**Assignee:** Dev 1

**Description:**
Generate financial reports and analytics.

**Acceptance Criteria:**
- [ ] Profit/loss report
- [ ] ROI trend analysis
- [ ] Cost breakdown
- [ ] Prize distribution

**Subtasks:**
1. Create `internal/services/financial_reports.go`:
   ```go
   type FinancialReportsService struct {
       financesQueries finances.Querier
   }
   
   type ProfitLossReport struct {
       Period         string
       TotalCost      int64
       TotalPrizes    int64
       NetProfit      int64
       ROI            float64
       Simulations    int
       BestSimulation int64
       WorstSimulation int64
   }
   
   func (s *FinancialReportsService) GenerateProfitLoss(
       ctx context.Context,
       dateFrom, dateTo time.Time,
   ) (*ProfitLossReport, error) {
       // Aggregate financial data
       // Calculate totals
       // Find best/worst
   }
   ```

2. Add trend analysis methods

**Testing:**
- Test report generation
- Test with various date ranges

---

#### Task 4.4.2: Financial Dashboard Endpoint
**Effort:** 3 hours  
**Priority:** Low  
**Assignee:** Dev 1

**Description:**
Create endpoint for financial dashboard data.

**Acceptance Criteria:**
- [ ] GET /api/v1/finances/dashboard
- [ ] Returns summary stats
- [ ] Returns recent activity

**Subtasks:**
1. Create endpoint
2. Aggregate dashboard data
3. Return JSON

**Testing:**
- Test endpoint
- Verify data accuracy

---

## Phase 4 Checklist

### Sprint 4.1 (Days 1-3)
- [ ] Task 4.1.1: Database schema
- [ ] Task 4.1.2: Financial queries
- [ ] Task 4.1.3: Financial service

### Sprint 4.2 (Days 4-5)
- [ ] Task 4.2.1: Integration with simulations
- [ ] Task 4.2.2: Financial endpoints

### Sprint 4.3 (Days 6-7)
- [ ] Task 4.3.1: Budget service
- [ ] Task 4.3.2: Budget endpoints

### Sprint 4.4 (Remaining)
- [ ] Task 4.4.1: Reports service
- [ ] Task 4.4.2: Dashboard endpoint

### Phase Gate
- [ ] All tasks completed
- [ ] Test coverage > 80%
- [ ] All tests passing
- [ ] Financial calculations verified
- [ ] Code reviewed
- [ ] Demo successful

---

## Metrics & KPIs

### Code Metrics
- **Lines of Code:** ~1800-2200
- **Test Coverage:** > 80%
- **Number of Tests:** > 40

### Data Metrics
- **Ledger Entry Accuracy:** 100%
- **ROI Calculation Accuracy:** ±0.01%
- **Budget Tracking Accuracy:** 100%

---

## Deliverables Summary

1. **Financial Tracking:** Complete ledger system
2. **Cost/Prize Calculation:** Accurate financial modeling
3. **ROI Analysis:** Profitability metrics
4. **Budget Management:** Budget creation and enforcement
5. **Financial Reports:** Analytics and insights

---

## Next Phase Preview

**Phase 5** will add:
- Dashboard API endpoints
- PDF report generation
- Charts and visualizations
- Email notifications

---

**Questions or Issues:**
Contact the development team or create an issue in the project tracker.
