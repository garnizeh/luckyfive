-- schema: migrations/003_create_configs.sql

-- name: GetConfig :one
SELECT * FROM configs
WHERE id = ?
LIMIT 1;

-- name: GetConfigByName :one
SELECT * FROM configs
WHERE name = ?
LIMIT 1;

-- name: ListConfigs :many
SELECT * FROM configs
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListConfigsByMode :many
SELECT * FROM configs
WHERE mode = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetDefaultConfigForMode :one
SELECT * FROM configs
WHERE mode = ? AND is_default = 1
LIMIT 1;

-- name: InsertConfig :one
INSERT INTO configs (name, description, recipe_json, tags, mode, created_by)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateConfig :exec
UPDATE configs SET
  description = ?,
  recipe_json = ?,
  tags = ?,
  updated_at = CURRENT_TIMESTAMP,
  last_used_at = CURRENT_TIMESTAMP,
  times_used = times_used + 1
WHERE id = ?;

-- name: SetConfigAsDefault :exec
UPDATE configs SET is_default = CASE WHEN configs.id = ? THEN 1 ELSE 0 END
WHERE mode = (SELECT c.mode FROM configs c WHERE c.id = ?);

-- name: DeleteConfig :exec
DELETE FROM configs WHERE id = ?;

-- name: GetConfigPresets :many
SELECT * FROM config_presets
WHERE is_active = 1
ORDER BY sort_order ASC;

-- name: GetConfigPresetByName :one
SELECT * FROM config_presets
WHERE name = ? AND is_active = 1
LIMIT 1;
