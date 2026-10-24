-- +goose Up
ALTER TABLE users ADD COLUMN avatar_version text;

-- +goose Down
ALTER TABLE users DROP COLUMN avatar_version;
