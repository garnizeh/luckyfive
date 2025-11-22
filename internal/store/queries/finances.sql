-- Financial database queries for comprehensive financial tracking
-- Schema: migrations/004_create_finances.sql

-- Ledger operations
-- name: CreateLedgerEntry :one
INSERT INTO ledger_entries (
    transaction_date, transaction_type, amount_cents,
    simulation_id, contest, description, metadata_json
) VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetLedgerEntries :many
SELECT * FROM ledger_entries
WHERE simulation_id = ?
ORDER BY transaction_date DESC
LIMIT ? OFFSET ?;

-- name: GetLedgerBalance :one
SELECT COALESCE(SUM(amount_cents), 0) as balance
FROM ledger_entries
WHERE simulation_id = ?;

-- name: GetLedgerByDateRange :many
SELECT * FROM ledger_entries
WHERE transaction_date >= ? AND transaction_date <= ?
ORDER BY transaction_date DESC;

-- Prize rules
-- name: UpsertPrizeRule :exec
INSERT INTO prize_rules (
    contest, prize_type, amount_cents, winners, total_collected_cents, notes
) VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(contest, prize_type) DO UPDATE SET
    amount_cents = excluded.amount_cents,
    winners = excluded.winners,
    total_collected_cents = excluded.total_collected_cents,
    notes = excluded.notes;

-- name: GetPrizeRule :one
SELECT * FROM prize_rules
WHERE contest = ? AND prize_type = ?
LIMIT 1;

-- name: ListPrizeRules :many
SELECT * FROM prize_rules
WHERE contest >= ? AND contest <= ?
ORDER BY contest DESC, prize_type ASC;

-- Bet costs
-- name: CreateBetCost :one
INSERT INTO bet_costs (
    effective_from, effective_to, cost_cents, numbers_count, region, notes
) VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetActiveBetCost :one
SELECT * FROM bet_costs
WHERE effective_from <= ? AND (effective_to IS NULL OR effective_to >= ?)
    AND region = ? AND numbers_count = ?
ORDER BY effective_from DESC
LIMIT 1;

-- Contest bets
-- name: CreateContestBet :one
INSERT INTO contest_bets (
    simulation_id, contest, bet_numbers, cost_cents, hits, prize_type, prize_cents
) VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetContestBets :many
SELECT * FROM contest_bets
WHERE simulation_id = ?
ORDER BY contest ASC;

-- name: GetContestBetByContest :one
SELECT * FROM contest_bets
WHERE simulation_id = ? AND contest = ?
LIMIT 1;

-- name: UpdateContestBetPrize :exec
UPDATE contest_bets
SET hits = ?, prize_type = ?, prize_cents = ?
WHERE id = ?;

-- Simulation finances
-- name: CreateSimulationFinances :one
INSERT INTO simulation_finances (
    simulation_id, total_bets, total_cost_cents,
    total_prizes_cents, net_profit_cents, roi_percentage
) VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSimulationFinances :one
SELECT * FROM simulation_finances
WHERE simulation_id = ?
LIMIT 1;

-- name: UpdateSimulationFinances :exec
UPDATE simulation_finances
SET total_bets = ?,
    total_cost_cents = ?,
    total_prizes_cents = ?,
    net_profit_cents = ?,
    roi_percentage = ?,
    break_even_contest = ?,
    best_prize_cents = ?,
    best_prize_contest = ?,
    quina_wins = ?,
    quadra_wins = ?,
    terno_wins = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE simulation_id = ?;

-- name: ListTopSimulationsByROI :many
SELECT sf.*, s.recipe_name, s.mode
FROM simulation_finances sf
JOIN simulations s ON sf.simulation_id = s.id
WHERE s.status = 'completed'
ORDER BY sf.roi_percentage DESC
LIMIT ? OFFSET ?;

-- Budgets
-- name: CreateBudget :one
INSERT INTO budgets (
    name, description, total_amount_cents, start_date, end_date
) VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetBudget :one
SELECT * FROM budgets WHERE id = ? LIMIT 1;

-- name: UpdateBudgetSpent :exec
UPDATE budgets
SET spent_cents = ?,
    remaining_cents = total_amount_cents - ?,
    status = CASE
        WHEN total_amount_cents - ? <= 0 THEN 'exhausted'
        WHEN ? > end_date THEN 'expired'
        ELSE 'active'
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: AllocateBudget :one
INSERT INTO budget_allocations (
    budget_id, simulation_id, sweep_id, allocated_cents
) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateAllocationSpent :exec
UPDATE budget_allocations
SET spent_cents = ?
WHERE id = ?;

-- name: GetBudgetAllocations :many
SELECT * FROM budget_allocations
WHERE budget_id = ?;
