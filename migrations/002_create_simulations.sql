-- Minimal schema for simulations to satisfy sqlc generation
CREATE TABLE IF NOT EXISTS simulations (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL
);
