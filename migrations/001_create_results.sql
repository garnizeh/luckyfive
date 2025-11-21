-- Minimal schema for sqlc: create draws and import_history tables
CREATE TABLE IF NOT EXISTS draws (
  contest INTEGER PRIMARY KEY,
  draw_date TEXT NOT NULL,
  bola1 INTEGER,
  bola2 INTEGER,
  bola3 INTEGER,
  bola4 INTEGER,
  bola5 INTEGER,
  source TEXT,
  raw_row TEXT
);

CREATE TABLE IF NOT EXISTS import_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  filename TEXT NOT NULL,
  imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  rows_inserted INTEGER,
  rows_skipped INTEGER,
  rows_errors INTEGER,
  source_hash TEXT,
  metadata TEXT
);
