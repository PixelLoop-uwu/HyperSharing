-- Plain SQL schema for sqlc. It must stay in sync with migrations/; sqlc does not
-- understand goose annotations, so migrations are the source of truth and this file
-- is a mirror for code generation only. Keep tables here matching the applied schema.

CREATE TABLE users (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  username       text NOT NULL,
  email          text NOT NULL,
  password_hash  text NOT NULL,
  email_verified boolean NOT NULL DEFAULT false,
  is_admin       boolean NOT NULL DEFAULT false,
  avatar_version text,
  created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_username_uniq ON users (lower(username));
CREATE UNIQUE INDEX users_email_uniq ON users (lower(email));

CREATE TABLE posts (
  id         uuid PRIMARY KEY,
  author_id  uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  short_key  text NOT NULL UNIQUE,
  title      text NOT NULL DEFAULT '',
  total_size bigint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX posts_author_idx ON posts (author_id, created_at DESC);
CREATE INDEX posts_anon_created_idx ON posts (created_at) WHERE author_id IS NULL;

CREATE TABLE sections (
  id       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id  uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  position int  NOT NULL,
  body     text NOT NULL DEFAULT '',
  UNIQUE (id, post_id),
  UNIQUE (post_id, position) DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE files (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id     uuid NOT NULL,
  section_id  uuid NOT NULL,
  position    int  NOT NULL,
  name        text NOT NULL,
  mime_type   text NOT NULL,
  size        bigint NOT NULL,
  hash        text,
  storage_key text NOT NULL,
  short_key   text NOT NULL UNIQUE,
  created_at  timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (section_id, post_id) REFERENCES sections(id, post_id) ON DELETE CASCADE,
  UNIQUE (section_id, position) DEFERRABLE INITIALLY DEFERRED
);
CREATE INDEX files_post_idx ON files (post_id);
