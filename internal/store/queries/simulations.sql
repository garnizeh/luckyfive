-- name: CreateSimulation :one
INSERT INTO simulations (
    recipe_name, recipe_json, mode, start_contest, end_contest, created_by
) VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSimulation :one
SELECT * FROM simulations
WHERE id = ?
LIMIT 1;

-- name: ListSimulations :many
SELECT * FROM simulations
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListSimulationsByStatus :many
SELECT * FROM simulations
WHERE status = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateSimulationStatus :exec
UPDATE simulations
SET status = ?, started_at = ?, worker_id = ?
WHERE id = ? AND status = 'pending';

-- name: CompleteSimulation :exec
UPDATE simulations
SET status = 'completed',
    finished_at = ?,
    run_duration_ms = ?,
    summary_json = ?,
    output_blob = ?,
    output_name = ?
WHERE id = ?;

-- name: FailSimulation :exec
UPDATE simulations
SET status = 'failed',
    finished_at = ?,
    error_message = ?,
    error_stack = ?
WHERE id = ?;

-- name: CancelSimulation :exec
UPDATE simulations
SET status = 'cancelled',
    finished_at = ?
WHERE id = ? AND status IN ('pending', 'running');

-- name: ClaimPendingSimulation :one
UPDATE simulations
SET status = 'running',
    started_at = ?,
    worker_id = ?
WHERE id = (
    SELECT id FROM simulations
    WHERE status = 'pending'
    ORDER BY created_at ASC
    LIMIT 1
)
RETURNING *;

-- name: InsertContestResult :exec
INSERT INTO simulation_contest_results (
    simulation_id, contest, actual_numbers, best_hits,
    best_prediction_index, best_prediction_numbers, predictions_json
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetContestResults :many
SELECT * FROM simulation_contest_results
WHERE simulation_id = ?
ORDER BY contest ASC
LIMIT ? OFFSET ?;

-- name: GetContestResultsByMinHits :many
SELECT * FROM simulation_contest_results
WHERE simulation_id = ? AND best_hits >= ?
ORDER BY best_hits DESC, contest ASC
LIMIT ? OFFSET ?;

-- name: CountSimulationsByStatus :one
SELECT COUNT(*) FROM simulations
WHERE status = ?;
