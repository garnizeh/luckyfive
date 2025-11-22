-- name: CreateSweep :one
INSERT INTO sweeps (
    name, description, config_json, created_by
) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetSweep :one
SELECT * FROM sweeps
WHERE id = ?
LIMIT 1;

-- name: GetSweepByName :one
SELECT * FROM sweeps
WHERE name = ?
LIMIT 1;

-- name: ListSweeps :many
SELECT * FROM sweeps
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateSweep :exec
UPDATE sweeps
SET name = ?,
    description = ?,
    config_json = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteSweep :exec
DELETE FROM sweeps
WHERE id = ?;

-- name: IncrementSweepUsage :exec
UPDATE sweeps
SET times_used = times_used + 1,
    last_used_at = CURRENT_TIMESTAMP
WHERE id = ?;