-- Migration: 004_create_finances.sql
-- Creates tables for finances.db: ledger, financial_summary view

-- Up migration

-- Production-ready schema for finances DB
CREATE TABLE IF NOT EXISTS ledger (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  account TEXT NOT NULL,
  amount_cents INTEGER NOT NULL,
  currency TEXT NOT NULL DEFAULT 'USD',
  description TEXT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE VIEW IF NOT EXISTS financial_summary AS
SELECT account, SUM(amount_cents) AS balance_cents, COUNT(*) AS entries
FROM ledger
GROUP BY account;

CREATE INDEX IF NOT EXISTS idx_ledger_account ON ledger(account);

-- Down (commented):
-- DROP VIEW IF EXISTS financial_summary;
-- DROP INDEX IF EXISTS idx_ledger_account;
-- DROP TABLE IF EXISTS ledger;
