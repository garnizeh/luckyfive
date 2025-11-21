# Phase 5: Reporting & Visualization — Detailed Tasks

**Duration:** 1 week (Week 8)  
**Estimated Effort:** 40 hours  
**Team:** 1-2 developers  
**Status:** Not Started

---

## Overview

Phase 5 implements comprehensive reporting and visualization capabilities including dashboard API endpoints, PDF report generation, data export, and chart data preparation. This phase makes simulation results accessible and actionable for stakeholders.

**Success Criteria:**
- ✅ Dashboard endpoints return aggregated data
- ✅ PDF reports generated successfully
- ✅ Charts render correctly with provided data
- ✅ Data export works in multiple formats
- ✅ Performance acceptable (< 2s for reports)
- ✅ Test coverage > 75%

---

## Dependencies

**From Previous Phases:**
- All simulation data available
- Financial tracking operational
- Comparison engine working

**External:**
- PDF generation library (e.g., go-pdf, wkhtmltopdf)
- Charting library decision (server-side vs client-side)

---

## Task Breakdown

### Sprint 5.1: Dashboard API (Days 1-3)

#### Task 5.1.1: Dashboard Data Service
**Effort:** 6 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Implement service to aggregate dashboard metrics from all databases.

**Acceptance Criteria:**
- [ ] Aggregates data from all 4 databases
- [ ] Caches results with TTL
- [ ] Returns structured dashboard data
- [ ] Performance optimized

