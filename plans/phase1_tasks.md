# Phase 1: Foundation & Database — Detailed Tasks

**Duration:** 2 weeks (Weeks 1-2)  
**Estimated Effort:** 80 hours  
**Team:** 1-2 developers  
**Status:** In Progress — Sprint 1.1 started (Tasks 1.1.1, 1.1.2 and 1.1.3 completed)

## Progress Update

Current state (sprint 1.1):

- Task 1.1.1: Project initialization — completed. Directories created and initial scaffold (including `.gitignore`, `README.md`, `LICENSE`) added and committed to the repository.
- Task 1.1.2: `Makefile` created and committed (basic build/test/generate targets). A basic verification was performed: `go mod tidy` ran and `go test ./...` completed without failures. Some `build` targets which rely on fully implemented `cmd/*/main.go` will be verified as those entrypoints are implemented.
- Task 1.1.3: Dependencies and developer tools installed and verified. Commands executed included `go get` additions, `go install` for `sqlc`, `mockgen`, and `golangci-lint`, `go mod tidy`, and quick build/test verification. See Task 1.1.3 section for tool versions and details.

Commits of note:
- 82dbd23 — initial project scaffold (created `.gitignore`, `LICENSE`, `README.md`, and initial `cmd/` files)
- f9c0b80 — added `Makefile` with build/test/generate targets
- 92acdda — docs: mark Task 1.1.3 done (dependencies & dev tools installed)
- 527d9e3 — test(config/logger): add invalid concurrency and logger level tests

Next immediate steps: configure sqlc + generate code & mocks (`Task 1.1.5`).
Recent work: production-ready migrations were added for results, simulations, configs and finances; SQL query files were expanded with additional useful queries; sqlc generation and mock generation were executed and mocks were added to `internal/store/*/mock/`.

Recent small wins:

- Task 1.1.4 (Logging & Configuration) implemented. Unit tests for `internal/config` and `internal/logger` were added and run locally (they pass). See commit `527d9e3` which added the tests.

---

## Overview

Phase 1 establishes the foundation for the entire project. This phase focuses on:
- Project structure and build system
- Database infrastructure (4 SQLite databases)
- Import functionality (XLSX → SQLite)
- Basic HTTP API setup
- Testing framework

**Success Criteria:**
- ✅ Can import 5000+ lottery results from XLSX in < 30 seconds
- ✅ All 4 databases created with proper indexes
- ✅ Health endpoint returns 200 OK
- ✅ Test coverage > 70%
- ✅ Build system works (Makefile targets)

---

## Task Breakdown

### Sprint 1.1: Project Setup & Infrastructure (Days 1-3)

#### Task 1.1.1: Project Initialization
**Effort:** 2 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Initialize Go module and create basic project structure following the design document.

**Acceptance Criteria:**
- [x] Go module initialized (`go mod init`)
- [x] Project follows structure from design doc
- [x] `.gitignore` configured for Go and SQLite files
- [x] README.md with setup instructions
- [x] License file added (if applicable)

**Notes:** The Go module (`go.mod`) was present in the workspace and remains unchanged. The created directory structure matches the planned layout in the phase plan.

**Subtasks:**
1. Run `go mod init github.com/garnizeh/luckyfive`
2. Create directory structure:
   ```
   cmd/
     api/
     worker/
     migrate/
   internal/
     handlers/
     services/
     store/
     middleware/
     models/
   pkg/
     predictor/
     utils/
   data/
   migrations/
   configs/
   docs/
   ```
3. Create `.gitignore`:
   ```
   # Binaries
   /cmd/*/main
   *.exe
   
   # Data
   /data/*.db
   /data/simulations/
   
   # IDE
   .vscode/
   .idea/
   
   # OS
   .DS_Store
   ```
4. Create basic `README.md`
5. Commit initial structure

**Testing:**
- Directory structure matches design
- `go mod tidy` runs without errors

---

#### Task 1.1.2: Makefile & Build System
**Effort:** 3 hours  
**Priority:** High  
**Assignee:** Dev 1

**Description:**
Create Makefile with targets for building, testing, running, and managing the project.

**Acceptance Criteria:**
- [ ] `make build` compiles all binaries (pending verification as cmd entrypoints are implemented)
- [x] `make test` runs all tests (basic verification performed using `go test ./...`)
- [ ] `make run-api` starts HTTP server (requires `cmd/api/main.go` implementation)
- [ ] `make migrate` runs database migrations (requires `cmd/migrate/main.go` and migrations)
- [ ] `make clean` removes build artifacts
- [ ] `make lint` runs linters (depends on `golangci-lint` being installed)

**Subtasks:**
1. Create `Makefile` with targets:
   ```makefile
   .PHONY: build test run-api run-worker migrate clean lint
   
   build:
       @echo "Building binaries..."
       @go build -o bin/api cmd/api/main.go
       @go build -o bin/worker cmd/worker/main.go
       @go build -o bin/migrate cmd/migrate/main.go
   
   test:
       @echo "Running tests..."
       @go test -v -cover ./...
   
   test-coverage:
       @go test -v -coverprofile=coverage.out ./...
       @go tool cover -html=coverage.out -o coverage.html
   
   run-api:
       @go run cmd/api/main.go
   
   run-worker:
       @go run cmd/worker/main.go
   
   migrate:
       @go run cmd/migrate/main.go up
   
   clean:
       @rm -rf bin/
       @rm -f coverage.out coverage.html
   
   lint:
       @golangci-lint run
   ```
2. Test each target
3. Document in README

