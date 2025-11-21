-- name: GetConfig :one
SELECT * FROM configs
WHERE id = ?
LIMIT 1;

-- name: ListConfigs :many
SELECT * FROM configs
ORDER BY id DESC
LIMIT ? OFFSET ?;