**Subtasks:**
1. Create `internal/services/dashboard.go`:
   ```go
   package services
   
   import (
       "context"
       "sync"
       "time"
       
       "github.com/garnizeh/luckyfive/internal/store/simulations"
       "github.com/garnizeh/luckyfive/internal/store/finances"
       "github.com/garnizeh/luckyfive/internal/store/results"
   )
   
   type DashboardService struct {
       simulationsQueries simulations.Querier
       financesQueries    finances.Querier
       resultsQueries     results.Querier
       logger             *slog.Logger
       cache              *DashboardCache
       mu                 sync.RWMutex
   }
   
   type DashboardCache struct {
       Data      *DashboardData
       UpdatedAt time.Time
       TTL       time.Duration
   }
   
   type DashboardData struct {
       Overview        OverviewStats        `json:"overview"`
       RecentActivity  []ActivityItem       `json:"recent_activity"`
       TopPerformers   TopPerformersStats   `json:"top_performers"`
       FinancialStats  FinancialOverview    `json:"financial_stats"`
       SystemHealth    SystemHealthStats    `json:"system_health"`
       Charts          ChartData            `json:"charts"`
   }
   
   type OverviewStats struct {
       TotalSimulations      int     `json:"total_simulations"`
       CompletedSimulations  int     `json:"completed_simulations"`
       RunningSimulations    int     `json:"running_simulations"`
       PendingSimulations    int     `json:"pending_simulations"`
       TotalSweeps           int     `json:"total_sweeps"`
       TotalConfigurations   int     `json:"total_configurations"`
       TotalContestsImported int     `json:"total_contests_imported"`
       LatestContest         int     `json:"latest_contest"`
   }
   
   type ActivityItem struct {
       ID          int64     `json:"id"`
       Type        string    `json:"type"` // "simulation", "sweep", "comparison"
       Name        string    `json:"name"`
       Status      string    `json:"status"`
       CreatedAt   time.Time `json:"created_at"`
       CompletedAt *time.Time `json:"completed_at,omitempty"`
       DurationMs  int64     `json:"duration_ms,omitempty"`
   }
   
   type TopPerformersStats struct {
       ByQuinaRate    []SimulationSummary `json:"by_quina_rate"`
       ByROI          []SimulationSummary `json:"by_roi"`
       ByAvgHits      []SimulationSummary `json:"by_avg_hits"`
       BestAllTime    *SimulationSummary  `json:"best_all_time"`
   }
   
   type SimulationSummary struct {
       ID             int64   `json:"id"`
       Name           string  `json:"name"`
       Mode           string  `json:"mode"`
       QuinaRate      float64 `json:"quina_rate"`
       ROI            float64 `json:"roi"`
       AvgHits        float64 `json:"avg_hits"`
       NetProfitCents int64   `json:"net_profit_cents"`
       CreatedAt      string  `json:"created_at"`
   }
   
   type FinancialOverview struct {
       TotalInvestedCents  int64   `json:"total_invested_cents"`
       TotalPrizesCents    int64   `json:"total_prizes_cents"`
       TotalProfitCents    int64   `json:"total_profit_cents"`
       AverageROI          float64 `json:"average_roi"`
       ProfitableSimsPct   float64 `json:"profitable_sims_pct"`
       TotalQuinaWins      int     `json:"total_quina_wins"`
       TotalBudgets        int     `json:"total_budgets"`
       ActiveBudgetsCents  int64   `json:"active_budgets_cents"`
   }
   
   type SystemHealthStats struct {
       WorkerStatus       string  `json:"worker_status"`
       JobQueueSize       int     `json:"job_queue_size"`
       AvgSimDurationMs   int64   `json:"avg_sim_duration_ms"`
       DatabaseSizeMB     float64 `json:"database_size_mb"`
       LastImportDate     string  `json:"last_import_date"`
       SystemUptime       string  `json:"system_uptime"`
   }
   
   type ChartData struct {
       SimulationsOverTime  []TimeSeriesPoint `json:"simulations_over_time"`
       ROIDistribution      []DistributionBin `json:"roi_distribution"`
       ProfitByMonth        []TimeSeriesPoint `json:"profit_by_month"`
       HitRateComparison    []HitRateData     `json:"hit_rate_comparison"`
   }
   
   type TimeSeriesPoint struct {
       Date  string  `json:"date"`
       Value float64 `json:"value"`
       Count int     `json:"count,omitempty"`
   }
   
   type DistributionBin struct {
       Label string  `json:"label"`
       Min   float64 `json:"min"`
       Max   float64 `json:"max"`
       Count int     `json:"count"`
   }
   
   type HitRateData struct {
       Name       string  `json:"name"`
       QuinaRate  float64 `json:"quina_rate"`
       QuadraRate float64 `json:"quadra_rate"`
       TernoRate  float64 `json:"terno_rate"`
   }
   
   func NewDashboardService(
       simulationsQueries simulations.Querier,
       financesQueries finances.Querier,
       resultsQueries results.Querier,
       logger *slog.Logger,
   ) *DashboardService {
       return &DashboardService{
           simulationsQueries: simulationsQueries,
           financesQueries:    financesQueries,
           resultsQueries:     resultsQueries,
           logger:             logger,
           cache: &DashboardCache{
               TTL: 5 * time.Minute,
           },
       }
   }
   
   func (s *DashboardService) GetDashboard(ctx context.Context) (*DashboardData, error) {
       // Check cache
       s.mu.RLock()
       if s.cache.Data != nil && time.Since(s.cache.UpdatedAt) < s.cache.TTL {
           data := s.cache.Data
           s.mu.RUnlock()
           return data, nil
       }
       s.mu.RUnlock()
       
       // Rebuild dashboard data
       s.logger.Info("rebuilding dashboard cache")
       
       data := &DashboardData{}
       
       // Run queries in parallel
       var wg sync.WaitGroup
       errChan := make(chan error, 6)
       
       wg.Add(1)
       go func() {
           defer wg.Done()
           overview, err := s.buildOverviewStats(ctx)
           if err != nil {
               errChan <- err
               return
           }
           data.Overview = *overview
       }()
       
       wg.Add(1)
       go func() {
           defer wg.Done()
           activity, err := s.buildRecentActivity(ctx)
           if err != nil {
               errChan <- err
               return
           }
           data.RecentActivity = activity
       }()
       
       wg.Add(1)
       go func() {
           defer wg.Done()
           performers, err := s.buildTopPerformers(ctx)
           if err != nil {
               errChan <- err
               return
           }
           data.TopPerformers = *performers
       }()
       
       wg.Add(1)
       go func() {
           defer wg.Done()
           financial, err := s.buildFinancialOverview(ctx)
           if err != nil {
               errChan <- err
               return
           }
           data.FinancialStats = *financial
       }()
       
       wg.Add(1)
       go func() {
           defer wg.Done()
           health, err := s.buildSystemHealth(ctx)
           if err != nil {
               errChan <- err
               return
           }
           data.SystemHealth = *health
       }()
       
       wg.Add(1)
       go func() {
           defer wg.Done()
           charts, err := s.buildChartData(ctx)
           if err != nil {
               errChan <- err
               return
           }
           data.Charts = *charts
       }()
       
       wg.Wait()
       close(errChan)
       
       // Check for errors
       for err := range errChan {
           if err != nil {
               return nil, fmt.Errorf("dashboard build error: %w", err)
           }
       }
       
       // Update cache
       s.mu.Lock()
       s.cache.Data = data
       s.cache.UpdatedAt = time.Now()
       s.mu.Unlock()
       
       return data, nil
   }
   
   func (s *DashboardService) buildOverviewStats(ctx context.Context) (*OverviewStats, error) {
       // Count simulations by status
       completed, _ := s.simulationsQueries.CountSimulationsByStatus(ctx, "completed")
       running, _ := s.simulationsQueries.CountSimulationsByStatus(ctx, "running")
       pending, _ := s.simulationsQueries.CountSimulationsByStatus(ctx, "pending")
       
       total := completed + running + pending
       
       // Get latest contest
       latestContest, _ := s.resultsQueries.GetLatestContest(ctx)
       
       return &OverviewStats{
           TotalSimulations:      int(total),
           CompletedSimulations:  int(completed),
           RunningSimulations:    int(running),
           PendingSimulations:    int(pending),
           LatestContest:         int(latestContest),
       }, nil
   }
   
   func (s *DashboardService) buildRecentActivity(ctx context.Context) ([]ActivityItem, error) {
       // Get recent simulations
       sims, err := s.simulationsQueries.ListSimulations(ctx, simulations.ListSimulationsParams{
           Limit:  10,
           Offset: 0,
       })
       if err != nil {
           return nil, err
       }
       
       activity := make([]ActivityItem, 0, len(sims))
       for _, sim := range sims {
           item := ActivityItem{
               ID:        sim.ID,
               Type:      "simulation",
               Name:      sim.RecipeName.String,
               Status:    sim.Status,
               CreatedAt: parseTime(sim.CreatedAt),
           }
           
           if sim.FinishedAt.Valid {
               t := parseTime(sim.FinishedAt.String)
               item.CompletedAt = &t
           }
           
           if sim.RunDurationMs.Valid {
               item.DurationMs = sim.RunDurationMs.Int64
           }
           
           activity = append(activity, item)
       }
       
       return activity, nil
   }
   
   func (s *DashboardService) buildTopPerformers(ctx context.Context) (*TopPerformersStats, error) {
       // Get top by ROI
       topROI, err := s.financesQueries.ListTopSimulationsByROI(ctx, finances.ListTopSimulationsByROIParams{
           Limit:  5,
           Offset: 0,
       })
       if err != nil {
           return nil, err
       }
       
       performers := &TopPerformersStats{
           ByROI: make([]SimulationSummary, 0, len(topROI)),
       }
       
       for _, sim := range topROI {
           performers.ByROI = append(performers.ByROI, SimulationSummary{
               ID:             sim.SimulationID,
               Name:           sim.RecipeName.String,
               Mode:           sim.Mode,
               ROI:            sim.RoiPercentage,
               NetProfitCents: sim.NetProfitCents,
           })
       }
       
       // TODO: Add by quina rate, avg hits
       
       return performers, nil
   }
   
   func (s *DashboardService) buildFinancialOverview(ctx context.Context) (*FinancialOverview, error) {
       // Aggregate financial data
       // This would require additional queries or aggregate functions
       
       return &FinancialOverview{
           // Placeholder values - implement proper aggregation
           TotalInvestedCents: 0,
           TotalPrizesCents:   0,
           TotalProfitCents:   0,
           AverageROI:         0,
       }, nil
   }
   
   func (s *DashboardService) buildSystemHealth(ctx context.Context) (*SystemHealthStats, error) {
       // Get worker status
       pendingCount, _ := s.simulationsQueries.CountSimulationsByStatus(ctx, "pending")
       
       return &SystemHealthStats{
           WorkerStatus: "healthy",
           JobQueueSize: int(pendingCount),
       }, nil
   }
   
   func (s *DashboardService) buildChartData(ctx context.Context) (*ChartData, error) {
       // Build chart datasets
       // This would involve time-series queries
       
       return &ChartData{
           SimulationsOverTime: []TimeSeriesPoint{},
           ROIDistribution:     []DistributionBin{},
       }, nil
   }
   
   func parseTime(s string) time.Time {
       t, _ := time.Parse(time.RFC3339, s)
       return t
   }
   ```

