"""Application settings.

Loaded from environment variables and an optional ``.env`` file.
Field names are case-insensitive, so ``postgres_user`` maps to ``POSTGRES_USER``.
"""

from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    # Application.
    app_name: str = "FileSharing"
    debug: bool = False
    api_v1_prefix: str = "/api/v1"
    request_id_header: str = "X-Request-ID"

    # PostgreSQL.
    postgres_user: str = "postgres"
    postgres_password: str = "123qwe"
    postgres_host: str = "localhost"
    postgres_port: int = 5432
    postgres_db: str = "dev"

    # Redis.
    redis_host: str = "localhost"
    redis_port: int = 6379
    redis_db: int = 0

    # S3-compatible storage (MinIO).
    s3_endpoint_url: str = "http://localhost:9000"
    # Optional public endpoint used only to presign URLs for end users.
    # Unset (default) means presigned URLs reuse ``s3_endpoint_url``.
    s3_public_endpoint_url: str | None = None
    s3_access_key_id: str = "minioadmin"
    s3_secret_access_key: str = "123qwe123"
    s3_region: str = "us-east-1"
    s3_bucket: str = "filesharing"
    # Multipart upload part size (MinIO requires >= 5 MiB for non-last parts).
    s3_multipart_part_size: int = 8 * 1024 * 1024
    # Lifetime of a presigned URL for get_object/upload_part.
    s3_presign_ttl_seconds: int = 3600

    # HTTP client (outbound requests to other services).
    http_default_client_name: str = "default"
    http_timeout_seconds: float = 30.0

    # Drafts (post upload flow).
    draft_ttl_seconds: int = 3600
    # An in-progress multipart upload is aborted by the cleanup job once it
    # has been incomplete this long. Defaults to the draft TTL: an expired
    # draft can never be published, so its upload is garbage.
    abandoned_upload_max_age_seconds: int = 3600

    # Published post cache (Redis). Served on GET /posts/{id} to skip the DB.
    # Posts are immutable after publish, so the only thing that expires is the
    # presigned URLs inside the cached payload — keep this TTL below
    # s3_presign_ttl_seconds so a cached entry never serves a dead URL.
    post_cache_ttl_seconds: int = 300

    # Post size limits (enforced at draft creation, before any S3 upload).
    # Files are counted by unique content hash, so the same file embedded
    # multiple times counts once toward these limits.
    post_max_files: int = 20
    post_max_total_size: int = 100 * 1024 * 1024  # bytes, sum of unique files
    post_max_file_size: int = 50 * 1024 * 1024  # bytes, single file
    post_max_paragraphs: int = 100  # text blocks in the post

    # Anonymous users — stricter post limits and a creation rate limit.
    # Applies when there is no authenticated session (see Sessions / auth).
    anon_post_max_files: int = 5
    anon_post_max_total_size: int = 20 * 1024 * 1024  # bytes
    anon_post_max_file_size: int = 10 * 1024 * 1024  # bytes
    anon_post_max_paragraphs: int = 50
    anon_draft_rate_limit: int = 10  # draft creations per window
    anon_draft_rate_window_seconds: int = 3600

    # Background jobs (arq worker).
    jobs_max_concurrent: int = 4
    job_timeout_seconds: int = 300

    # Validation.
    sha256_hex_pattern: str = r"^[0-9a-f]{64}$"
    email_pattern: str = r"^[^@\s]+@[^@\s]+\.[^@\s]+$"
    # Login name: letters, digits and underscore, 3-32 chars.
    username_pattern: str = r"^[a-zA-Z0-9_]{3,32}$"
    password_min_length: int = 8

    # Password hashing (PBKDF2-HMAC-SHA256).
    password_hash_algorithm: str = "pbkdf2_sha256"
    password_hash_iterations: int = 600_000
    password_salt_bytes: int = 16
    password_hash_bytes: int = 32

    # Sessions (opaque token in an httpOnly cookie, stored in Redis).
    session_cookie_name: str = "sid"
    session_ttl_seconds: int = 7 * 24 * 3600  # sliding: refreshed on each use
    # Set True behind HTTPS — the cookie is then only sent over TLS.
    session_cookie_secure: bool = False
    # Login brute-force protection (per email, not per IP — see users/service.py).
    login_rate_limit: int = 5
    login_rate_window_seconds: int = 900

    @property
    def database_url(self) -> str:
        """Async SQLAlchemy DSN."""
        return (
            f"postgresql+asyncpg://{self.postgres_user}:{self.postgres_password}"
            f"@{self.postgres_host}:{self.postgres_port}/{self.postgres_db}"
        )

    @property
    def redis_url(self) -> str:
        return f"redis://{self.redis_host}:{self.redis_port}/{self.redis_db}"


@lru_cache
def get_settings() -> Settings:
    return Settings()


settings = get_settings()
