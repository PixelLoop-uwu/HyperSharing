-- name: GetFileByShortKey :one
SELECT id, post_id, name, mime_type, size, storage_key
FROM files
WHERE short_key = sqlc.arg(short_key);
