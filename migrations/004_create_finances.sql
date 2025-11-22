-- Migration: 004_create_finances.sql
-- Adds comprehensive financial tracking tables to existing finances.db schema

-- Up migration

-- Prize table (Quina official prize structure)
CREATE TABLE IF NOT EXISTS prize_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    contest INTEGER NOT NULL,
    prize_type TEXT NOT NULL,  -- 'quina', 'quadra', 'terno'
    amount_cents INTEGER NOT NULL,
    winners INTEGER,
    total_collected_cents INTEGER,
    notes TEXT,
    UNIQUE(contest, prize_type)
);

-- Bet costs (per region/year)
CREATE TABLE IF NOT EXISTS bet_costs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    effective_from TEXT NOT NULL,
    effective_to TEXT,
    cost_cents INTEGER NOT NULL,
    numbers_count INTEGER DEFAULT 5,
    region TEXT DEFAULT 'BR',
    notes TEXT
);

-- Extended ledger entries with more financial tracking (keeping original ledger table)
CREATE TABLE IF NOT EXISTS ledger_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    transaction_date TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    transaction_type TEXT NOT NULL,  -- 'bet', 'prize', 'budget', 'adjustment'
    amount_cents INTEGER NOT NULL,  -- negative for costs, positive for prizes
    simulation_id INTEGER,
    contest INTEGER,
    description TEXT,
    metadata_json TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE SET NULL
);

-- Simulation financial summary
CREATE TABLE IF NOT EXISTS simulation_finances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    simulation_id INTEGER NOT NULL UNIQUE,
    total_bets INTEGER NOT NULL DEFAULT 0,
    total_cost_cents INTEGER NOT NULL DEFAULT 0,
    total_prizes_cents INTEGER NOT NULL DEFAULT 0,
    net_profit_cents INTEGER NOT NULL DEFAULT 0,
    roi_percentage REAL NOT NULL DEFAULT 0.0,
    break_even_contest INTEGER,
    best_prize_cents INTEGER DEFAULT 0,
    best_prize_contest INTEGER,
    quina_wins INTEGER DEFAULT 0,
    quadra_wins INTEGER DEFAULT 0,
    terno_wins INTEGER DEFAULT 0,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
);

-- Contest bets (what was actually bet)
CREATE TABLE IF NOT EXISTS contest_bets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    simulation_id INTEGER NOT NULL,
    contest INTEGER NOT NULL,
    bet_numbers TEXT NOT NULL,  -- JSON array
    cost_cents INTEGER NOT NULL,
    hits INTEGER,
    prize_type TEXT,  -- 'quina', 'quadra', 'terno', null
    prize_cents INTEGER DEFAULT 0,
    placed_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE,
    UNIQUE(simulation_id, contest, bet_numbers)
);

-- Budget management
CREATE TABLE IF NOT EXISTS budgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    total_amount_cents INTEGER NOT NULL,
    spent_cents INTEGER DEFAULT 0,
    remaining_cents INTEGER,
    start_date TEXT NOT NULL,
    end_date TEXT,
    status TEXT DEFAULT 'active',  -- 'active', 'exhausted', 'expired'
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS budget_allocations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    budget_id INTEGER NOT NULL,
    simulation_id INTEGER,
    sweep_id INTEGER,
    allocated_cents INTEGER NOT NULL,
    spent_cents INTEGER DEFAULT 0,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (budget_id) REFERENCES budgets(id) ON DELETE CASCADE,
    FOREIGN KEY (simulation_id) REFERENCES simulations(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_ledger_entries_simulation_id ON ledger_entries(simulation_id);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_transaction_date ON ledger_entries(transaction_date);
CREATE INDEX IF NOT EXISTS idx_contest_bets_simulation_id ON contest_bets(simulation_id);
CREATE INDEX IF NOT EXISTS idx_contest_bets_contest ON contest_bets(contest);
CREATE INDEX IF NOT EXISTS idx_prize_rules_contest ON prize_rules(contest);
CREATE INDEX IF NOT EXISTS idx_simulation_finances_roi ON simulation_finances(roi_percentage DESC);

-- Insert default bet cost (2.50 BRL = 250 cents)
INSERT OR IGNORE INTO bet_costs (effective_from, cost_cents, numbers_count, region, notes)
VALUES ('2020-01-01', 250, 5, 'BR', 'Default Quina bet cost');

-- Insert some sample prize rules (these would be updated with real data)
INSERT OR IGNORE INTO prize_rules (contest, prize_type, amount_cents, notes)
VALUES
(1, 'quina', 50000000, 'Sample quina prize - 500k BRL'),
(1, 'quadra', 500000, 'Sample quadra prize - 5k BRL'),
(1, 'terno', 10000, 'Sample terno prize - 100 BRL');

-- Down (commented):
-- DROP INDEX IF EXISTS idx_simulation_finances_roi;
-- DROP INDEX IF EXISTS idx_prize_rules_contest;
-- DROP INDEX IF EXISTS idx_contest_bets_contest;
-- DROP INDEX IF EXISTS idx_contest_bets_simulation_id;
-- DROP INDEX IF EXISTS idx_ledger_entries_transaction_date;
-- DROP INDEX IF EXISTS idx_ledger_entries_simulation_id;
-- DROP TABLE IF EXISTS budget_allocations;
-- DROP TABLE IF EXISTS budgets;
-- DROP TABLE IF EXISTS contest_bets;
-- DROP TABLE IF EXISTS simulation_finances;
-- DROP TABLE IF EXISTS ledger_entries;
-- DELETE FROM bet_costs WHERE region = 'BR' AND cost_cents = 250;
-- DROP TABLE IF EXISTS bet_costs;
-- DROP TABLE IF EXISTS prize_rules;