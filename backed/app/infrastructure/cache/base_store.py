"""String key/value store with TTL and Pydantic serialization.

Covers exactly one pattern: caching, sessions, temporary tokens. Hash/Set/
Sorted Set, rate limiting, distributed locks and pub/sub are separate small
classes under the concrete task — not unified here.
"""

from typing import Generic, TypeVar

import redis.asyncio as aioredis
from pydantic import BaseModel

ModelType = TypeVar("ModelType", bound=BaseModel)


class BaseRedisStore(Generic[ModelType]):
    def __init__(
        self,
        redis: aioredis.Redis,
        model: type[ModelType],
        *,
        ttl: int | None = None,
    ) -> None:
        self._redis = redis
        self._model = model
        self._ttl = ttl

    async def get(self, key: str) -> ModelType | None:
        raw = await self._redis.get(key)
        if raw is None:
            return None
        return self._model.model_validate_json(raw)

    async def set(self, key: str, value: ModelType, *, ttl: int | None = None) -> None:
        effective_ttl = ttl if ttl is not None else self._ttl
        payload = value.model_dump_json()
        if effective_ttl is not None:
            await self._redis.set(key, payload, ex=effective_ttl)
        else:
            await self._redis.set(key, payload)

    async def delete(self, key: str) -> None:
        await self._redis.delete(key)

    async def exists(self, key: str) -> bool:
        return bool(await self._redis.exists(key))
