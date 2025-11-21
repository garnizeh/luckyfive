# Quina Simulation HTTP API — Design Document

## Answers received (incorporated below)

1. Preferred project layout: my proposal — filenames and layout that are easy to maintain/expand, flexible and simple.
2. Artifacts storage: store full artifacts (JSON/CSV) inside the databases as blobs (no external files).
3. Start with a single API instance (no cluster / load-balanced instances initially).
4. Authentication: not required now ("not now").
5. Retention / response size: ignore for now ("we will not worry about this for now").
6. Sweeps / optimization: you want the project to handle maximizing lottery gains via statistics; the API should include analysis and sweep facilities.
7. Financial ledger: you also need a balance ledger to record money added (from pocket or prize) and money spent (games played), with timestamps, counts and amounts.

> I incorporated all answers below. If you want any of these choices changed later I will update the document and remove the corresponding answered item.

---

## Executive summary

This document describes a HTTP API that exposes the functionality required to run, persist, and analyze Quina lottery simulations. The API design reuses the existing algorithms in `tools/loader/loader.go` and the Makefile flows, and organizes them into services, handlers and background workers. The system uses a distinct SQLite database file per dataset (official results, simulations, configs, finances) and stores full artifacts (JSON/CSV) as BLOBs inside the appropriate database(s) for simpler maintenance and atomicity.

Goals:
- Provide endpoints to import official results from XLSX
- Create and run simulations (synchronously or asynchronously)
- Persist simulation metadata and per-contest results
- Analyze many simulations to discover top parameter configurations
- Allow saving and applying named configurations
- Keep everything modular, following Go idioms and preferring stdlib

---

## What I inspected

- `tools/loader/loader.go` — import logic, schema detection and the full algorithmic core (generateAdvancedPredictions, runSimulation, etc.). Uses `excelize` and `modernc.org/sqlite`.
- `tools/simulations/compare_simulations.go` — reading simulation JSON artifacts and ranking by metrics.
- `tools/sweep/concurrent_sweep.go` and `tools/thinker/thinker.go` — parallel sweep and optimizer tooling invoking many runs.
- `tools/cards/compare_cards.go`, `tools/count/extract_cnt3.go` — helper tools for comparing cards and extracting metrics.
- `Makefile` — shows standard flows: `init-db`, `sim`, `sweep`, `gen-contest`, and defaults for simulation flags.

Takeaway: the algorithm is implemented and battle-tested in `loader.go`. The API should wrap, refactor and persist its outputs.

---

## High-level architecture

- cmd/api: HTTP server wiring
- internal/handlers: thin HTTP handlers (results, simulations, configs, analysis, jobs)
- internal/services: domain logic (Importer, SimulationService, EngineService, AnalysisService, ConfigService, JobWorker)
- internal/store: small helpers and migrations
- pkg/model: shared structs (GameConfig, SimulationResult, ContestResult)
- data/: persistent sqlite files and artifacts

Workers & jobs:
- A job worker polls the `simulations` table and executes pending simulation jobs. Jobs may be run synchronously if requested.

---

## Database files and schemas (one SQLite file per dataset)

Recommended file names under `data/` (configurable):
- `data/results.db` — official historical draws
- `data/simulations.db` — simulation jobs, metadata, per-contest normalized results and stored artifacts (BLOBs)
- `data/configs.db` — saved simulation parameter presets
- `data/finances.db` — ledger for money in/out (bets, prizes, manual adjustments)

Basic schema sketches (use JSON columns stored as TEXT):

### `results.db`

Table: `draws`

```sql
CREATE TABLE IF NOT EXISTS draws (
  contest INTEGER PRIMARY KEY,
  draw_date TEXT,
  bola1 INTEGER, bola2 INTEGER, bola3 INTEGER, bola4 INTEGER, bola5 INTEGER,
  source TEXT,
  imported_at TEXT,
  raw_row TEXT
);
CREATE INDEX IF NOT EXISTS idx_draws_contest ON draws(contest);
```

Notes: `raw_row` can store JSON/Text of original row when import layout varies.

### `simulations.db`

Table: `simulations`

```sql
CREATE TABLE IF NOT EXISTS simulations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at TEXT,
  started_at TEXT,
  finished_at TEXT,
  status TEXT,
  params TEXT,
  start_contest INTEGER,
  end_contest INTEGER,
  sim_preds INTEGER,
  summary TEXT,
  output_blob BLOB,     -- full JSON/CSV artifact stored in the DB
  output_name TEXT,     -- logical artifact filename
  log_blob BLOB,        -- optional log stored as blob
  error TEXT
);
CREATE INDEX IF NOT EXISTS idx_sim_status ON simulations(status);
```

Table: `simulation_contest_results`

