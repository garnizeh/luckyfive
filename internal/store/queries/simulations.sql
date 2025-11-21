-- name: GetSimulation :one
SELECT * FROM simulations
WHERE id = ?
LIMIT 1;

-- name: ListSimulations :many
SELECT * FROM simulations
ORDER BY id DESC
LIMIT ? OFFSET ?;
