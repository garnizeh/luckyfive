-- Production-ready schema for configs DB
CREATE TABLE IF NOT EXISTS configs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key TEXT NOT NULL UNIQUE,
  value TEXT,
  description TEXT,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS config_presets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  mode TEXT NOT NULL,
  settings TEXT NOT NULL, -- JSON blob
  is_default INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Trigger to ensure only one default preset per mode
CREATE TRIGGER IF NOT EXISTS trg_config_presets_single_default
BEFORE INSERT ON config_presets
WHEN NEW.is_default = 1
BEGIN
  UPDATE config_presets SET is_default = 0 WHERE mode = NEW.mode;
END;

-- Down (commented):
-- DROP TRIGGER IF EXISTS trg_config_presets_single_default;
-- DROP TABLE IF EXISTS config_presets;
-- DROP TABLE IF EXISTS configs;