```sql
CREATE TABLE IF NOT EXISTS simulation_contest_results (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  simulation_id INTEGER,
  contest INTEGER,
  actual TEXT,
  best_hits INTEGER,
  best_prediction_index INTEGER,
  predictions TEXT,
  FOREIGN KEY(simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_scr_simulation_id ON simulation_contest_results(simulation_id);
```

### `configs.db`

```sql
CREATE TABLE IF NOT EXISTS configs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT,
  params TEXT,
  is_default INTEGER DEFAULT 0,
  created_at TEXT,
  updated_at TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_configs_name ON configs(name);
```

Notes:
- JSON fields are stored as TEXT and parsed by the app. This keeps queries simple and lets the API return full structured objects.
- Artifacts (full JSON/CSV) are stored as BLOBs inside `simulations.db` (in `simulations.output_blob`) to keep artifacts atomic with metadata and simplify backups.

### `finances.db` (ledger)

```sql
CREATE TABLE IF NOT EXISTS ledger (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts TEXT,
  kind TEXT,           -- 'in' or 'out'
  source TEXT,         -- 'pocket', 'prize', 'ticket', 'other'
  amount REAL,
  currency TEXT DEFAULT 'BRL',
  description TEXT,
  simulation_id INTEGER NULL, -- link to a simulation if relevant
  cards_count INTEGER NULL,   -- how many cards purchased in this spend
  cost_per_card REAL NULL
);
CREATE INDEX IF NOT EXISTS idx_ledger_ts ON ledger(ts);
```

---

## HTTP API (resources & endpoints)

All endpoints use JSON. File uploads use multipart/form-data.

### Health
- GET `/api/v1/health` — 200 OK { status }

### Results (official draws)
- POST `/api/v1/results/upload` — upload XLSX (multipart `file`); returns `artifact_id` and path
- POST `/api/v1/results/import` — { artifact_id, sheet } — import into `results.db` returns inserted rows
- GET `/api/v1/results/{contest}` — get draw for contest
- GET `/api/v1/results` — list draws (limit, offset)

### Configs
- GET `/api/v1/configs` — list saved configs
- POST `/api/v1/configs` — create config {name, params, is_default}
- GET `/api/v1/configs/{id}`
- PUT `/api/v1/configs/{id}`
- POST `/api/v1/configs/{id}/apply` — set default

### Simulations
- POST `/api/v1/simulations` — create job; body example:

```json
{
  "start_contest": 6877,
  "end_contest": 6881,
  "sim_prev_max": 500,
  "sim_preds": 25,
  "params": {...GameConfig...},
  "save_output": true,
  "run_async": true
}
```

- Response: 202 { simulation_id, status }
- GET `/api/v1/simulations/{id}` — metadata + summary
- GET `/api/v1/simulations/{id}/results` — per-contest detailed results (paginated)
- POST `/api/v1/simulations/{id}/cancel` — request cancel
- GET `/api/v1/simulations` — list simulations; filter by status and date

### Analysis & Sweeps
- POST `/api/v1/analysis/run` — start a sweep or portfolio optimization (specify ranges and runs)
- GET `/api/v1/analysis/{job_id}` — status + summary
- GET `/api/v1/analysis/best-configs?sort=quina&top=5` — top N configs across completed sims

### Artifacts
- GET `/api/v1/artifacts/{id}` — return artifact file

---

## Services & responsibilities

1. ImportService — upload and import XLSX to `results.db` (uses `excelize`, batch inserts, schema detection)
2. EngineService — port of algorithms from `loader.go` (predictor functions, scoring, evolutionary steps) but pure functions returning `SimulationResult`
3. SimulationService — create simulation job, persist job metadata, orchestrate runs (sync or enqueue for worker)
4. JobWorker — background worker(s) that claim pending jobs from `simulations` table and run them; supports cancellation and updates job status
5. AnalysisService — aggregate simulation summaries and run higher-level sweeps (randomized parameter scheduling)
6. ConfigService — CRUD for named parameter presets and default config

Notes on job reservation:
- Use a small transaction to atomically pick a pending job and mark it reserved/running. The worker should update `started_at` and `status` and write heartbeats if desired.

---

## Data flows (examples)

1) XLSX import
- Upload file → artifact row
- Import request → ImportService reads XLSX → inserts into `results.db` draws table → returns count

2) Create simulation job
- POST /simulations → SimulationService inserts `simulations` row with `status=pending`
- Worker polls, reserves job, loads `results.db` draws, calls EngineService.runSimulationCore(...) → returns SimulationResult
- Persist per-contest rows in `simulation_contest_results`, write JSON artifact to `data/simulations/`, update `simulations.summary`, `status=completed`.

