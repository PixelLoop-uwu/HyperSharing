-- +goose Up
CREATE TABLE users (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  username       text NOT NULL,
  email          text NOT NULL,
  password_hash  text NOT NULL,
  email_verified boolean NOT NULL DEFAULT false,
  is_admin       boolean NOT NULL DEFAULT false,
  created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_username_uniq ON users (lower(username));
CREATE UNIQUE INDEX users_email_uniq ON users (lower(email));

-- +goose Down
DROP TABLE users;
