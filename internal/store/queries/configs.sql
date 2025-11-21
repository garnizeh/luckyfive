-- name: GetConfig :one
SELECT * FROM configs
WHERE id = ?
LIMIT 1;

-- name: ListConfigs :many
SELECT * FROM configs
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: InsertConfig :one
INSERT INTO configs (key, value, description)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateConfig :exec
UPDATE configs SET value = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE key = ?;

-- name: GetConfigByKey :one
SELECT * FROM configs WHERE key = ? LIMIT 1;

-- name: InsertPreset :one
INSERT INTO config_presets (name, mode, settings, is_default)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListPresetsByMode :many
SELECT * FROM config_presets WHERE mode = ? ORDER BY id DESC LIMIT ? OFFSET ?;

-- name: GetDefaultPresetForMode :one
SELECT * FROM config_presets WHERE mode = ? AND is_default = 1 LIMIT 1;
