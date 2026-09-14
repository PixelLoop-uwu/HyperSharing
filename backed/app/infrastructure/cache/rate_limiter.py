"""Fixed-window rate limiter backed by Redis.

Rate limiting is a small standalone class, not ``BaseRedisStore`` — a counter
with expiry is a different pattern from the string k/v cache.
"""

from dataclasses import dataclass
from typing import Annotated, Any, cast

import redis.asyncio as aioredis
from fastapi import Depends

from app.infrastructure.cache.redis_client import RedisDep

# INCR + EXPIRE atomically, so the TTL is set exactly once on the first hit.
_INCR_EXPIRE_LUA = """
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return {count, redis.call('TTL', KEYS[1])}
"""


@dataclass(frozen=True)
class RateLimitResult:
    allowed: bool
    remaining: int
    retry_after_seconds: int


class RateLimiter:
    def __init__(self, redis: aioredis.Redis) -> None:
        self._redis = redis

    async def hit(
        self, key: str, *, limit: int, window_seconds: int
    ) -> RateLimitResult:
        """Record one request for ``key`` and report the window state."""
        # redis-py types eval() as `Awaitable[str] | str` (the method is shared
        # by the sync and async clients); cast through Any so strict mypy accepts
        # the await + unpack, and pass window_seconds as a str like the stub wants.
        count, ttl = await cast(
            Any, self._redis.eval(_INCR_EXPIRE_LUA, 1, key, str(window_seconds))
        )
        count = int(count)
        ttl = int(ttl)
        return RateLimitResult(
            allowed=count <= limit,
            remaining=max(0, limit - count),
            retry_after_seconds=ttl,
        )


async def get_rate_limiter(redis: RedisDep) -> RateLimiter:
    return RateLimiter(redis)


RateLimiterDep = Annotated[RateLimiter, Depends(get_rate_limiter)]