2. Add additional aggregate queries to sqlc if needed

**Testing:**
- Mock all Querier interfaces
- Test cache behavior
- Test parallel execution
- Test error handling

---

#### Task 5.1.2: Dashboard Endpoint
**Effort:** 3 hours  
**Priority:** Critical  
**Assignee:** Dev 2

**Description:**
Implement dashboard API endpoint.

**Acceptance Criteria:**
- [ ] GET /api/v1/dashboard
- [ ] Returns cached data
- [ ] Optional ?refresh=true param
- [ ] Proper error handling

**Subtasks:**
1. Create `internal/handlers/dashboard.go`:
   ```go
   func GetDashboard(dashSvc *services.DashboardService) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           // Check if refresh requested
           refresh := r.URL.Query().Get("refresh") == "true"
           
           if refresh {
               dashSvc.InvalidateCache()
           }
           
           data, err := dashSvc.GetDashboard(r.Context())
           if err != nil {
               WriteError(w, 500, err)
               return
           }
           
           WriteJSON(w, 200, data)
       }
   }
   ```

2. Wire into router

**Testing:**
- Test endpoint
- Test cache behavior
- Test refresh parameter

---

#### Task 5.1.3: Chart Data Endpoints
**Effort:** 4 hours  
**Priority:** High  
**Assignee:** Dev 2

