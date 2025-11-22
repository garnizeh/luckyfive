-- Migration: 007_create_sweep_execution.sql
-- Creates tables for sweep execution tracking in simulations.db

-- Up migration

-- Table: sweep_jobs
CREATE TABLE IF NOT EXISTS sweep_jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,

  -- Sweep configuration
  sweep_config_json TEXT NOT NULL,

  -- Contest range for all simulations
  base_contest_range TEXT NOT NULL,

  -- Status tracking
  status TEXT NOT NULL DEFAULT 'pending'
    CHECK(status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
  total_combinations INTEGER NOT NULL,
  completed_simulations INTEGER DEFAULT 0,
  failed_simulations INTEGER DEFAULT 0,

  -- Timing
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  started_at TEXT,
  finished_at TEXT,
  run_duration_ms INTEGER,

  -- Ownership
  created_by TEXT
);

CREATE INDEX IF NOT EXISTS idx_sweep_jobs_status ON sweep_jobs(status);
CREATE INDEX IF NOT EXISTS idx_sweep_jobs_created ON sweep_jobs(created_at DESC);

-- Table: sweep_simulations
CREATE TABLE IF NOT EXISTS sweep_simulations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  sweep_job_id INTEGER NOT NULL,
  simulation_id INTEGER NOT NULL,
  variation_index INTEGER NOT NULL,
  variation_params TEXT NOT NULL,  -- JSON of parameter values

  FOREIGN KEY (sweep_job_id) REFERENCES sweep_jobs(id) ON DELETE CASCADE,
  FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sweep_simulations_sweep_job_id ON sweep_simulations(sweep_job_id);
CREATE INDEX IF NOT EXISTS idx_sweep_simulations_simulation_id ON sweep_simulations(simulation_id);

-- Table: comparisons
CREATE TABLE IF NOT EXISTS comparisons (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,
  simulation_ids TEXT NOT NULL,  -- JSON array of simulation IDs
  metric TEXT NOT NULL,          -- Primary metric for comparison
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  result_json TEXT               -- Cached comparison results
);

CREATE INDEX IF NOT EXISTS idx_comparisons_created ON comparisons(created_at DESC);

-- Table: comparison_metrics
CREATE TABLE IF NOT EXISTS comparison_metrics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  comparison_id INTEGER NOT NULL,
  simulation_id INTEGER NOT NULL,
  metric_name TEXT NOT NULL,
  metric_value REAL NOT NULL,
  rank INTEGER,
  percentile REAL,

  FOREIGN KEY (comparison_id) REFERENCES comparisons(id) ON DELETE CASCADE,
  FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_comparison_metrics_comparison_id ON comparison_metrics(comparison_id);
CREATE INDEX IF NOT EXISTS idx_comparison_metrics_rank ON comparison_metrics(comparison_id, rank);

-- Down migration
-- DROP INDEX IF EXISTS idx_comparison_metrics_rank;
-- DROP INDEX IF EXISTS idx_comparison_metrics_comparison_id;
-- DROP TABLE IF EXISTS comparison_metrics;
-- DROP INDEX IF EXISTS idx_comparisons_created;
-- DROP TABLE IF EXISTS comparisons;
-- DROP INDEX IF EXISTS idx_sweep_simulations_simulation_id;
-- DROP INDEX IF EXISTS idx_sweep_simulations_sweep_job_id;
-- DROP TABLE IF EXISTS sweep_simulations;
-- DROP INDEX IF EXISTS idx_sweep_jobs_created;
-- DROP INDEX IF EXISTS idx_sweep_jobs_status;
-- DROP TABLE IF EXISTS sweep_jobs;