-- schema: migrations/002_create_simulations.sql

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

-- name: ListSimulationsByMode :many
SELECT * FROM simulations
WHERE mode = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: InsertSimulation :one
INSERT INTO simulations (
  recipe_name, recipe_json, mode, start_contest, end_contest,
  status, created_by
) VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateSimulationStatus :exec
UPDATE simulations SET
  status = ?,
  started_at = ?,
  finished_at = ?,
  run_duration_ms = ?,
  summary_json = ?,
  error_message = ?,
  error_stack = ?
WHERE id = ?;

-- name: UpdateSimulationArtifacts :exec
UPDATE simulations SET
  output_blob = ?,
  output_name = ?,
  log_blob = ?
WHERE id = ?;

-- name: DeleteSimulation :exec
DELETE FROM simulations WHERE id = ?;

-- name: CountSimulations :one
SELECT COUNT(*) FROM simulations;

-- name: GetSimulationByRecipeName :one
SELECT * FROM simulations
WHERE recipe_name = ?
LIMIT 1;

-- name: InsertSimulationContestResult :one
INSERT INTO simulation_contest_results (
  simulation_id, contest, actual_numbers, best_hits,
  best_prediction_index, best_prediction_numbers, predictions_json
) VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListSimulationContestResults :many
SELECT * FROM simulation_contest_results
WHERE simulation_id = ?
ORDER BY contest ASC
LIMIT ? OFFSET ?;

-- name: ListSimulationContestResultsByHits :many
SELECT * FROM simulation_contest_results
WHERE simulation_id = ? AND best_hits >= ?
ORDER BY best_hits DESC, contest ASC
LIMIT ? OFFSET ?;

-- name: CountSimulationContestResults :one
SELECT COUNT(*) FROM simulation_contest_results
WHERE simulation_id = ?;

-- name: GetSimulationContestResult :one
SELECT * FROM simulation_contest_results
WHERE simulation_id = ? AND contest = ?
LIMIT 1;

-- name: InsertAnalysisJob :one
INSERT INTO analysis_jobs (
  job_type, config_json, status, total_simulations
) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateAnalysisJobStatus :exec
UPDATE analysis_jobs SET
  status = ?,
  started_at = ?,
  finished_at = ?,
  completed_simulations = ?,
  failed_simulations = ?,
  top_configs_json = ?,
  error_message = ?
WHERE id = ?;

-- name: UpdateAnalysisJobProgress :exec
UPDATE analysis_jobs SET
  completed_simulations = ?,
  failed_simulations = ?
WHERE id = ?;

-- name: UpdateAnalysisJobArtifacts :exec
UPDATE analysis_jobs SET
  report_blob = ?,
  report_name = ?
WHERE id = ?;

-- name: GetAnalysisJob :one
SELECT * FROM analysis_jobs
WHERE id = ?
LIMIT 1;

-- name: ListAnalysisJobs :many
SELECT * FROM analysis_jobs
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAnalysisJobsByStatus :many
SELECT * FROM analysis_jobs
WHERE status = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListAnalysisJobsByType :many
SELECT * FROM analysis_jobs
WHERE job_type = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: DeleteAnalysisJob :exec
DELETE FROM analysis_jobs WHERE id = ?;

-- name: CountAnalysisJobs :one
SELECT COUNT(*) FROM analysis_jobs;