**Description:**
Implement specific chart data endpoints for detailed visualizations.

**Acceptance Criteria:**
- [ ] GET /api/v1/charts/simulations-timeline
- [ ] GET /api/v1/charts/roi-distribution
- [ ] GET /api/v1/charts/performance-comparison
- [ ] GET /api/v1/charts/profit-trend

**Subtasks:**
1. Create `internal/handlers/charts.go`
2. Implement each chart endpoint
3. Add query parameters for filtering (date range, mode, etc.)

**Testing:**
- Test all endpoints
- Test filtering
- Verify data format

---

### Sprint 5.2: PDF Report Generation (Days 4-6)

#### Task 5.2.1: PDF Library Integration
**Effort:** 3 hours  
**Priority:** High  
**Assignee:** Dev 1

**Description:**
Integrate PDF generation library and create base templates.

**Acceptance Criteria:**
- [ ] PDF library installed and configured
- [ ] Base template created
- [ ] Header/footer support
- [ ] Table rendering working

**Subtasks:**
1. Choose PDF library:
   - Option A: `github.com/jung-kurt/gofpdf` (pure Go)
   - Option B: `github.com/SebastiaanKlippert/go-wkhtmltopdf` (wraps wkhtmltopdf)
   - Recommendation: gofpdf for simpler deployment

2. Install dependency:
   ```bash
   go get github.com/jung-kurt/gofpdf
   ```

3. Create `pkg/reports/pdf.go`:
   ```go
   package reports
   
   import (
       "github.com/jung-kurt/gofpdf"
   )
   
   type PDFReport struct {
       pdf *gofpdf.Fpdf
   }
   
   func NewPDFReport() *PDFReport {
       pdf := gofpdf.New("P", "mm", "A4", "")
       pdf.SetAutoPageBreak(true, 10)
       
       return &PDFReport{pdf: pdf}
   }
   
   func (r *PDFReport) AddHeader(title, subtitle string) {
       r.pdf.AddPage()
       r.pdf.SetFont("Arial", "B", 20)
       r.pdf.Cell(0, 10, title)
       r.pdf.Ln(10)
       
       if subtitle != "" {
           r.pdf.SetFont("Arial", "", 12)
           r.pdf.Cell(0, 6, subtitle)
           r.pdf.Ln(8)
       }
   }
   
   func (r *PDFReport) AddSection(heading string) {
       r.pdf.SetFont("Arial", "B", 14)
       r.pdf.Cell(0, 8, heading)
       r.pdf.Ln(8)
   }
   
   func (r *PDFReport) AddText(text string) {
       r.pdf.SetFont("Arial", "", 11)
       r.pdf.MultiCell(0, 5, text, "", "", false)
       r.pdf.Ln(5)
   }
   
   func (r *PDFReport) AddTable(headers []string, rows [][]string) {
       r.pdf.SetFont("Arial", "B", 10)
       r.pdf.SetFillColor(200, 200, 200)
       
       // Calculate column widths
       colWidth := 190.0 / float64(len(headers))
       
       // Headers
       for _, header := range headers {
           r.pdf.CellFormat(colWidth, 7, header, "1", 0, "C", true, 0, "")
       }
       r.pdf.Ln(-1)
       
       // Rows
       r.pdf.SetFont("Arial", "", 9)
       r.pdf.SetFillColor(255, 255, 255)
       
       for _, row := range rows {
           for _, cell := range row {
               r.pdf.CellFormat(colWidth, 6, cell, "1", 0, "L", false, 0, "")
           }
           r.pdf.Ln(-1)
       }
       
       r.pdf.Ln(5)
   }
   
   func (r *PDFReport) SaveToFile(filename string) error {
       return r.pdf.OutputFileAndClose(filename)
   }
   
   func (r *PDFReport) Bytes() ([]byte, error) {
       var buf bytes.Buffer
       err := r.pdf.Output(&buf)
       return buf.Bytes(), err
   }
   ```

