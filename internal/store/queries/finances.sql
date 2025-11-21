-- name: GetLedgerEntry :one
SELECT * FROM ledger
WHERE id = ?
LIMIT 1;

-- name: ListLedgerEntries :many
SELECT * FROM ledger
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- NOTE: ledger table used in production migration. The view `financial_summary` provides balances by account.

-- name: InsertLedgerEntry :one
INSERT INTO ledger (account, amount_cents, currency, description)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListLedgerForAccount :many
SELECT * FROM ledger WHERE account = ? ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: GetAccountBalance :one
SELECT account, balance_cents, entries FROM financial_summary WHERE account = ?;

-- name: SumLedgerBetweenDates :one
SELECT SUM(amount_cents) FROM ledger WHERE created_at BETWEEN ? AND ?;
