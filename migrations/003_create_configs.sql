-- Minimal schema for configs to satisfy sqlc generation
CREATE TABLE IF NOT EXISTS configs (
  id INTEGER PRIMARY KEY,
  key TEXT NOT NULL,
  value TEXT
);