**Testing:**
- Test basic PDF generation
- Test tables
- Verify formatting

---

#### Task 5.2.2: Simulation Report Generator
**Effort:** 8 hours  
**Priority:** High  
**Assignee:** Dev 1

**Description:**
Implement comprehensive simulation report generation.

**Acceptance Criteria:**
- [ ] Simulation summary section
- [ ] Contest results table
- [ ] Financial summary
- [ ] Performance metrics
- [ ] Charts (if possible)

**Subtasks:**
1. Create `pkg/reports/simulation_report.go`:
   ```go
   package reports
   
   import (
       "fmt"
       "time"
       
       "github.com/garnizeh/luckyfive/internal/services"
   )
   
   type SimulationReportData struct {
       Simulation      services.Simulation
       Summary         services.Summary
       FinancialSummary services.FinancialSummary
       ContestResults  []services.ContestResult
       TopPredictions  []services.ContestResult
   }
   
   type SimulationReportGenerator struct{}
   
   func NewSimulationReportGenerator() *SimulationReportGenerator {
       return &SimulationReportGenerator{}
   }
   
   func (g *SimulationReportGenerator) Generate(data SimulationReportData) ([]byte, error) {
       pdf := NewPDFReport()
       
       // Title page
       pdf.AddHeader(
           "Simulation Report",
           fmt.Sprintf("Simulation #%d - %s", data.Simulation.ID, data.Simulation.RecipeName),
       )
       
       pdf.AddText(fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
       pdf.AddText(fmt.Sprintf("Mode: %s", data.Simulation.Mode))
       pdf.AddText(fmt.Sprintf("Contest Range: %d - %d",
           data.Simulation.StartContest, data.Simulation.EndContest))
       
       // Executive Summary
       pdf.AddSection("Executive Summary")
       pdf.AddText(fmt.Sprintf("Total Contests: %d", data.Summary.TotalContests))
       pdf.AddText(fmt.Sprintf("Average Hits: %.2f", data.Summary.AverageHits))
       pdf.AddText(fmt.Sprintf("Quina Hit Rate: %.2f%%", data.Summary.HitRateQuina*100))
       pdf.AddText(fmt.Sprintf("Quadra Hit Rate: %.2f%%", data.Summary.HitRateQuadra*100))
       pdf.AddText(fmt.Sprintf("Terno Hit Rate: %.2f%%", data.Summary.HitRateTerno*100))
       
       // Financial Summary
       pdf.AddSection("Financial Summary")
       pdf.AddText(fmt.Sprintf("Total Investment: R$ %.2f",
           float64(data.FinancialSummary.TotalCostCents)/100))
       pdf.AddText(fmt.Sprintf("Total Prizes: R$ %.2f",
           float64(data.FinancialSummary.TotalPrizesCents)/100))
       pdf.AddText(fmt.Sprintf("Net Profit: R$ %.2f",
           float64(data.FinancialSummary.NetProfitCents)/100))
       pdf.AddText(fmt.Sprintf("ROI: %.2f%%", data.FinancialSummary.ROIPercentage))
       
       if data.FinancialSummary.BreakEvenContest > 0 {
           pdf.AddText(fmt.Sprintf("Break-even Contest: %d",
               data.FinancialSummary.BreakEvenContest))
       }
       
       // Top Results
       pdf.AddSection("Top 10 Results")
       
       headers := []string{"Contest", "Actual Numbers", "Prediction", "Hits"}
       rows := [][]string{}
       
       for i, result := range data.TopPredictions {
           if i >= 10 {
               break
           }
           
           rows = append(rows, []string{
               fmt.Sprintf("%d", result.Contest),
               formatNumbers(result.ActualNumbers),
               formatNumbers(result.BestPrediction),
               fmt.Sprintf("%d", result.BestHits),
           })
       }
       
       pdf.AddTable(headers, rows)
       
       // Recipe Configuration
       pdf.AddSection("Configuration")
       pdf.AddText("Recipe JSON:")
       pdf.AddText(data.Simulation.RecipeJson)
       
       return pdf.Bytes()
   }
   
   func formatNumbers(nums []int) string {
       strs := make([]string, len(nums))
       for i, n := range nums {
           strs[i] = fmt.Sprintf("%02d", n)
       }
       return strings.Join(strs, ", ")
   }
   ```

