-- +goose Up
CREATE TABLE sections (
  id       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id  uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  position int  NOT NULL,
  body     text NOT NULL DEFAULT '',
  UNIQUE (id, post_id),
  UNIQUE (post_id, position) DEFERRABLE INITIALLY DEFERRED
);

-- +goose Down
DROP TABLE sections;
