-- Migration: 005_create_sweeps.sql
-- Creates sweeps table in configs.db for storing sweep configurations

-- Up migration

CREATE TABLE IF NOT EXISTS sweeps (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  description TEXT,

  -- Sweep configuration JSON (full sweep.SweepConfig)
  config_json TEXT NOT NULL,

  -- Metadata
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_by TEXT,

  -- Usage statistics
  times_used INTEGER DEFAULT 0,
  last_used_at TEXT
);

CREATE INDEX IF NOT EXISTS ux_sweeps_name ON sweeps(name);
CREATE INDEX IF NOT EXISTS idx_sweeps_created_at ON sweeps(created_at);
CREATE INDEX IF NOT EXISTS idx_sweeps_usage ON sweeps(times_used DESC, last_used_at DESC);
CREATE INDEX IF NOT EXISTS idx_sweeps_updated ON sweeps(updated_at DESC);

-- Down migration
-- DROP INDEX IF EXISTS idx_sweeps_updated;
-- DROP INDEX IF EXISTS idx_sweeps_usage;
-- DROP INDEX IF EXISTS idx_sweeps_created_at;
-- DROP INDEX IF EXISTS ux_sweeps_name;
-- DROP TABLE IF EXISTS sweeps;