3) Analysis
- Query `simulations` table summary JSONs (or read artifacts) → aggregate and present top configs (like `compare_simulations.go`)

---

## Example request/response (create simulation)

Request: POST `/api/v1/simulations`

Response 202 Accepted:

```json
{
  "simulation_id": 101,
  "status": "pending",
  "created_at": "2025-11-20T12:34:56Z"
}
```

Later GET `/api/v1/simulations/101` returns summary and `output_file` path.

---

## Integration with existing Makefile and CLI tools

- Keep `Makefile` targets but add `make api` that runs the HTTP server for local testing.
- Reuse algorithms from `tools/loader/loader.go` by moving algorithmic functions into `internal/services/engine.go` and import them both from the API and from the CLI tools (or keep CLI wrappers that call the new packages).
- `concurrent_sweep.go` and `thinker.go` behaviors can either be replaced by analysis endpoints that schedule many simulation jobs, or kept as CLI tools that call the API endpoints.

---

## Implementation notes & choices (recommendations)

- Routing: `net/http` + small router (stdlib or `chi`) — stdlib suffices for a minimal version.
- DB driver: `modernc.org/sqlite` for pure-Go builds (already used in tools).
- XLSX: `github.com/xuri/excelize/v2` (already used).
- Keep engine functions pure and deterministic by injecting `rand.Rand` seeds.
- Persist full simulation JSON outputs as files (easier to inspect and avoids SQLite size bloat). Store file path in `simulations.output_file`.
- For multi-instance deployments, recommend moving job queue to a dedicated broker (Redis) or use per-worker DB reservation with careful locking.

---

## Minimal implementation plan (next steps)

1. Create project skeleton and migration scripts for `results.db`, `simulations.db`, `configs.db`.
2. Implement ImportService and one endpoint to upload/import XLSX.
3. Port engine core functions to `internal/services/engine.go` with tests (seeded RNG).
4. Implement SimulationService and JobWorker; implement POST `/simulations` and worker loop that runs a simple job and persists results.
5. Implement Analysis endpoints mirroring `compare_simulations.go`.

If you'd like, I can start implementing step 1 and 2 now and push the files here for review.

---

## Next actions for you

All questions have been answered and incorporated into this document. If you want to change any of the choices (artifact storage, DB filenames, starting with a single API instance, etc.), tell me which item to update and I will revise the doc and carry the change into the implementation plan.

If everything looks good I will proceed to implement the project skeleton and the importer (migrations + ImportService) next.

---

## Change log

- 2025-11-20: Initial design doc created based on repository `tools` and `Makefile` contents.

---

## Expanded details — Part 1: Architecture diagrams and detailed handlers/services (results & configs)

This is the first incremental expansion. I'll add more small focused sections after you review this part.

### Table of contents (expanded)

1. Questions / Answers
2. Executive summary
3. What I inspected
4. High-level architecture
5. Database files and schemas
6. HTTP API (resources & endpoints)
7. Services & responsibilities
8. Data flows
9. Expanded details (this section)
   - Architecture diagrams (Mermaid)
   - Handlers: detailed request/response + validation
   - Services: responsibilities and key methods
10. Implementation notes
11. Minimal implementation plan
12. Next actions

### Architecture diagram (High level)

```mermaid
flowchart LR
  subgraph API[API Server]
    A[HTTP Handlers]
  end
  subgraph SVC[Services]
    B1[ImportService]
    B2[SimulationService]
    B3[EngineService]
    B4[AnalysisService]
    B5[ConfigService]
    B6[JobWorker]
    B7[FinanceService]
  end
  subgraph DBs[Databases]
    D1[(results.db)]
    D2[(simulations.db)]
    D3[(configs.db)]
    D4[(finances.db)]
  end

  A --> B1
  A --> B2
  A --> B4
  A --> B5
  A --> B7
  B1 --> D1
  B2 --> B3
  B2 --> D2
  B3 --> D1
  B4 --> D2
  B5 --> D3
  B7 --> D4
  B6 --> B2
  B6 --> D2
```

Explanation: handlers receive HTTP requests and call services. `EngineService` implements the core algorithms and reads official results from `results.db`. `SimulationService` orchestrates runs and writes artifacts into `simulations.db`. `JobWorker` polls the `simulations` table and executes jobs.

### Data flow diagram (Import + Simulation)