**Testing:**
- All make targets execute without errors
- Binaries are created in `bin/` directory

---

#### Task 1.1.3: Dependencies Installation
**Effort:** 2 hours  
**Priority:** High  
**Assignee:** Dev 1

**Description:**
Add required Go dependencies to `go.mod`.

**Acceptance Criteria:**
- [x] All required libraries added
- [x] `go.mod` and `go.sum` committed
- [x] No version conflicts detected during `go mod tidy`
- [x] Dependencies documented in README

**Completed:** Task verified and committed on 2025-11-21. See commit 92acdda for the doc update and go.mod/go.sum changes.

**Subtasks:**
1. Install dependencies:
   ```bash
   # Database
   go get modernc.org/sqlite
   
   # HTTP router
   go get github.com/go-chi/chi/v5
   
   # XLSX parsing
   go get github.com/xuri/excelize/v2
   
   # Testing
   go get github.com/stretchr/testify
   
   # Mocking
   go get go.uber.org/mock/mockgen
   
   # Logging
   # (stdlib log/slog - no install needed)
   
   # Validation
   go get github.com/go-playground/validator/v10
   ```
2. Install sqlc for code generation:
   ```bash
   # Install sqlc binary
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   
   # Verify installation
   sqlc version
   ```
3. Run `go mod tidy`
4. Verify all imports work

**Testing:**
- `go mod verify` passes
- `go build ./...` succeeds
- `sqlc version` returns version info

**Progress & Verification:**

The dependency installation and developer tool installation were performed locally and verified. Commands executed included adding libraries with `go get`, installing developer tools (`sqlc`, `mockgen`, `golangci-lint`) via `go install`, running `go mod tidy`, and performing quick build/test verification.

Observed tool versions on the environment where the commands were run:

- `sqlc version`: v1.30.0
- `mockgen -version`: v1.6.0
- `golangci-lint version`: golangci-lint has version v1.64.8
- `go version`: go1.25.4 linux/amd64

All commands completed without errors. `go.mod` and `go.sum` were updated and committed.

**Documentation:** The project's `README.md` was updated to document the installed Go libraries and developer tools (sqlc, mockgen, golangci-lint) and include install/run commands. See commit e4668ed.

---


#### Task 1.1.4: Logging & Configuration Setup
**Effort:** 3 hours  
**Priority:** Medium  
**Assignee:** Dev 1

**Description:**
Set up structured logging and configuration management.

**Acceptance Criteria:**
- [x] Structured logging configured (slog)
- [x] Log levels configurable (DEBUG, INFO, WARN, ERROR)
- [x] Configuration loaded from environment variables
- [x] Default config file created (`.env.example`)

**Subtasks completed:**
1. `internal/config/config.go` added — provides `Config` struct and `Load()` which reads environment variables with sensible defaults.
2. `internal/logger/logger.go` added — provides `New(level string) *slog.Logger` returning a JSON handler logger.
3. `.env.example` added at project root with recommended environment variables.

**Testing / Verification:**
- Package builds and basic `go test ./...` ran successfully (packages without tests reported as such).


---

#### Task 1.1.5: sqlc Configuration
**Effort:** 3 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Configure sqlc for type-safe database code generation with mockable Querier interfaces.

**Acceptance Criteria:**
- [ ] `sqlc.yaml` configured for all 4 databases
- [ ] Querier interfaces generated
- [ ] Mock generation configured
- [ ] Make target for code generation

**Subtasks:**
1. Create `sqlc.yaml` in project root:
   ```yaml
   version: "2"
   sql:
     - schema: "migrations/001_create_results.sql"
       queries: "internal/store/queries/results.sql"
       engine: "sqlite"
       gen:
         go:
           package: "results"
           out: "internal/store/results"
           emit_interface: true
           emit_json_tags: true
           emit_prepared_queries: false
           emit_exact_table_names: false
           query_parameter_limit: 5
     
     - schema: "migrations/002_create_simulations.sql"
       queries: "internal/store/queries/simulations.sql"
       engine: "sqlite"
       gen:
         go:
           package: "simulations"
           out: "internal/store/simulations"
           emit_interface: true
           emit_json_tags: true
           emit_prepared_queries: false
           emit_exact_table_names: false
     
     - schema: "migrations/003_create_configs.sql"
       queries: "internal/store/queries/configs.sql"
       engine: "sqlite"
       gen:
         go:
           package: "configs"
           out: "internal/store/configs"
           emit_interface: true
           emit_json_tags: true
           emit_prepared_queries: false
           emit_exact_table_names: false
     
     - schema: "migrations/004_create_finances.sql"
       queries: "internal/store/queries/finances.sql"
       engine: "sqlite"
       gen:
         go:
           package: "finances"
           out: "internal/store/finances"
           emit_interface: true
           emit_json_tags: true
           emit_prepared_queries: false
           emit_exact_table_names: false
   ```

2. Create directory structure:
   ```bash
   mkdir -p internal/store/queries
   mkdir -p internal/store/results
   mkdir -p internal/store/simulations
   mkdir -p internal/store/configs
   mkdir -p internal/store/finances
   ```

3. Add sqlc targets to Makefile:
   ```makefile
   .PHONY: sqlc sqlc-generate mock-generate
   
   sqlc-generate:
       @echo "Generating sqlc code..."
       @sqlc generate
   
   mock-generate: sqlc-generate
       @echo "Generating mocks..."
       @mockgen -source=internal/store/results/querier.go -destination=internal/store/results/mock/querier.go -package=mock
       @mockgen -source=internal/store/simulations/querier.go -destination=internal/store/simulations/mock/querier.go -package=mock
       @mockgen -source=internal/store/configs/querier.go -destination=internal/store/configs/mock/querier.go -package=mock
       @mockgen -source=internal/store/finances/querier.go -destination=internal/store/finances/mock/querier.go -package=mock
   
   generate: sqlc-generate mock-generate
   ```

