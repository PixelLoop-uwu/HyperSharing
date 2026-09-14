# FileSharing backend

Imgur-like сервис, но с загрузкой файлов любого типа (не только изображений).
Бэкенд на FastAPI.

## Что реализовано

- **Аутентификация** (`/auth`) — регистрация / логин / логаут / `me`. Сессия —
  opaque-токен в httpOnly cookie (`sid`), хранится в Redis со sliding-TTL
  (7 дней), без refresh-токенов.
- **Посты** (`/posts`) — черновик с presigned-URL на multipart-загрузку файлов
  напрямую в MinIO, публикация с проверкой целостности, чтение списка и по id.
- **Владение постами** — залогиненный пост получает `user_id`, анонимный
  остаётся `NULL`; публиковать черновик может только его создатель.
- **Дедупликация файлов** — байты лежат в S3 по content-hash
  (`objects/{sha256}`), в БД таблица `files` с уникальным `hash`: одинаковый
  файл хранится один раз и шарится между постами.
- **Idempotency publish** — `posts.draft_id` (UNIQUE) мапит черновик 1:1 в пост:
  повторный `publish` того же черновика (retry/таймаут) возвращает уже созданный
  пост, а не дублирует и не падает с 404.
- **Лимиты и rate-limit** — залогиненные получают полные лимиты, анонимы —
  урезанные + rate-limit на создание черновиков; логин защищён rate-limit'ом по
  email.
- **Фоновая джоба** (arq) — `cleanup_abandoned_uploads` раз в 10 минут делает
  `abort_multipart_upload` для брошенных multipart-загрузок (источник правды —
  S3, не Redis).
- **Health-check** — `GET /healthz`, реально проверяет DB/Redis/S3.
- **Логирование** (loguru) с `request_id` через middleware.
- **CI** — GitHub Actions: ruff + mypy + pytest на каждый push/PR.

## Стек

- **FastAPI** + **SQLAlchemy (async)** + **PostgreSQL** (asyncpg)
- **Redis** (redis.asyncio) — кэш, сессии, временные токены
- **S3-совместимое хранилище (MinIO)** через **aioboto3**
- **arq** — фоновые задачи на Redis
- **loguru** — структурированные логи с `request_id`
- **Alembic** — миграции (не `create_all()`)

## Структура

```
app/
  core/                  # не знает о внешнем мире: config, exceptions, security, logging
  middleware.py          # request_id для логов
  infrastructure/        # конкретные интеграции
    db/                  #   engine + session + BaseRepository + Base
    cache/               #   redis_client + BaseRedisStore + rate_limiter
    http/                #   именованные httpx-клиенты
    storage/             #   s3_client + BaseS3Store + MultipartUploadManager
    jobs/                #   arq-интеграция (WorkerSettings, redis, lifecycle)
    lifespan.py          #   композиция всех lifespan'ов через AsyncExitStack
  api/                   # health-check, агрегатный api_router
  features/
    users/               #   модели/сервис/роутер: auth, сессии
    posts/               #   модели/сервис/роутер: черновик, публикация, чтение
    storage/             #   файлы (content-addressed) + схема S3-ключей
  main.py
alembic/                 # миграции
tests/                   # unit-тесты (pytest)
docker-compose.yml       # postgres + redis + minio одной командой
```

Ключевое правило: `core/` не знает про Redis/БД/httpx/S3; каждый инфраструктурный
модуль владеет своим lifespan + `Depends` рядом со своим кодом. `AsyncSession`
никогда не хранится в `app.state` — только `engine` и `sessionmaker`; сама сессия
создаётся per-request.

## Быстрый старт

```bash
# 1. Поднять окружение (Postgres + Redis + MinIO)
docker compose up -d

# 2. Создать виртуальное окружение и установить зависимости
uv sync

# 3. Настроить окружение
cp .env.example .env

# 4. Применить миграции
uv run alembic upgrade head

# 5. Запустить API
uv run uvicorn app.main:app --reload
```

Health-check: `GET http://localhost:8000/healthz`.

## API

Все маршруты, кроме `/healthz`, смонтированы под префиксом `/api/v1`.

### Ошибки

Ошибки приложения возвращаются в едином формате:

```json
{"error": {"code": "not_found", "detail": "post ... not found"}}
```