```mermaid
sequenceDiagram
  participant Client
  participant API
  participant ImportSvc
  participant ResultsDB
  participant SimulationSvc
  participant EngineSvc
  participant SimDB

  Client->>API: POST /results/upload (file)
  API->>ImportSvc: store artifact, validate
  ImportSvc->>ResultsDB: INSERT draws
  ResultsDB-->>ImportSvc: OK
  ImportSvc-->>API: 201 {artifact_id}

  Client->>API: POST /simulations {params}
  API->>SimulationSvc: create job (status=pending)
  API-->>Client: 202 {simulation_id}
  JobWorker->>SimulationSvc: claim job
  SimulationSvc->>EngineSvc: runSimulation(start,end,params)
  EngineSvc->>ResultsDB: SELECT previous draws
  EngineSvc-->>SimulationSvc: SimulationResult
  SimulationSvc->>SimDB: INSERT contest results, artifact blob, summary
  SimDB-->>SimulationSvc: OK
  SimulationSvc-->>JobWorker: finish
```

### Handlers — detailed (Part A): Results endpoints

These handlers are thin: validate input, call ImportService or DB read functions, return JSON and proper HTTP codes.

- POST /api/v1/results/upload
  - Request: multipart/form-data with field `file` (xlsx). Optional `source` string.
  - Validation: ensure file present and extension `.xlsx` or `.xls`.
  - Service call: ImportService.SaveArtifact(ctx, filename, reader) -> artifactID
  - Response 201 Created: { artifact_id, filename, created_at }

- POST /api/v1/results/import
  - Request body JSON: { "artifact_id": int, "sheet": string (optional), "replace": bool }
  - Validation: artifact exists; sheet optional
  - Service call: ImportService.ImportArtifact(ctx, artifactID, sheet, replace) -> {inserted, skipped}
  - Response 200: { inserted: N, skipped: M }

- GET /api/v1/results/{contest}
  - Path param: contest integer
  - Validation: contest > 0
  - Service call: ImportService.GetDraw(ctx, contest) -> Draw struct
  - Response 200: { contest, date, balls: [b1...b5] } or 404

Implementation notes for ImportService (see Services below).

### Handlers — detailed (Part B): Config endpoints

- GET /api/v1/configs
  - Params: limit, offset
  - Service: ConfigService.List(ctx, limit, offset)
  - Response: 200 { configs: [{id,name,params,is_default,...}] }

- POST /api/v1/configs
  - Body: { name: string, params: GameConfig, is_default: bool }
  - Validation: name non-empty, params validated ranges (alpha>0 etc.)
  - Service: ConfigService.Create(ctx, name, params, isDefault)
  - Response 201 { id, name }

- POST /api/v1/configs/{id}/apply
  - Service: ConfigService.SetDefault(ctx, id)
  - Response 200 { id, is_default:true }

### Errors & status codes (common patterns)

- 200 OK — success with payload
- 201 Created — resource created
- 202 Accepted — job accepted (async)
- 400 Bad Request — validation errors (return JSON with `errors` array)
- 404 Not Found — resource not found
- 500 Internal Server Error — unexpected errors

### Services — detailed (Part A): ImportService

Responsibilities:
- Save uploaded artifacts (artifact metadata and blob stored in `simulations.db` or dedicated artifacts table depending on type)
- Parse XLSX and map columns to `draws` table
- Batch insert rows and return a summary

Key methods (Go signatures, conceptual):

```go
type ImportService interface {
  SaveArtifact(ctx context.Context, name string, rd io.Reader) (artifactID int64, err error)
  ImportArtifact(ctx context.Context, artifactID int64, sheet string, replace bool) (inserted int, skipped int, err error)
  GetDraw(ctx context.Context, contest int) (*Draw, error)
}
```

Implementation notes:
- Use `excelize` for reading xlsx. Use the loader's `loadSpreadsheet`, `buildTable`, `insertRows` logic as reference.
- Use `sqlc` generated queries to insert/select draws (see later section for sqlc setup).
- Batch insert in transactions (e.g., 500 rows per tx).

### Services — detailed (Part B): ConfigService

Responsibilities:
- CRUD for named parameter presets stored in `configs.db`
- Enforce a single `is_default` config (transactionally clear existing default and set new one)

Key methods:

```go
type ConfigService interface {
  List(ctx context.Context, limit, offset int) ([]Config, error)
  Create(ctx context.Context, cfg Config) (int64, error)
  Get(ctx context.Context, id int64) (Config, error)
  Update(ctx context.Context, id int64, cfg Config) error
  SetDefault(ctx context.Context, id int64) error
}
```

Implementation notes:
- Use `sqlc` to generate typed DB access for `configs.db` (simple queries). Prefer parametrized queries for safety.

---

End of Part 1. After you review, I will expand Part 2 which will include:

- Detailed Service APIs for SimulationService, EngineService, JobWorker and AnalysisService
- SQL schema files prepared for `sqlc` (one file per DB) and example `.sql` queries
- Mermaid diagrams for Job lifecycle and Analysis flows
- Error handling patterns, logging, and metrics suggestions

I updated the todo list to mark this step in progress and will move it to completed after your approval of Part 1.