2. Add chart generation (optional, using external library like go-chart)

**Testing:**
- Test with sample simulation data
- Verify all sections present
- Check formatting

---

#### Task 5.2.3: Report Generation Endpoint
**Effort:** 4 hours  
**Priority:** High  
**Assignee:** Dev 2

**Description:**
Implement endpoint to generate and download PDF reports.

**Acceptance Criteria:**
- [ ] GET /api/v1/simulations/:id/report.pdf
- [ ] Generates PDF on-demand
- [ ] Proper content-type headers
- [ ] Optional caching

**Subtasks:**
1. Create `internal/handlers/reports.go`:
   ```go
   func DownloadSimulationReport(
       simSvc *services.SimulationService,
       finSvc *services.FinancialService,
       reportGen *reports.SimulationReportGenerator,
   ) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           simID := chi.URLParam(r, "id")
           id, _ := strconv.ParseInt(simID, 10, 64)
           
           // Fetch simulation
           sim, err := simSvc.GetSimulation(r.Context(), id)
           if err != nil {
               WriteError(w, 404, err)
               return
           }
           
           // Fetch financial summary
           finSummary, err := finSvc.GetFinancialSummary(r.Context(), id)
           if err != nil {
               WriteError(w, 500, err)
               return
           }
           
           // Fetch contest results
           results, err := simSvc.GetContestResults(r.Context(), id)
           if err != nil {
               WriteError(w, 500, err)
               return
           }
           
           // Generate report
           reportData := reports.SimulationReportData{
               Simulation:       *sim,
               FinancialSummary: *finSummary,
               ContestResults:   results,
               TopPredictions:   getTopResults(results, 20),
           }
           
           pdfBytes, err := reportGen.Generate(reportData)
           if err != nil {
               WriteError(w, 500, err)
               return
           }
           
           // Send PDF
           w.Header().Set("Content-Type", "application/pdf")
           w.Header().Set("Content-Disposition",
               fmt.Sprintf("attachment; filename=simulation_%d.pdf", id))
           w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
           
           w.Write(pdfBytes)
       }
   }
   ```

2. Wire into router

**Testing:**
- Test report download
- Verify PDF opens correctly
- Test with various simulations

---

### Sprint 5.3: Data Export (Days 7)

#### Task 5.3.1: CSV Export Service
**Effort:** 5 hours  
**Priority:** Medium  
**Assignee:** Dev 1

**Description:**
Implement CSV export for simulation results.

**Acceptance Criteria:**
- [ ] Export contest results to CSV
- [ ] Export ledger to CSV
- [ ] Export sweep results to CSV
- [ ] Proper CSV formatting

**Subtasks:**
1. Create `pkg/export/csv.go`:
   ```go
   package export
   
   import (
       "encoding/csv"
       "fmt"
       "io"
   )
   
   type CSVExporter struct{}
   
   func NewCSVExporter() *CSVExporter {
       return &CSVExporter{}
   }
   
   func (e *CSVExporter) ExportContestResults(
       w io.Writer,
       results []services.ContestResult,
   ) error {
       writer := csv.NewWriter(w)
       defer writer.Flush()
       
       // Headers
       headers := []string{
           "Contest", "Actual Numbers", "Best Prediction",
           "Hits", "Prize Type", "Prize Amount",
       }
       writer.Write(headers)
       
       // Rows
       for _, result := range results {
           row := []string{
               fmt.Sprintf("%d", result.Contest),
               formatNumbers(result.ActualNumbers),
               formatNumbers(result.BestPrediction),
               fmt.Sprintf("%d", result.BestHits),
               result.PrizeType,
               fmt.Sprintf("%.2f", float64(result.PrizeCents)/100),
           }
           writer.Write(row)
       }
       
       return nil
   }
   
   func (e *CSVExporter) ExportLedger(
       w io.Writer,
       entries []finances.LedgerEntry,
   ) error {
       writer := csv.NewWriter(w)
       defer writer.Flush()
       
       headers := []string{
           "Date", "Type", "Amount", "Contest", "Description",
       }
       writer.Write(headers)
       
       for _, entry := range entries {
           row := []string{
               entry.TransactionDate,
               entry.TransactionType,
               fmt.Sprintf("%.2f", float64(entry.AmountCents)/100),
               fmt.Sprintf("%d", entry.Contest.Int64),
               entry.Description.String,
           }
           writer.Write(row)
       }
       
       return nil
   }
   ```

