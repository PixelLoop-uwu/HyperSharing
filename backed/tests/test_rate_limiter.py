"""Unit tests for RateLimiter (the fixed-window Redis counter)."""

from app.infrastructure.cache.rate_limiter import RateLimiter


class FakeRedis:
    """Stands in for ``redis.asyncio.Redis``.

    The real rate limiter only calls one method, ``eval``, with a Lua script
    and arguments. This fake records those arguments and returns a scripted
    ``[count, ttl]``, so no real Redis is needed.
    """

    def __init__(self, count: int, ttl: int) -> None:
        self._count = count
        self._ttl = ttl
        self.last_key: str | None = None
        self.last_window: str | None = None

    async def eval(self, script: str, numkeys: int, *args):
        # Mirrors redis.asyncio: eval(script, numkeys, *keys_and_args).
        self.last_key = args[0]
        self.last_window = args[1]
        return [self._count, self._ttl]


async def test_hit_allows_under_limit() -> None:
    redis = FakeRedis(count=1, ttl=3600)

    result = await RateLimiter(redis).hit("rate:k", limit=10, window_seconds=3600)

    assert result.allowed
    assert result.remaining == 9
    assert result.retry_after_seconds == 3600


async def test_hit_blocks_over_limit() -> None:
    redis = FakeRedis(count=11, ttl=42)

    result = await RateLimiter(redis).hit("rate:k", limit=10, window_seconds=3600)

    assert not result.allowed
    assert result.remaining == 0
    assert result.retry_after_seconds == 42


async def test_hit_passes_key_and_window_to_redis() -> None:
    redis = FakeRedis(count=1, ttl=3600)

    await RateLimiter(redis).hit("my-key", limit=10, window_seconds=123)

    assert redis.last_key == "my-key"
    assert redis.last_window == "123"
