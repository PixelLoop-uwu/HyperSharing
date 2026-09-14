"""Redis client, its lifespan and dependency."""

from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from typing import Annotated

import redis.asyncio as aioredis
from fastapi import Depends

from app.core.config import settings
from app.infrastructure.container import Container, ContainerDep, require_not_none


async def get_redis(container: ContainerDep) -> aioredis.Redis:
    return require_not_none(container.redis, "redis")


RedisDep = Annotated[aioredis.Redis, Depends(get_redis)]


@asynccontextmanager
async def redis_lifespan(container: Container) -> AsyncIterator[None]:
    client = aioredis.Redis(
        host=settings.redis_host,
        port=settings.redis_port,
        db=settings.redis_db,
        decode_responses=True,
    )
    container.redis = client
    try:
        yield
    finally:
        await client.aclose()
