"""Redis connection settings for the arq worker."""

from arq.connections import RedisSettings

from app.core.config import settings


def make_redis_settings() -> RedisSettings:
    """Build arq's ``RedisSettings`` from the shared app Redis config."""
    return RedisSettings(
        host=settings.redis_host,
        port=settings.redis_port,
        database=settings.redis_db,
    )
