-- name: CreateComparison :one
INSERT INTO comparisons (
    name, description, simulation_ids, metric
) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetComparison :one
SELECT * FROM comparisons WHERE id = ? LIMIT 1;

-- name: ListComparisons :many
SELECT * FROM comparisons
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateComparisonResult :exec
UPDATE comparisons
SET result_json = ?
WHERE id = ?;

-- name: DeleteComparison :exec
DELETE FROM comparisons WHERE id = ?;

-- name: InsertComparisonMetric :exec
INSERT INTO comparison_metrics (
    comparison_id, simulation_id, metric_name, metric_value, rank, percentile
) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetComparisonMetrics :many
SELECT * FROM comparison_metrics
WHERE comparison_id = ?
ORDER BY rank ASC;

-- name: GetComparisonMetricsBySimulation :many
SELECT * FROM comparison_metrics
WHERE simulation_id = ?
ORDER BY comparison_id ASC;