-- name: GetSimulation :one
SELECT * FROM simulations
WHERE id = ?
LIMIT 1;

-- name: ListSimulations :many
SELECT * FROM simulations
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: InsertSimulation :one
INSERT INTO simulations (name, description, settings, status)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateSimulation :exec
UPDATE simulations SET name = ?, description = ?, settings = ?, status = ? WHERE id = ?;

-- name: DeleteSimulation :exec
DELETE FROM simulations WHERE id = ?;

-- name: CreateAnalysisJob :one
INSERT INTO analysis_jobs (simulation_id, status, worker)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListAnalysisJobsBySimulation :many
SELECT * FROM analysis_jobs WHERE simulation_id = ? ORDER BY id DESC LIMIT ? OFFSET ?;

-- name: InsertSimulationResult :one
INSERT INTO simulation_contest_results (simulation_id, contest, result_blob)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListSimulationResults :many
SELECT * FROM simulation_contest_results WHERE simulation_id = ? ORDER BY id ASC LIMIT ? OFFSET ?;
