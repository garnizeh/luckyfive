-- Minimal schema for finances to satisfy sqlc generation
CREATE TABLE IF NOT EXISTS finances (
  id INTEGER PRIMARY KEY,
  account TEXT NOT NULL,
  balance INTEGER NOT NULL DEFAULT 0
);
