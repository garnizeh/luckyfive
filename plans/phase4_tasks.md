# Phase 4: Financial Tracking — Detailed Tasks

**Duration:** 1 week (Week 7)  
**Estimated Effort:** 40 hours  
**Team:** 1 developer  
**Status:** Sprint 4.1 Complete ✅

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

#### Task 4.1.1: Finances Database Schema ✅
**Effort:** 4 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Define comprehensive financial tracking schema in finances.db.

**Acceptance Criteria:**
- [x] Ledger table created
- [x] Bet tracking tables created
- [x] Prize rules table created
- [x] Budget management tables created

**Implementation Status:** ✅ Complete
- Created comprehensive financial schema in `migrations/009_create_comprehensive_finances.sql`
- Added prize_rules table for Quina prize structure tracking
- Added bet_costs table for regional bet pricing
- Added ledger_entries table for detailed financial transactions
- Added simulation_finances table for financial summaries
- Added contest_bets table for tracking individual bets
- Added budgets and budget_allocations tables for budget management
- Created appropriate indexes for performance
- Inserted default bet cost and sample prize rules
- Migration successfully applied to finances.db

**Subtasks:**
1. Create `migrations/009_create_comprehensive_finances.sql` ✅ (comprehensive schema created)
2. Create migration runner (existing migrator used) ✅
3. Apply migration ✅ (successfully applied to all databases)

---

#### Task 4.1.2: Financial Queries (sqlc) ✅
**Effort:** 5 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Define sqlc queries for financial operations.

**Acceptance Criteria:**
- [x] Ledger CRUD queries
- [x] Financial summary queries
- [x] Prize rule queries
- [x] Budget queries
- [x] Querier interface generated

**Implementation Status:** ✅ Complete
- Created `internal/store/queries/finances.sql` with comprehensive SQL queries for all financial operations
- Updated `sqlc.yaml` to include both simulations and finances schemas for proper JOIN support
- Ran `make generate` successfully to create FinancesQuerier interface
- All CRUD operations defined for ledger, prizes, bets, finances, and budgets
- Queries include proper indexing and aggregation functions---

#### Task 4.1.3: Financial Service Implementation ✅
**Effort:** 8 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Implement FinancialService to manage all financial operations.

**Acceptance Criteria:**
- [x] Ledger management methods
- [x] Cost calculation methods
- [x] Prize calculation methods
- [x] ROI calculation methods
- [x] Uses Querier interface (mockable)

**Implementation Status:** ✅ Complete
- Created `internal/services/financial.go` with FinancialService struct and methods
- Implemented GetBetCost, GetPrize, CalculateSimulationFinances, GetFinancialSummary methods
- Added budget management methods (CreateBudget, AllocateBudget, CheckBudgetAvailability)
- Added ledger query methods (GetLedgerEntries, GetLedgerBalance)
- Created comprehensive unit tests in `internal/services/financial_test.go`
- Tests cover success and failure scenarios for all core methods
- All tests passing

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
- [x] Task 4.1.1: Database schema
- [x] Task 4.1.2: Financial queries
- [x] Task 4.1.3: Financial service

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