Стабильные коды: `not_found` (404), `unauthorized` (401), `forbidden` (403),
`conflict` (409), `validation_error` (422), `rate_limited` (429, с заголовком
`Retry-After`), `storage_error` (502). Ошибки валидации Pydantic — стандартные 422
FastAPI.

### Аутентификация — `/auth`

| Метод | Путь             | Тело запроса    | Ответ              |
|-------|------------------|-----------------|--------------------|
| POST  | `/auth/register` | `RegisterIn`    | 201 `UserRead`     |
| POST  | `/auth/login`    | `LoginIn`       | 200 `UserRead`     |
| POST  | `/auth/logout`   | —               | 204 (без тела)     |
| GET   | `/auth/me`       | —               | 200 `UserRead`     |

`register`/`login` ставят httpOnly cookie `sid` (samesite=lax); `logout` её
удаляет. `me` без валидной сессии — 401.

**`RegisterIn`**

```json
{
  "email": "user@example.com",   // ≤255, формат email
  "username": "alice",           // ^[a-zA-Z0-9_]{3,32}$
  "password": "password123"      // 8..128
}
```

**`LoginIn`**

```json
{"email": "user@example.com", "password": "password123"}
```

**`UserRead`**

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "username": "alice",
  "created_at": "2026-09-15T12:00:00Z"
}
```

### Посты — `/posts`

| Метод | Путь               | Тело запроса        | Ответ                          |
|-------|--------------------|---------------------|--------------------------------|
| POST  | `/posts/draft`     | `PostDraftCreate`   | 201 `PostDraftRead`            |
| POST  | `/posts/publish`   | `PostPublishIn`     | 201 `PostFullRead`             |
| GET   | `/posts/{post_id}` | —                   | 200 `PostFullRead` / 404       |
| GET   | `/posts`           | query: `offset`,`limit` | 200 `PostRead[]`           |

`GET /posts` — пагинация `offset` (≥0, по умолч. 0) и `limit` (1..100, по умолч.
100).

**`PostDraftCreate`** — структура поста; блоки различаются полем `kind`:

```json
{
  "title": "опционально",              // ≤255
  "items": [
    {"kind": "paragraph", "text": "…"},
    {
      "kind": "file",
      "hash": "64-символьный hex sha256",
      "size": 123,
      "content_type": "text/plain",     // опционально
      "part_checksums": ["64 hex sha256", "…"]
    }
  ]
}
```

`part_checksums` — SHA256 (hex) байт каждой части по порядку; число записей должно
равняться числу частей, которые сервер выводит из `size` и `part_size`
(8 MiB по умолчанию). Один и тот же файл (по `hash`), вставленный несколько раз,
считается один раз и к лимитам, и к multipart-загрузке.

**`PostDraftRead`**

```json
{
  "draft_id": "uuid",
  "expires_at": "2026-09-15T13:00:00Z",
  "files": [
    {
      "hash": "64 hex",
      "part_size": 8388608,
      "parts": [
        {
          "part_number": 1,
          "url": "https://… presigned upload_part …",
          "checksum_sha256": "base64-encoded sha256"
        }
      ]
    }
  ]
}
```

**Загрузка части (клиент → S3 напрямую).** На каждый `url` из `parts` клиент
делает `PUT` байтами части с заголовками:

- `x-amz-checksum-sha256: <checksum_sha256>`
- `x-amz-sdk-checksum-algorithm: SHA256`

MinIO отбракует часть, чьи байты не совпали с подписанным checksum. Из ответа
клиент берёт заголовок `ETag` (в кавычках) — его нужно вернуть в `publish`.

**`PostPublishIn`**

```json
{
  "draft_id": "uuid",
  "files": [
    {"hash": "64 hex", "parts": [{"part_number": 1, "etag": "\"…\""}]}
  ]
}
```

**`PostFullRead`**

```json
{
  "id": "uuid",
  "title": "string|null",
  "user_id": "uuid|null",
  "items": [
    {"kind": "paragraph", "text": "…"},
    {
      "kind": "file",
      "hash": "64 hex",
      "size": 123,
      "content_type": "text/plain",
      "url": "https://… presigned get_object …"
    }
  ],
  "created_at": "2026-09-15T12:00:00Z"
}
```

**`PostRead`** (элемент списка)

```json
{
  "id": "uuid",
  "title": "string|null",
  "user_id": "uuid|null",
  "created_at": "…",
  "updated_at": "…"
}
```

### Полный флоу загрузки

1. `POST /posts/draft` — структура поста + по каждому файлу (`size`, `hash`,
   `part_checksums`). Сервер проверяет лимиты, открывает multipart-upload для
   каждого уникального файла и отдаёт presigned-URL на каждую часть.
2. Клиент `PUT`-ит части напрямую в MinIO по этим URL (с checksum-заголовками).
3. `POST /posts/publish` — `draft_id` + ETag каждой части. Сервер вызывает
   `complete_multipart_upload` (с предварительной сверкой размера через
   `list_parts`), пишет `posts`/`post_items`/`files`, удаляет черновик из Redis и
   кладёт пост в кэш.

## Фоновые задачи (arq)

Воркер — отдельный процесс на том же Redis, что и приложение; инфраструктурные
клиенты поднимает через тот же `populate_container`, что и API. Запуск из корня
backend:

```bash
uv run arq app.infrastructure.jobs.worker.WorkerSettings
```

Задачи — бизнес-логика, поэтому лежат в фичах (`features/posts/jobs.py`); пакет
`infrastructure/jobs/` держит только arq-интеграцию (`WorkerSettings`,
redis-настройки, lifecycle). Реализована:

- `cleanup_abandoned_uploads` — каждые 10 минут (`cron`) делает
  `abort_multipart_upload` для загрузок, чей `Initiated` старше
  `abandoned_upload_max_age_seconds`. Ищет через `list_multipart_uploads`, потому
  что TTL черновика уже удалил `upload_id` из Redis — S3 остаётся источником
  правды.

## Конфигурация

Все настройки — через переменные окружения / `.env` (см. `.env.example`).
Имена регистронезависимы, поэтому `postgres_user` == `POSTGRES_USER`.

## Архитектурные решения и trade-off'ы

- **Redis для черновиков постов, а не БД.** Черновик — короткоживущее,
  эфемерное состояние с TTL; Redis даёт дешёвый expiry и не плодит мусор в
  основной БД. В БД попадает только опубликованный пост.
- **Chunked (multipart) upload напрямую в S3.** Тяжёлые байты идут от клиента
  в MinIO по presigned URL, минуя backend — он не тратит RAM/CPU/пропускную
  способность. Части грузятся параллельно и докачиваются по отдельности.
- **Per-part SHA256 checksum + проверка размера.** Клиент присылает SHA256 каждой
  части в `POST /post/draft`, сервер base64-кодирует и подписывает его в presigned
  URL `upload_part` (`x-amz-checksum-sha256`), а MinIO отбраковывает часть, чьи
  байты не совпали, прямо на этапе загрузки. Плюс сверка итогового размера
  (`list_parts` до `complete_multipart_upload`). `ChecksumAlgorithm` на
  `create_multipart_upload` не задаём — в MinIO это ломает `CompleteMultipartUpload`
  (падает с `InvalidPart`), тогда как per-part проверка и так работает на `upload_part`.
- **Хеш всего файла — всё ещё доверенные метаданные клиента.** Per-part checksum
  ловит порчу при передаче и подмену части, но не доказывает, что объект целиком
  хешируется в заявленный `hash` (sha256 не композиционален). Это потребовало бы
  `ChecksumType=FULL_OBJECT` (MinIO не поддерживает) или worker со скачиванием
  файла — от него осознанно отказались.
- **Generic-репозитории только под реальное дублирование.** `BaseRepository`
  (SQL), `BaseRedisStore` (string k/v) и `BaseS3Store` (put/get/delete) покрывают
  один повторяющийся паттерн. Сложные запросы, hash/rate-limit/locks и
  multipart — отдельные классы, а не попытки предугадать все кейсы заранее.

## Что дальше

- **Интеграционные тесты** (testcontainers: реальные Postgres/Redis/MinIO).
- **Разделить `PostService` на read/write** — `PostReadService` (`get_full`,
  `list_posts`) и `PostWriteService` (`create_draft`, `publish`). Draft и publish
  отдельными классами не дробить: это одна жизнь черновика, граница уже есть
  (`Draft` в Redis). `_presign_get_url` вынести в общий хелпер. Делать, когда
  начнутся unit-тесты read-пути или фильтры/пагинация.
- **Structured logging** — сейчас `loguru` + `request_id`; можно поднять до
  `structlog`, если нужно.
