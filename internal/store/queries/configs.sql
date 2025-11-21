-- name: CreateConfig :one
INSERT INTO configs (
    name, description, recipe_json, tags, is_default, mode, created_by
) VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

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
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: ListConfigsByMode :many
SELECT * FROM configs
WHERE mode = ?
ORDER BY times_used DESC, name ASC
LIMIT ? OFFSET ?;

-- name: UpdateConfig :exec
UPDATE configs
SET description = ?,
    recipe_json = ?,
    tags = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteConfig :exec
DELETE FROM configs
WHERE id = ?;

-- name: SetDefaultConfig :exec
UPDATE configs
SET is_default = CASE WHEN id = ? THEN 1 ELSE 0 END
WHERE mode = ?;

-- name: GetDefaultConfig :one
SELECT * FROM configs
WHERE is_default = 1 AND mode = ?
LIMIT 1;

-- name: IncrementConfigUsage :exec
UPDATE configs
SET times_used = times_used + 1,
    last_used_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetPreset :one
SELECT * FROM config_presets
WHERE name = ?
LIMIT 1;

-- name: ListPresets :many
SELECT * FROM config_presets
WHERE is_active = 1
ORDER BY sort_order ASC;
