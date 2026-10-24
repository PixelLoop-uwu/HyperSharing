-- +goose Up
CREATE TABLE files (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id     uuid NOT NULL,
  section_id  uuid NOT NULL,
  position    int  NOT NULL,
  name        text NOT NULL, -- оригинальное имя, только для отображения
  mime_type   text NOT NULL,
  size        bigint NOT NULL,
  hash        text, -- sha256 hex, необязателен, в v1 не проверяется
  storage_key text NOT NULL, -- ключ объекта в MinIO
  short_key   text NOT NULL UNIQUE, -- ~6 символов в публичном URL
  created_at  timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (section_id, post_id) REFERENCES sections(id, post_id) ON DELETE CASCADE,
  UNIQUE (section_id, position) DEFERRABLE INITIALLY DEFERRED
);
CREATE INDEX files_post_idx ON files (post_id);

-- +goose Down
DROP TABLE files;
