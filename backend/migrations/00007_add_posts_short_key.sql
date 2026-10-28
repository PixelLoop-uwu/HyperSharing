-- +goose Up
-- Публичный идентификатор поста в URL: /{short_key} (как у файлов /f/{short_key},
-- но без префикса). Ключ генерирует приложение при публикации.
ALTER TABLE posts ADD COLUMN short_key text;
-- Бэкфилл для уже опубликованных постов: 8 hex-символов из md5(id). Уникальность
-- здесь не гарантируется математически, но для существующих строк вероятность
-- коллизии ничтожна; новые ключи генерирует приложение криптостойко.
UPDATE posts SET short_key = substr(md5(id::text), 1, 8) WHERE short_key IS NULL;
ALTER TABLE posts ALTER COLUMN short_key SET NOT NULL;
ALTER TABLE posts ADD CONSTRAINT posts_short_key_key UNIQUE (short_key);

-- +goose Down
ALTER TABLE posts DROP CONSTRAINT posts_short_key_key;
ALTER TABLE posts DROP COLUMN short_key;
