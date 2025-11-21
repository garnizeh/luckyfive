-- schema: migrations/001_create_results.sql

-- name: GetDraw :one
SELECT * FROM draws
WHERE contest = ?
LIMIT 1;

-- name: ListDraws :many
SELECT * FROM draws
ORDER BY contest DESC
LIMIT ? OFFSET ?;

-- name: ListDrawsByDateRange :many
SELECT * FROM draws
WHERE draw_date BETWEEN ? AND ?
ORDER BY contest DESC
LIMIT ? OFFSET ?;

-- name: ListDrawsByContestRange :many
SELECT * FROM draws
WHERE contest BETWEEN ? AND ?
ORDER BY contest ASC;

-- name: InsertDraw :exec
INSERT INTO draws (
  contest, draw_date, bola1, bola2, bola3, bola4, bola5, source, raw_row
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpsertDraw :exec
INSERT INTO draws (
  contest, draw_date, bola1, bola2, bola3, bola4, bola5, source, raw_row
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(contest) DO UPDATE SET
  draw_date = excluded.draw_date,
  bola1 = excluded.bola1,
  bola2 = excluded.bola2,
  bola3 = excluded.bola3,
  bola4 = excluded.bola4,
  bola5 = excluded.bola5,
  source = excluded.source,
  raw_row = excluded.raw_row,
  imported_at = CURRENT_TIMESTAMP;

-- name: CountDraws :one
SELECT COUNT(*) FROM draws;

-- name: GetContestRange :one
SELECT MIN(contest) as min_contest, MAX(contest) as max_contest
FROM draws;

-- name: InsertImportHistory :one
INSERT INTO import_history (
  filename, rows_inserted, rows_skipped, rows_errors, source_hash, metadata
) VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetImportHistory :many
SELECT * FROM import_history
ORDER BY imported_at DESC
LIMIT ? OFFSET ?;
