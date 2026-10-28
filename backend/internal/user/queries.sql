-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES (sqlc.arg(username), sqlc.arg(email), sqlc.arg(password_hash))
RETURNING id, username, email, email_verified, is_admin, created_at;

-- name: GetUserByID :one
SELECT id, username, email, email_verified, is_admin, created_at, avatar_version
FROM users
WHERE id = sqlc.arg(id);

-- name: GetUserAuthByEmail :one
SELECT id, username, email, email_verified, is_admin, created_at, password_hash, avatar_version
FROM users
WHERE lower(email) = lower(sqlc.arg(email));

-- name: SetAvatarVersion :exec
UPDATE users SET avatar_version = sqlc.arg(version) WHERE id = sqlc.arg(id);

-- name: ClearAvatarVersion :exec
UPDATE users SET avatar_version = NULL WHERE id = sqlc.arg(id);

-- name: GetAvatarVersion :one
SELECT avatar_version FROM users WHERE id = sqlc.arg(id);

-- name: AvatarUserIDs :many
SELECT id, avatar_version FROM users WHERE avatar_version IS NOT NULL;
