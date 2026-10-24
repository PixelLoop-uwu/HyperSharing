-- name: GetFileByShortKey :one
SELECT f.id, f.post_id, f.name, f.mime_type, f.size, f.storage_key, p.author_id, p.created_at
FROM files f
JOIN posts p ON p.id = f.post_id
WHERE f.short_key = sqlc.arg(short_key);
