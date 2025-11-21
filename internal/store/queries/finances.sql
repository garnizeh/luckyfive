-- name: GetFinance :one
SELECT * FROM finances
WHERE id = ?
LIMIT 1;

-- name: ListFinances :many
SELECT * FROM finances
ORDER BY id DESC
LIMIT ? OFFSET ?;
