-- +goose Up
DELETE FROM posts WHERE author_id IS NULL;
ALTER TABLE posts ALTER COLUMN author_id SET NOT NULL;

-- +goose Down
ALTER TABLE posts ALTER COLUMN author_id DROP NOT NULL;