2. Add JSON export

**Testing:**
- Test CSV generation
- Verify format
- Test with large datasets

---

#### Task 5.3.2: Export Endpoints
**Effort:** 4 hours  
**Priority:** Medium  
**Assignee:** Dev 2

**Description:**
Implement export endpoints.

**Acceptance Criteria:**
- [ ] GET /api/v1/simulations/:id/export.csv
- [ ] GET /api/v1/simulations/:id/export.json
- [ ] GET /api/v1/finances/ledger.csv
- [ ] Proper content-type headers

**Subtasks:**
1. Add export handlers
2. Wire into router
3. Add format parameter

**Testing:**
- Test all formats
- Test downloads
- Verify data integrity

---

### Sprint 5.4: Notifications & Polish (Remaining Time)

#### Task 5.4.1: Email Notification Service (Optional)
**Effort:** 6 hours  
**Priority:** Low  
**Assignee:** Dev 1

**Description:**
Implement email notifications for simulation completion.

**Acceptance Criteria:**
- [ ] Email service configured
- [ ] Template system
- [ ] Send on simulation complete
- [ ] Optional: daily digest

**Subtasks:**
1. Choose email library (e.g., gomail)
2. Create templates
3. Integrate with simulation service

**Testing:**
- Test email sending
- Test templates

---

#### Task 5.4.2: Webhook Support (Optional)
**Effort:** 4 hours  
**Priority:** Low  
**Assignee:** Dev 2

**Description:**
Add webhook support for external integrations.

**Acceptance Criteria:**
- [ ] Webhook configuration
- [ ] Event triggers
- [ ] Retry logic
- [ ] Signature verification

**Subtasks:**
1. Create webhook service
2. Add event emitters
3. Add configuration endpoints

**Testing:**
- Test webhook delivery
- Test retry logic

---

## Phase 5 Checklist

### Sprint 5.1 (Days 1-3)
- [ ] Task 5.1.1: Dashboard service
- [ ] Task 5.1.2: Dashboard endpoint
- [ ] Task 5.1.3: Chart endpoints

### Sprint 5.2 (Days 4-6)
- [ ] Task 5.2.1: PDF library integration
- [ ] Task 5.2.2: Simulation report generator
- [ ] Task 5.2.3: Report endpoint

### Sprint 5.3 (Day 7)
- [ ] Task 5.3.1: CSV export service
- [ ] Task 5.3.2: Export endpoints

### Sprint 5.4 (Remaining)
- [ ] Task 5.4.1: Email notifications (optional)
- [ ] Task 5.4.2: Webhooks (optional)

### Phase Gate
- [ ] All critical tasks completed
- [ ] Test coverage > 75%
- [ ] All tests passing
- [ ] PDF reports working
- [ ] Dashboard functional
- [ ] Code reviewed
- [ ] Demo successful

---

## Metrics & KPIs

### Code Metrics
- **Lines of Code:** ~1600-2000
- **Test Coverage:** > 75%
- **Number of Tests:** > 30

### Performance Metrics
- **Dashboard Load Time:** < 500ms (cached)
- **PDF Generation Time:** < 2s
- **CSV Export Time:** < 1s for 1000 rows

---

## Deliverables Summary

1. **Dashboard API:** Comprehensive metrics and analytics
2. **PDF Reports:** Professional simulation reports
3. **Chart Data:** Ready-to-render visualization data
4. **Data Export:** CSV and JSON export capabilities
5. **Notifications:** Email alerts (optional)

---

## Next Phase Preview

**Phase 6** will focus on:
- Performance optimization
- Production hardening
- Deployment automation
- Final documentation
- User acceptance testing

---

**Questions or Issues:**
Contact the development team or create an issue in the project tracker.
