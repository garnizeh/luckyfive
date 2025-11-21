-- Production-ready schema for simulations DB
CREATE TABLE IF NOT EXISTS simulations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  settings TEXT, -- JSON blob with simulation parameters
  status TEXT NOT NULL DEFAULT 'idle'
);

CREATE TABLE IF NOT EXISTS simulation_contest_results (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  simulation_id INTEGER NOT NULL,
  contest INTEGER NOT NULL,
  result_blob TEXT, -- serialized result details (JSON)
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS analysis_jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  simulation_id INTEGER NOT NULL,
  started_at TEXT,
  finished_at TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  worker TEXT,
  result_summary TEXT,
  FOREIGN KEY(simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_simulations_status ON simulations(status);
CREATE INDEX IF NOT EXISTS idx_simulation_results_simulation_id ON simulation_contest_results(simulation_id);

-- Down (commented):
-- DROP INDEX IF EXISTS idx_simulations_status;
-- DROP INDEX IF EXISTS idx_simulation_results_simulation_id;
-- DROP TABLE IF EXISTS analysis_jobs;
-- DROP TABLE IF EXISTS simulation_contest_results;
-- DROP TABLE IF EXISTS simulations;
