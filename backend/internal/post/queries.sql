-- name: CreatePost :one
INSERT INTO posts (id, author_id, title, total_size)
VALUES (sqlc.arg(id), sqlc.arg(author_id), sqlc.arg(title), sqlc.arg(total_size))
RETURNING id, author_id, title, total_size, created_at;

-- name: CreateSection :one
INSERT INTO sections (id, post_id, position, body)
VALUES (sqlc.arg(id), sqlc.arg(post_id), sqlc.arg(position), sqlc.arg(body))
RETURNING id, post_id, position, body;

-- name: CreateFile :one
INSERT INTO files (id, post_id, section_id, position, name, mime_type, size, storage_key, short_key)
VALUES (sqlc.arg(id), sqlc.arg(post_id), sqlc.arg(section_id), sqlc.arg(position), sqlc.arg(name), sqlc.arg(mime_type), sqlc.arg(size), sqlc.arg(storage_key), sqlc.arg(short_key))
RETURNING id, post_id, section_id, position, name, mime_type, size, storage_key, short_key, created_at;

-- name: GetPost :one
SELECT p.id, p.author_id, p.title, p.total_size, p.created_at, u.username, u.avatar_version
FROM posts p
LEFT JOIN users u ON u.id = p.author_id
WHERE p.id = sqlc.arg(id);

-- name: GetSectionsByPost :many
SELECT id, post_id, position, body
FROM sections
WHERE post_id = sqlc.arg(post_id)
ORDER BY position;

-- name: GetFilesByPost :many
SELECT id, post_id, section_id, position, name, mime_type, size, storage_key, short_key, created_at
FROM files
WHERE post_id = sqlc.arg(post_id)
ORDER BY position;

-- name: DeletePost :one
DELETE FROM posts
WHERE id = sqlc.arg(id)
RETURNING id;

-- name: GetUserStorageUsed :one
SELECT COALESCE(SUM(f.size), 0)::bigint AS total
FROM files f
JOIN posts p ON p.id = f.post_id
WHERE p.author_id = sqlc.arg(author_id);

-- name: DeleteExpiredAnonPosts :exec
DELETE FROM posts
WHERE author_id IS NULL AND created_at < sqlc.arg(cutoff);

-- name: AllStorageKeys :many
SELECT storage_key FROM files;

-- name: ListPostsByAuthor :many
SELECT id, title, total_size, created_at
FROM posts
WHERE author_id = sqlc.arg(author_id)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListPostsByAuthorAfter :many
SELECT id, title, total_size, created_at
FROM posts
WHERE author_id = sqlc.arg(author_id)
  AND (created_at, id) < (sqlc.arg(cursor_created_at)::timestamptz, sqlc.arg(cursor_id)::uuid)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_limit);
