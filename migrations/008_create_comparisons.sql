-- Migration: 008_create_comparisons.sql
-- Creates tables for storing simulation comparison results

-- Up migration

-- Table: comparisons
CREATE TABLE IF NOT EXISTS comparisons (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,
  simulation_ids TEXT NOT NULL,  -- JSON array of simulation IDs
  metric TEXT NOT NULL,          -- Primary metric for comparison ('quina_rate', 'avg_hits', 'roi', etc.)
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  result_json TEXT                -- Cached comparison results
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