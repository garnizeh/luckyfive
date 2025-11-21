-- Migration: 002_create_simulations.sql
-- Creates tables for simulations.db: simulations, simulation_contest_results, analysis_jobs

-- Up migration

-- Table: simulations
CREATE TABLE IF NOT EXISTS simulations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  started_at TEXT,
  finished_at TEXT,
  status TEXT NOT NULL DEFAULT 'pending'
    CHECK(status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),

  -- Configuration
  recipe_name TEXT,
  recipe_json TEXT NOT NULL,  -- Full JSON recipe for reproducibility
  mode TEXT NOT NULL CHECK(mode IN ('simple', 'advanced')),

  -- Contest range
  start_contest INTEGER NOT NULL,
  end_contest INTEGER NOT NULL,

  -- Execution metadata
  worker_id TEXT,
  run_duration_ms INTEGER,

  -- Results summary (JSON)
  summary_json TEXT,

  -- Artifacts stored as BLOBs
  output_blob BLOB,
  output_name TEXT,
  log_blob BLOB,

  -- Error tracking
  error_message TEXT,
  error_stack TEXT,

  -- Ownership
  created_by TEXT,

  CHECK(end_contest >= start_contest)
);

CREATE INDEX IF NOT EXISTS idx_sim_status ON simulations(status);
CREATE INDEX IF NOT EXISTS idx_sim_created ON simulations(created_at);
CREATE INDEX IF NOT EXISTS idx_sim_recipe_name ON simulations(recipe_name);
CREATE INDEX IF NOT EXISTS idx_sim_mode ON simulations(mode);

-- Table: simulation_contest_results
CREATE TABLE IF NOT EXISTS simulation_contest_results (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  simulation_id INTEGER NOT NULL,
  contest INTEGER NOT NULL,

  -- Actual drawn numbers
  actual_numbers TEXT NOT NULL,  -- JSON array: [1,5,12,34,56]

  -- Best prediction performance
  best_hits INTEGER NOT NULL,
  best_prediction_index INTEGER,
  best_prediction_numbers TEXT,  -- JSON array

  -- All predictions for this contest
  predictions_json TEXT NOT NULL,  -- Full prediction set

  -- Metadata
  processed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY(simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scr_simulation_id ON simulation_contest_results(simulation_id);
CREATE INDEX IF NOT EXISTS idx_scr_contest ON simulation_contest_results(contest);
CREATE INDEX IF NOT EXISTS idx_scr_hits ON simulation_contest_results(best_hits);

-- Table: analysis_jobs
CREATE TABLE IF NOT EXISTS analysis_jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  started_at TEXT,
  finished_at TEXT,
  status TEXT NOT NULL DEFAULT 'pending',

  job_type TEXT NOT NULL CHECK(job_type IN ('sweep', 'optimization', 'comparison')),

  -- Configuration
  config_json TEXT NOT NULL,  -- Parameter ranges, etc.

  -- Results
  total_simulations INTEGER,
  completed_simulations INTEGER,
  failed_simulations INTEGER,

  -- Best configs found
  top_configs_json TEXT,

  -- Artifacts
  report_blob BLOB,
  report_name TEXT,

  error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_analysis_status ON analysis_jobs(status);
CREATE INDEX IF NOT EXISTS idx_analysis_type ON analysis_jobs(job_type);

-- Down migration
-- DROP INDEX IF EXISTS idx_analysis_type;
-- DROP INDEX IF EXISTS idx_analysis_status;
-- DROP INDEX IF EXISTS idx_scr_hits;
-- DROP INDEX IF EXISTS idx_scr_contest;
-- DROP INDEX IF EXISTS idx_scr_simulation_id;
-- DROP INDEX IF EXISTS idx_sim_mode;
-- DROP INDEX IF EXISTS idx_sim_recipe_name;
-- DROP INDEX IF EXISTS idx_sim_created;
-- DROP INDEX IF EXISTS idx_sim_status;
-- DROP TABLE IF EXISTS analysis_jobs;
-- DROP TABLE IF EXISTS simulation_contest_results;
-- DROP TABLE IF EXISTS simulations;
