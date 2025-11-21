-- Migration: 003_create_configs.sql
-- Creates tables for configs.db: configs, config_presets

-- Up migration

-- Production-ready schema for configs DB
-- Stores simulation configurations and presets for simple/advanced modes
CREATE TABLE IF NOT EXISTS configs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  description TEXT,

  -- Recipe JSON (full simulation configuration)
  recipe_json TEXT NOT NULL,

  -- Categorization
  tags TEXT,  -- JSON array: ["conservative", "proven"]

  -- Defaults
  is_default INTEGER DEFAULT 0 CHECK(is_default IN (0, 1)),
  mode TEXT NOT NULL CHECK(mode IN ('simple', 'advanced')),

  -- Metadata
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_by TEXT,

  -- Usage statistics
  times_used INTEGER DEFAULT 0,
  last_used_at TEXT
);

CREATE INDEX IF NOT EXISTS ux_configs_name ON configs(name);
CREATE INDEX IF NOT EXISTS idx_configs_default ON configs(is_default);
CREATE INDEX IF NOT EXISTS idx_configs_mode ON configs(mode);

-- Trigger to ensure only one default per mode
CREATE TRIGGER IF NOT EXISTS trg_one_default_per_mode
BEFORE UPDATE OF is_default ON configs
WHEN NEW.is_default = 1
BEGIN
  UPDATE configs SET is_default = 0
  WHERE mode = NEW.mode AND id != NEW.id;
END;

CREATE TABLE IF NOT EXISTS config_presets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  description TEXT,
  recipe_json TEXT NOT NULL,
  risk_level TEXT CHECK(risk_level IN ('low', 'medium', 'high')),
  is_active INTEGER DEFAULT 1,
  sort_order INTEGER DEFAULT 0
);

-- Pre-populated with default presets
INSERT INTO config_presets (name, display_name, description, recipe_json, risk_level, sort_order)
VALUES
  ('conservative', 'Conservative', 'Low risk, proven parameters',
   '{"alpha":0.3,"beta":0.25,"gamma":0.25,"delta":0.2,"sim_prev_max":500,"sim_preds":20}',
   'low', 1),
  ('balanced', 'Balanced', 'Medium risk, good balance',
   '{"alpha":0.35,"beta":0.25,"gamma":0.2,"delta":0.2,"sim_prev_max":400,"sim_preds":25}',
   'medium', 2),
  ('aggressive', 'Aggressive', 'Higher risk, more predictions',
   '{"alpha":0.4,"beta":0.3,"gamma":0.15,"delta":0.15,"sim_prev_max":300,"sim_preds":30}',
   'high', 3);

-- Down migration
-- DROP TRIGGER IF EXISTS trg_one_default_per_mode;
-- DROP INDEX IF EXISTS idx_configs_mode;
-- DROP INDEX IF EXISTS idx_configs_default;
-- DROP INDEX IF EXISTS ux_configs_name;
-- DROP TABLE IF EXISTS config_presets;
-- DROP TABLE IF EXISTS configs;
