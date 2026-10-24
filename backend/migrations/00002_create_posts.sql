-- +goose Up
CREATE TABLE posts (
  id         uuid PRIMARY KEY, -- равен id драфта, задаётся приложением
  author_id  uuid REFERENCES users(id) ON DELETE CASCADE, -- NULL = аноним
  title      text NOT NULL DEFAULT '',
  total_size bigint NOT NULL DEFAULT 0, -- денормализация, считается при публикации
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX posts_author_idx ON posts (author_id, created_at DESC);
CREATE INDEX posts_anon_created_idx ON posts (created_at) WHERE author_id IS NULL;

-- +goose Down
DROP TABLE posts;
