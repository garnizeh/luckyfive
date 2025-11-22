-- name: CreateSweepJob :one
INSERT INTO sweep_jobs (
    name, description, sweep_config_json, base_contest_range,
    total_combinations, created_by
) VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSweepJob :one
SELECT * FROM sweep_jobs WHERE id = ? LIMIT 1;

-- name: ListSweepJobs :many
SELECT * FROM sweep_jobs
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateSweepJobProgress :exec
UPDATE sweep_jobs
SET completed_simulations = ?,
    failed_simulations = ?,
    status = ?
WHERE id = ?;

-- name: FinishSweepJob :exec
UPDATE sweep_jobs
SET status = ?,
    finished_at = ?,
    run_duration_ms = ?
WHERE id = ?;

-- name: CreateSweepSimulation :exec
INSERT INTO sweep_simulations (
    sweep_job_id, simulation_id, variation_index, variation_params
) VALUES (?, ?, ?, ?);

-- name: GetSweepSimulations :many
SELECT * FROM sweep_simulations
WHERE sweep_job_id = ?
ORDER BY variation_index ASC;

-- name: GetSweepSimulationDetails :many
SELECT
    ss.*,
    s.status,
    s.summary_json,
    s.run_duration_ms
FROM sweep_simulations ss
JOIN simulations s ON ss.simulation_id = s.id
WHERE ss.sweep_job_id = ?
ORDER BY ss.variation_index ASC;

-- name: CreateComparison :one
INSERT INTO comparisons (
    name, description, simulation_ids, metric
) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetComparison :one
SELECT * FROM comparisons WHERE id = ? LIMIT 1;

-- name: UpdateComparisonResult :exec
UPDATE comparisons
SET result_json = ?
WHERE id = ?;

-- name: InsertComparisonMetric :exec
INSERT INTO comparison_metrics (
    comparison_id, simulation_id, metric_name, metric_value, rank, percentile
) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetComparisonMetrics :many
SELECT * FROM comparison_metrics
WHERE comparison_id = ?
ORDER BY rank ASC;