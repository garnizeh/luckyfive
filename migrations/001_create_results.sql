-- Migration: 001_create_results.sql
-- Creates tables for results.db: draws, import_history

-- Up migration

-- Production-ready schema for results DB
-- Creates draws table with data integrity checks and import history tracking
CREATE TABLE IF NOT EXISTS draws (
  contest INTEGER PRIMARY KEY,
  draw_date TEXT NOT NULL,
  bola1 INTEGER NOT NULL CHECK(bola1 BETWEEN 1 AND 80),
  bola2 INTEGER NOT NULL CHECK(bola2 BETWEEN 1 AND 80),
  bola3 INTEGER NOT NULL CHECK(bola3 BETWEEN 1 AND 80),
  bola4 INTEGER NOT NULL CHECK(bola4 BETWEEN 1 AND 80),
  bola5 INTEGER NOT NULL CHECK(bola5 BETWEEN 1 AND 80),
  source TEXT,
  imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  raw_row TEXT,
  -- ensure ascending order of balls when inserted (application should also enforce)
  CHECK(bola1 < bola2 AND bola2 < bola3 AND bola3 < bola4 AND bola4 < bola5)
);

CREATE INDEX IF NOT EXISTS idx_draws_draw_date ON draws(draw_date);
CREATE INDEX IF NOT EXISTS idx_draws_imported_at ON draws(imported_at);

CREATE TABLE IF NOT EXISTS import_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  filename TEXT NOT NULL,
  imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  rows_inserted INTEGER NOT NULL DEFAULT 0,
  rows_skipped INTEGER NOT NULL DEFAULT 0,
  rows_errors INTEGER NOT NULL DEFAULT 0,
  source_hash TEXT,
  metadata TEXT
);

-- Down migration
-- DROP INDEX IF EXISTS idx_draws_imported_at;
-- DROP INDEX IF EXISTS idx_draws_draw_date;
-- DROP TABLE IF EXISTS import_history;
-- DROP TABLE IF EXISTS draws;