4. Create `.sqlc` directory for generated files tracking (add to .gitignore if needed)

5. Document sqlc workflow in README:
   ```markdown
   ## Code Generation
   
   This project uses sqlc for type-safe database queries.
   
   ### Generate code:
   ```bash
   make generate  # Generates sqlc code + mocks
   ```
   
   ### After modifying SQL:
   1. Update queries in `internal/store/queries/*.sql`
   2. Run `make generate`
   3. Update tests to use new generated code
   ```

**Testing:**
- `make generate` runs without errors
- Querier interfaces generated in each package
- Mock files created successfully
- Generated code compiles

---

### Sprint 1.2: Database Infrastructure (Days 4-7)

#### Task 1.2.1: Migration System
**Effort:** 4 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Create database migration system for managing schema versions.

**Acceptance Criteria:**
- [ ] Migration runner implemented
- [ ] Up/down migrations supported
- [ ] Migration tracking table created
- [ ] Idempotent migrations (can run multiple times)

**Subtasks:**
1. Create `internal/store/migrate.go`:
   ```go
   type Migrator struct {
       db *sql.DB
   }
   
   func NewMigrator(db *sql.DB) *Migrator
   func (m *Migrator) Up() error
   func (m *Migrator) Down() error
   func (m *Migrator) Version() (int, error)
   ```
2. Create migration tracking table:
   ```sql
   CREATE TABLE IF NOT EXISTS schema_migrations (
       version INTEGER PRIMARY KEY,
       applied_at TEXT NOT NULL
   );
   ```
3. Create `cmd/migrate/main.go`
4. Implement migration file loading from `migrations/`

**Testing:**
- Can run migrations up and down
- Version tracking works correctly
- Re-running migrations is safe

---

#### Task 1.2.2: Results Database Schema + Queries
**Effort:** 4 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Create migrations for `results.db` and SQL queries for sqlc generation.

**Acceptance Criteria:**
- [ ] `draws` table created with constraints
- [ ] `import_history` table created
- [ ] Indexes added for performance
- [ ] SQL queries defined for sqlc
- [ ] Migration file tested

**Subtasks:**
1. Create `migrations/001_create_results.sql`:
   ```sql
   -- Up migration
   CREATE TABLE IF NOT EXISTS draws (
     contest INTEGER PRIMARY KEY,
     draw_date TEXT NOT NULL,
     bola1 INTEGER NOT NULL CHECK(bola1 BETWEEN 1 AND 80),
     bola2 INTEGER NOT NULL CHECK(bola2 BETWEEN 1 AND 80),
     bola3 INTEGER NOT NULL CHECK(bola3 BETWEEN 1 AND 80),
     bola4 INTEGER NOT NULL CHECK(bola4 BETWEEN 1 AND 80),
     bola5 INTEGER NOT NULL CHECK(bola5 BETWEEN 1 AND 80),
     source TEXT,
     imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
     raw_row TEXT,
     
     CHECK(bola1 < bola2 AND bola2 < bola3 AND bola3 < bola4 AND bola4 < bola5)
   );
   
   CREATE INDEX IF NOT EXISTS idx_draws_date ON draws(draw_date);
   CREATE INDEX IF NOT EXISTS idx_draws_imported ON draws(imported_at);
   
   CREATE TABLE IF NOT EXISTS import_history (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     filename TEXT NOT NULL,
     imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
     rows_inserted INTEGER NOT NULL,
     rows_skipped INTEGER NOT NULL,
     rows_errors INTEGER NOT NULL,
     source_hash TEXT,
     metadata TEXT
   );
   
   -- Down migration
   -- DROP TABLE IF EXISTS draws;
   -- DROP TABLE IF EXISTS import_history;
   ```

2. Create `internal/store/queries/results.sql`:
   ```sql
   -- name: GetDraw :one
   SELECT * FROM draws
   WHERE contest = ?
   LIMIT 1;
   
   -- name: ListDraws :many
   SELECT * FROM draws
   ORDER BY contest DESC
   LIMIT ? OFFSET ?;
   
   -- name: ListDrawsByDateRange :many
   SELECT * FROM draws
   WHERE draw_date BETWEEN ? AND ?
   ORDER BY contest DESC
   LIMIT ? OFFSET ?;
   
   -- name: ListDrawsByContestRange :many
   SELECT * FROM draws
   WHERE contest BETWEEN ? AND ?
   ORDER BY contest ASC;
   
   -- name: InsertDraw :exec
   INSERT INTO draws (
     contest, draw_date, bola1, bola2, bola3, bola4, bola5, source, raw_row
   ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
   
   -- name: UpsertDraw :exec
   INSERT INTO draws (
     contest, draw_date, bola1, bola2, bola3, bola4, bola5, source, raw_row
   ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
   ON CONFLICT(contest) DO UPDATE SET
     draw_date = excluded.draw_date,
     bola1 = excluded.bola1,
     bola2 = excluded.bola2,
     bola3 = excluded.bola3,
     bola4 = excluded.bola4,
     bola5 = excluded.bola5,
     source = excluded.source,
     raw_row = excluded.raw_row,
     imported_at = CURRENT_TIMESTAMP;
   
   -- name: CountDraws :one
   SELECT COUNT(*) FROM draws;
   
   -- name: GetContestRange :one
   SELECT MIN(contest) as min_contest, MAX(contest) as max_contest
   FROM draws;
   
   -- name: InsertImportHistory :one
   INSERT INTO import_history (
     filename, rows_inserted, rows_skipped, rows_errors, source_hash, metadata
   ) VALUES (?, ?, ?, ?, ?, ?)
   RETURNING *;
   
   -- name: GetImportHistory :many
   SELECT * FROM import_history
   ORDER BY imported_at DESC
   LIMIT ? OFFSET ?;
   ```

3. Test migration with `make migrate`
4. Run `make generate` to create Go code

**Testing:**
- Tables created successfully
- Constraints enforce data integrity
- Indexes exist
- sqlc generates Querier interface

---

#### Task 1.2.3: Simulations Database Schema
**Effort:** 4 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Create migrations for `simulations.db`.

**Acceptance Criteria:**
- [ ] `simulations` table created
- [ ] `simulation_contest_results` table created
- [ ] `analysis_jobs` table created
- [ ] Foreign keys and indexes configured

**Subtasks:**
1. Create `migrations/002_create_simulations.sql` (see design doc for full schema)
2. Include all three tables:
   - `simulations`
   - `simulation_contest_results`
   - `analysis_jobs`
3. Add indexes for status, contest, and simulation_id

**Testing:**
- All tables created
- Foreign keys work
- Cascade deletes tested

---

#### Task 1.2.4: Configs Database Schema
**Effort:** 2 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Create migrations for `configs.db`.

**Acceptance Criteria:**
- [ ] `configs` table created
- [ ] `config_presets` table created with default data
- [ ] Trigger for single default per mode

**Subtasks:**
1. Create `migrations/003_create_configs.sql` (see design doc)
2. Insert default presets (conservative, balanced, aggressive)
3. Create trigger for `is_default` enforcement

**Testing:**
- Presets inserted
- Trigger prevents multiple defaults per mode

---

#### Task 1.2.5: Finances Database Schema
**Effort:** 2 hours  
**Priority:** Critical  
**Assignee:** Dev 1

**Description:**
Create migrations for `finances.db`.

**Acceptance Criteria:**
- [ ] `ledger` table created
- [ ] `financial_summary` view created
- [ ] Indexes for queries

**Subtasks:**
1. Create `migrations/004_create_finances.sql`
2. Create ledger table
3. Create financial_summary view

**Testing:**
- View aggregates correctly
- Indexes improve query performance

---

#### Task 1.2.6: Database Access Layer with sqlc Integration
**Effort:** 6 hours  
**Priority:** High  
**Assignee:** Dev 1

**Description:**
Create database connection management and integrate sqlc-generated Queriers.

**Acceptance Criteria:**
- [ ] Connection pool for each database
- [ ] Querier instances created for each DB
- [ ] Transaction helpers with Querier support
- [ ] Proper connection closing

**Subtasks:**
1. Create `internal/store/db.go`:
   ```go
   package store
   
   import (
       "database/sql"
       "fmt"
       
       "github.com/garnizeh/luckyfive/internal/store/results"
       "github.com/garnizeh/luckyfive/internal/store/simulations"
       "github.com/garnizeh/luckyfive/internal/store/configs"
       "github.com/garnizeh/luckyfive/internal/store/finances"
       _ "modernc.org/sqlite"
   )
   
   type DB struct {
       ResultsDB     *sql.DB
       SimulationsDB *sql.DB
       ConfigsDB     *sql.DB
       FinancesDB    *sql.DB
       
       // Querier interfaces (mockable)
       Results     results.Querier
       Simulations simulations.Querier
       Configs     configs.Querier
       Finances    finances.Querier
   }
   
   type Config struct {
       ResultsPath     string
       SimulationsPath string
       ConfigsPath     string
       FinancesPath    string
   }
   
   func Open(cfg Config) (*DB, error) {
       db := &DB{}
       
       // Open Results DB
       resultsDB, err := sql.Open("sqlite", cfg.ResultsPath)
       if err != nil {
           return nil, fmt.Errorf("open results db: %w", err)
       }
       db.ResultsDB = resultsDB
       db.Results = results.New(resultsDB)
       
       // Open Simulations DB
       simulationsDB, err := sql.Open("sqlite", cfg.SimulationsPath)
       if err != nil {
           db.ResultsDB.Close()
           return nil, fmt.Errorf("open simulations db: %w", err)
       }
       db.SimulationsDB = simulationsDB
       db.Simulations = simulations.New(simulationsDB)
       
       // Open Configs DB
       configsDB, err := sql.Open("sqlite", cfg.ConfigsPath)
       if err != nil {
           db.ResultsDB.Close()
           db.SimulationsDB.Close()
           return nil, fmt.Errorf("open configs db: %w", err)
       }
       db.ConfigsDB = configsDB
       db.Configs = configs.New(configsDB)
       
       // Open Finances DB
       financesDB, err := sql.Open("sqlite", cfg.FinancesPath)
       if err != nil {
           db.ResultsDB.Close()
           db.SimulationsDB.Close()
           db.ConfigsDB.Close()
           return nil, fmt.Errorf("open finances db: %w", err)
       }
       db.FinancesDB = financesDB
       db.Finances = finances.New(financesDB)
       
       // Configure connection pools
       for _, sqlDB := range []*sql.DB{resultsDB, simulationsDB, configsDB, financesDB} {
           sqlDB.SetMaxOpenConns(25)
           sqlDB.SetMaxIdleConns(5)
       }
       
       return db, nil
   }
   
   func (db *DB) Close() error {
       var errs []error
       if err := db.ResultsDB.Close(); err != nil {
           errs = append(errs, err)
       }
       if err := db.SimulationsDB.Close(); err != nil {
           errs = append(errs, err)
       }
       if err := db.ConfigsDB.Close(); err != nil {
           errs = append(errs, err)
       }
       if err := db.FinancesDB.Close(); err != nil {
           errs = append(errs, err)
       }
       if len(errs) > 0 {
           return fmt.Errorf("close errors: %v", errs)
       }
       return nil
   }
   
   // WithTx executes a function within a transaction
   func (db *DB) WithTx(fn func(*sql.Tx) error) error {
       tx, err := db.ResultsDB.Begin()
       if err != nil {
           return err
       }
       defer tx.Rollback()
       
       if err := fn(tx); err != nil {
           return err
       }
       
       return tx.Commit()
   }
   ```

2. Create `internal/store/tx.go` for transaction helpers:
   ```go
   package store
   
   import (
       "context"
       "database/sql"
       
       "github.com/garnizeh/luckyfive/internal/store/results"
   )
   
   // WithResultsTx wraps a function in a results DB transaction
   func (db *DB) WithResultsTx(ctx context.Context, fn func(results.Querier) error) error {
       tx, err := db.ResultsDB.BeginTx(ctx, nil)
       if err != nil {
           return err
       }
       defer tx.Rollback()
       
       q := results.New(tx)
       if err := fn(q); err != nil {
           return err
       }
       
       return tx.Commit()
   }
   
   // Similar functions for other databases...
   ```

3. Create example test showing mock usage:
   ```go
   package services_test
   
   import (
       "testing"
       
       "github.com/stretchr/testify/assert"
       "go.uber.org/mock/gomock"
       
       "github.com/garnizeh/luckyfive/internal/store/results/mock"
   )
   
   func TestImportService_WithMock(t *testing.T) {
       ctrl := gomock.NewController(t)
       defer ctrl.Finish()
       
       mockQuerier := mock.NewMockQuerier(ctrl)
       
       // Set expectations
       mockQuerier.EXPECT().
           InsertDraw(gomock.Any(), gomock.Any()).
           Return(nil).
           Times(1)
       
       // Use mock in service
       // ... test code
       
       assert.NotNil(t, mockQuerier)
   }
   ```

**Testing:**
- All databases open successfully
- Querier interfaces work correctly
- Mock generation successful
- Transaction helpers tested
- Connection pooling works

---

### Sprint 1.3: Import Service (Days 8-10)

#### Task 1.3.1: XLSX Parser Implementation
**Effort:** 6 hours  
**Priority:** Critical  
**Assignee:** Dev 1 or Dev 2

**Description:**
Implement XLSX parsing with auto-detection of column layout.

**Acceptance Criteria:**
- [ ] Can parse XLSX files using excelize
- [ ] Auto-detects column names (flexible mapping)
- [ ] Handles different date formats
- [ ] Validates data integrity

**Subtasks:**
1. Create `internal/services/import.go`:
   ```go
   package services
   
   import (
       "context"
       "log/slog"
       
       "github.com/garnizeh/luckyfive/internal/store/results"
   )
   
   type ImportService struct {
       queries results.Querier  // Mockable interface
       logger  *slog.Logger
   }
   
   func NewImportService(queries results.Querier, logger *slog.Logger) *ImportService {
       return &ImportService{
           queries: queries,
           logger:  logger,
       }
   }
   
   func (s *ImportService) ParseXLSX(reader io.Reader, sheet string) ([]Draw, error)
   ```

2. Implement column detection logic:
   - Try common header names: "Concurso", "Contest", "Num"
   - Try common date formats: "DD/MM/YYYY", "YYYY-MM-DD"

3. Create `internal/models/draw.go`:
   ```go
   type Draw struct {
       Contest    int
       DrawDate   time.Time
       Bola1      int
       Bola2      int
       Bola3      int
       Bola4      int
       Bola5      int
       Source     string
       ImportedAt time.Time
       RawRow     string
   }
   
   func (d *Draw) Validate() error
   ```

4. Add validation logic

**Testing:**
- Test with sample XLSX files
- Test with different column layouts
- Test with invalid data
- Unit tests for validation
- **Mock test:** Use mock Querier to test service logic without DB

---

#### Task 1.3.2: Batch Insert Implementation
**Effort:** 4 hours  
**Priority:** High  
**Assignee:** Dev 1 or Dev 2

**Description:**
Implement batch insert for performance using sqlc-generated queries.

**Acceptance Criteria:**
- [ ] Batch inserts in transactions
- [ ] Configurable batch size
- [ ] Rollback on error
- [ ] Progress reporting

**Subtasks:**
1. Update `internal/services/import.go`:
   ```go
   func (s *ImportService) BatchInsertDraws(ctx context.Context, draws []models.Draw, batchSize int) (*ImportStats, error) {
       // Use transaction wrapper
       // Call s.queries.InsertDraw() or s.queries.UpsertDraw()
   }
   ```

2. Use sqlc-generated InsertDraw/UpsertDraw methods

3. Implement transaction-based batching using transaction Querier:
   ```go
   tx, err := db.ResultsDB.BeginTx(ctx, nil)
   if err != nil {
       return nil, err
   }
   defer tx.Rollback()
   
   txQueries := results.New(tx)
   
   for i, draw := range draws {
       err := txQueries.InsertDraw(ctx, results.InsertDrawParams{
           Contest: draw.Contest,
           // ... other fields
       })
       if err != nil {
           return nil, err
       }
       
       if (i+1)%batchSize == 0 {
           // Commit batch and start new tx if needed
       }
   }
   
   return tx.Commit()
   ```

4. Add progress callback for UI updates
5. Handle duplicate contest numbers using UpsertDraw

**Testing:**
- Test with 5000+ rows
- Verify transaction rollback on error
- Measure insert performance (<30s for 5000 rows)
- **Mock test:** Mock Querier to test batch logic without DB

---

#### Task 1.3.3: Import Service Complete
**Effort:** 5 hours  
**Priority:** High  
**Assignee:** Dev 1 or Dev 2

**Description:**
Complete ImportService with all methods from design doc.

**Acceptance Criteria:**
- [ ] SaveArtifact() implemented
- [ ] ImportArtifact() implemented
- [ ] GetDraw() implemented
- [ ] ListDraws() implemented
- [ ] ValidateDrawData() implemented

**Subtasks:**
1. Implement all interface methods
2. Add error handling
3. Add logging
4. Create import history records

**Testing:**
- Unit tests for each method
- Integration test: full import flow
- Test error scenarios

---

### Sprint 1.4: HTTP API Foundation (Days 11-14)

#### Task 1.4.1: HTTP Server Setup
**Effort:** 4 hours  
**Priority:** Critical  
**Assignee:** Dev 2

**Description:**
Set up HTTP server with routing and middleware.

**Acceptance Criteria:**
- [ ] Chi router configured
- [ ] Middleware stack (logging, recovery, CORS)
- [ ] Graceful shutdown
- [ ] Request ID tracking

**Subtasks:**
1. Create `cmd/api/main.go`:
   ```go
   func main() {
       cfg := config.Load()
       logger := logger.New(cfg.LogLevel)
       db := store.Open(cfg.Database)
       
       router := setupRouter(logger, db)
       
       server := &http.Server{
           Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
           Handler: router,
       }
       
       // Graceful shutdown logic
   }
   ```
2. Create `internal/middleware/logging.go`
3. Create `internal/middleware/recovery.go`
4. Create `internal/middleware/cors.go`

**Testing:**
- Server starts and stops gracefully
- Middleware executes in correct order
- CORS headers present

---

#### Task 1.4.2: Health Endpoint
**Effort:** 2 hours  
**Priority:** High  
**Assignee:** Dev 2

**Description:**
Implement `/api/v1/health` endpoint.

**Acceptance Criteria:**
- [ ] Returns 200 OK with system status
- [ ] Checks database connectivity
- [ ] Returns version information

**Subtasks:**
1. Create `internal/handlers/health.go`:
   ```go
   func HealthCheck(db *store.DB) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           status := checkDatabases(db)
           json.NewEncoder(w).Encode(status)
       }
   }
   ```
2. Add database ping checks
3. Return structured JSON response

**Testing:**
- Endpoint returns 200
- Database status accurate
- Response matches design doc format

---

#### Task 1.4.3: Results Upload Endpoint
**Effort:** 4 hours  
**Priority:** High  
**Assignee:** Dev 2

**Description:**
Implement `POST /api/v1/results/upload` for XLSX file upload.

**Acceptance Criteria:**
- [ ] Accepts multipart/form-data
- [ ] File size validation (max 50MB)
- [ ] File type validation (.xlsx, .xls)
- [ ] Returns artifact_id

**Subtasks:**
1. Create `internal/handlers/results.go`
2. Implement upload handler:
   ```go
   func UploadResults(importSvc *services.ImportService) http.HandlerFunc
   ```
3. Store uploaded file temporarily or in memory
4. Return artifact metadata

**Testing:**
- Upload valid XLSX file
- Reject invalid file types
- Reject oversized files
- Response matches design doc

---

#### Task 1.4.4: Results Import Endpoint
**Effort:** 4 hours  
**Priority:** High  
**Assignee:** Dev 2

**Description:**
Implement `POST /api/v1/results/import` to trigger import.

**Acceptance Criteria:**
- [ ] Accepts JSON request body
- [ ] Validates artifact_id exists
- [ ] Calls ImportService
- [ ] Returns import statistics

**Subtasks:**
1. Implement import handler
2. Validate request body
3. Call ImportService.ImportArtifact()
4. Return detailed results (inserted, skipped, errors)

**Testing:**
- Import succeeds with valid artifact
- Returns correct statistics
- Handles errors gracefully

---

#### Task 1.4.5: Results Query Endpoints
**Effort:** 5 hours  
**Priority:** Medium  
**Assignee:** Dev 2

**Description:**
Implement `GET /api/v1/results/{contest}` and `GET /api/v1/results`.

**Acceptance Criteria:**
- [ ] Single contest retrieval works
- [ ] List with pagination works
- [ ] Filtering by date range works
- [ ] Returns 404 for missing contests

**Subtasks:**
1. Implement GetDraw handler
2. Implement ListDraws handler with pagination
3. Add query parameter parsing
4. Add filtering logic

**Testing:**
- Retrieve existing contest
- 404 for non-existent contest
- Pagination works correctly
- Filters apply correctly

---

#### Task 1.4.6: Error Handling & Validation
**Effort:** 3 hours  
**Priority:** High  
**Assignee:** Dev 2

**Description:**
Implement consistent error handling and request validation.

**Acceptance Criteria:**
- [ ] Standard error response format
- [ ] Input validation with helpful messages
- [ ] HTTP status codes match design doc
- [ ] Request ID in all responses

**Subtasks:**
1. Create `internal/handlers/errors.go`:
   ```go
   func WriteError(w http.ResponseWriter, code int, err APIError)
   func WriteJSON(w http.ResponseWriter, code int, data interface{})
   ```
2. Create `internal/models/errors.go`
3. Add validation using validator library

**Testing:**
- Error responses match format
- Validation messages are clear
- Status codes correct

---

### Sprint 1.5: Testing & Documentation (Throughout Phase)

#### Task 1.5.1: Unit Tests - Database Layer with Mocks
**Effort:** 7 hours  
**Priority:** High  
**Assignee:** Dev 1

**Description:**
Write unit tests for database layer using mocks for 100% coverage.

**Acceptance Criteria:**
- [ ] Migration tests (up/down)
- [ ] sqlc query tests (integration)
- [ ] Mock-based unit tests for services
- [ ] Transaction tests
- [ ] Coverage > 90% for store package

**Subtasks:**
1. Create integration tests using in-memory SQLite:
   ```go
   // internal/store/results_integration_test.go
   func TestResultsQueries_Integration(t *testing.T) {
       db, err := sql.Open("sqlite", ":memory:")
       require.NoError(t, err)
       
       // Run migration
       _, err = db.Exec(migrationSQL)
       require.NoError(t, err)
       
       queries := results.New(db)
       
       // Test InsertDraw
       err = queries.InsertDraw(ctx, results.InsertDrawParams{...})
       assert.NoError(t, err)
       
       // Test GetDraw
       draw, err := queries.GetDraw(ctx, 1)
       assert.NoError(t, err)
       assert.Equal(t, 1, draw.Contest)
   }
   ```

2. Create mock-based unit tests:
   ```go
   // internal/services/import_test.go
   func TestImportService_ImportDraws_Success(t *testing.T) {
       ctrl := gomock.NewController(t)
       defer ctrl.Finish()
       
       mockQuerier := mock.NewMockQuerier(ctrl)
       logger := slog.New(slog.NewTextHandler(io.Discard, nil))
       svc := NewImportService(mockQuerier, logger)
       
       // Expect InsertDraw to be called
       mockQuerier.EXPECT().
           InsertDraw(gomock.Any(), gomock.Any()).
           Return(nil).
           Times(3)
       
       draws := []models.Draw{{Contest: 1}, {Contest: 2}, {Contest: 3}}
       stats, err := svc.BatchInsertDraws(context.Background(), draws, 100)
       
       assert.NoError(t, err)
       assert.Equal(t, 3, stats.Inserted)
   }
   
   func TestImportService_ImportDraws_Error(t *testing.T) {
       ctrl := gomock.NewController(t)
       defer ctrl.Finish()
       
       mockQuerier := mock.NewMockQuerier(ctrl)
       svc := NewImportService(mockQuerier, slog.Default())
       
       // Expect error
       mockQuerier.EXPECT().
           InsertDraw(gomock.Any(), gomock.Any()).
           Return(errors.New("db error"))
       
       draws := []models.Draw{{Contest: 1}}
       _, err := svc.BatchInsertDraws(context.Background(), draws, 100)
       
       assert.Error(t, err)
   }
   ```

3. Test migration up/down
4. Test transaction rollback scenarios
5. Test all edge cases and errors

**Testing:**
- `make test` passes
- Coverage report shows > 90%
- Both integration and mock tests pass
- All error paths covered

---

#### Task 1.5.2: Unit Tests - Import Service
**Effort:** 6 hours  
**Priority:** High  
**Assignee:** Dev 1 or Dev 2

**Description:**
Write unit tests for ImportService.

**Acceptance Criteria:**
- [ ] XLSX parsing tests
- [ ] Validation tests
- [ ] Batch insert tests
- [ ] Coverage > 75%

**Subtasks:**
1. Create test XLSX files (valid and invalid)
2. Test column detection with variations
3. Test data validation
4. Test error scenarios

**Testing:**
- All test cases pass
- Edge cases covered

---

#### Task 1.5.3: Integration Tests - Import Flow
**Effort:** 4 hours  
**Priority:** Medium  
**Assignee:** Dev 2

**Description:**
Write integration tests for complete import workflow.

**Acceptance Criteria:**
- [ ] End-to-end import test
- [ ] Tests use real XLSX file
- [ ] Verifies database state

**Subtasks:**
1. Create `tests/integration/import_test.go`
2. Test: upload → import → query flow
3. Verify database records

**Testing:**
- Integration test passes
- Database state correct after import

---

#### Task 1.5.4: API Tests - HTTP Endpoints
**Effort:** 5 hours  
**Priority:** High  
**Assignee:** Dev 2

**Description:**
Write HTTP endpoint tests.

**Acceptance Criteria:**
- [ ] Health endpoint tested
- [ ] Upload endpoint tested
- [ ] Import endpoint tested
- [ ] Query endpoints tested

**Subtasks:**
1. Create `internal/handlers/*_test.go`
2. Use `httptest` for testing
3. Test success and error cases
4. Verify response formats

**Testing:**
- All endpoint tests pass
- Response formats validated

---

#### Task 1.5.5: Documentation - API Endpoints
**Effort:** 3 hours  
**Priority:** Medium  
**Assignee:** Dev 2

**Description:**
Document implemented API endpoints.

**Acceptance Criteria:**
- [ ] OpenAPI/Swagger spec started
- [ ] Endpoint examples in README
- [ ] cURL examples provided

**Subtasks:**
1. Create `docs/api.md` or `openapi.yaml`
2. Document each endpoint with examples
3. Add to main README

**Testing:**
- Examples work as documented
- Spec validates (if using OpenAPI)

---

#### Task 1.5.6: Documentation - Setup Guide
**Effort:** 2 hours  
**Priority:** Medium  
**Assignee:** Dev 1

**Description:**
Update README with setup and usage instructions.

**Acceptance Criteria:**
- [ ] Prerequisites listed
- [ ] Build instructions
- [ ] Run instructions
- [ ] Example usage

**Subtasks:**
1. Update `README.md`
2. Add quickstart section
3. Add troubleshooting section

**Testing:**
- Fresh clone follows instructions successfully

---

### Phase 1 Checklist (current)

- [x] Task 1.1.1: Project initialized
- [x] Task 1.1.2: Makefile created
- [x] Task 1.1.3: Dependencies installed (including sqlc)
- [x] Task 1.1.4: Logging & configuration implemented (unit tests added)

**Progress:** The Makefile and dependency work are in place. `internal/config` and `internal/logger` were implemented and unit tests were added for both packages; those tests pass locally. The implementation files and tests are committed in the repository.

**Verification:**
- `go mod tidy` ran successfully
- `go test ./...` ran successfully; `internal/config` and `internal/logger` unit tests pass

Notes: some `build` targets (e.g., `bin/api`, `bin/worker`) may not produce binaries yet because corresponding `cmd/*` `main.go` files are placeholders or not fully implemented — those targets will be verified as the entrypoints are completed.

### Sprint 1.1 (Days 1-3)
- [x] Task 1.1.1: Project initialized
- [x] Task 1.1.2: Makefile created
- [x] Task 1.1.3: Dependencies installed (including sqlc) and documented in `README.md`
- [x] Task 1.1.4: Logging configured
    - [x] Task 1.1.5: sqlc configured with Querier interfaces (queriers generated and mocks created)

### Sprint 1.2 (Days 4-7)
- [ ] Task 1.2.1: Migration system working
- [ ] Task 1.2.2: Results schema created
- [ ] Task 1.2.3: Simulations schema created
- [ ] Task 1.2.4: Configs schema created
- [ ] Task 1.2.5: Finances schema created
- [ ] Task 1.2.6: DB access layer implemented

### Sprint 1.3 (Days 8-10)
- [ ] Task 1.3.1: XLSX parser working
- [ ] Task 1.3.2: Batch insert implemented
- [ ] Task 1.3.3: ImportService complete

### Sprint 1.4 (Days 11-14)
- [ ] Task 1.4.1: HTTP server running
- [ ] Task 1.4.2: Health endpoint working
- [ ] Task 1.4.3: Upload endpoint working
- [ ] Task 1.4.4: Import endpoint working
- [ ] Task 1.4.5: Query endpoints working
- [ ] Task 1.4.6: Error handling implemented

### Sprint 1.5 (Throughout)
- [ ] Task 1.5.1: DB layer tests passing
- [ ] Task 1.5.2: Import service tests passing
- [ ] Task 1.5.3: Integration tests passing
- [ ] Task 1.5.4: API tests passing
- [ ] Task 1.5.5: API documented
- [ ] Task 1.5.6: Setup guide complete

### Phase Gate
- [ ] All tasks completed
- [ ] Test coverage > 70%
- [ ] All tests passing
- [ ] Code reviewed
- [ ] Demo successful
- [ ] Stakeholder approval

---

## Metrics & KPIs

### Code Metrics
- **Lines of Code:** ~2500-3000 (including generated code)
- **Test Coverage:** > 90% (using mocks)
- **Number of Tests:** > 60
- **Packages Created:** ~12
- **Generated Files:** ~8 (sqlc + mocks)

### Performance Metrics
- **XLSX Import Time:** < 30s for 5000 rows
- **API Response Time:** < 100ms for health check
- **Database Init Time:** < 5s for all 4 databases

### Quality Metrics
- **Linting Errors:** 0
- **Code Review Comments:** < 20 per PR
- **Bugs Found in Testing:** < 10

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| XLSX format variations | Medium | Test with multiple file formats |
| Performance issues with large files | Medium | Batch inserts, streaming |
| Schema changes needed | Low | Migration system supports rollback |
| Time overrun | Medium | Prioritize critical tasks first |

---

## Dependencies

**External:**
- None (all within team control)

**Internal:**
- Go 1.21+ installed
- SQLite support in OS
- Git for version control

---

## Deliverables Summary

At the end of Phase 1, the following will be complete:

1. **Runnable API Server:**
   - Health check endpoint
   - Results upload endpoint
   - Results import endpoint
   - Results query endpoints

2. **Database Infrastructure:**
   - 4 SQLite databases with schemas
   - Migration system
   - Seed data (config presets)

3. **Import Functionality:**
   - XLSX parsing with auto-detection
   - Batch insert (5000+ rows in <30s)
   - Import history tracking

4. **Testing:**
   - 50+ unit tests
   - Integration tests
   - >70% code coverage

5. **Documentation:**
   - API documentation
   - Setup guide
   - Architecture docs

---

## Next Phase Preview

**Phase 2** will build on this foundation:
- Port prediction algorithms from `tools/loader.go`
- Implement simulation engine
- Build background worker system
- Add configuration management
- Create simple and advanced simulation endpoints

**Preparation for Phase 2:**
- Review existing algorithm code in `tools/loader.go`
- Identify reusable components
- Plan algorithm refactoring

---

**Questions or Issues:**
Contact the development team or create an issue in the project tracker.

