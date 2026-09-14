"""Unit tests for get_post_limits (the per-request post-limits dependency)."""

import uuid
from types import SimpleNamespace

import pytest

from app.core.exceptions import RateLimitError
from app.features.posts.limits import PostLimits, get_post_limits
from app.features.users.models import User
from app.infrastructure.cache.rate_limiter import RateLimitResult


class FakeRateLimiter:
    """Returns a scripted result and records the key/limit/window it was given."""

    def __init__(self, result: RateLimitResult) -> None:
        self._result = result
        self.last_key: str | None = None
        self.last_limit: int | None = None
        self.last_window: int | None = None

    async def hit(
        self, key: str, *, limit: int, window_seconds: int
    ) -> RateLimitResult:
        self.last_key = key
        self.last_limit = limit
        self.last_window = window_seconds
        return self._result


def _request(host: str | None) -> SimpleNamespace:
    """Build a minimal object shaped like a FastAPI Request (only ``.client``)."""
    client = SimpleNamespace(host=host) if host is not None else None
    return SimpleNamespace(client=client)


def _user() -> User:
    user = User(email="a@b.com", username="alice", password_hash="x")
    user.id = uuid.uuid4()
    return user


async def test_returns_anonymous_limits_when_allowed() -> None:
    limiter = FakeRateLimiter(
        RateLimitResult(allowed=True, remaining=9, retry_after_seconds=3600)
    )

    limits = await get_post_limits(_request("1.2.3.4"), limiter, None)

    assert limits == PostLimits.for_anonymous()
    assert limiter.last_key == "rate:anon:draft:1.2.3.4"


async def test_raises_rate_limit_error_with_retry_after() -> None:
    limiter = FakeRateLimiter(
        RateLimitResult(allowed=False, remaining=0, retry_after_seconds=42)
    )

    with pytest.raises(RateLimitError) as exc_info:
        await get_post_limits(_request("1.2.3.4"), limiter, None)

    assert exc_info.value.status_code == 429
    assert exc_info.value.headers == {"Retry-After": "42"}


async def test_falls_back_to_unknown_ip() -> None:
    limiter = FakeRateLimiter(
        RateLimitResult(allowed=True, remaining=9, retry_after_seconds=3600)
    )

    await get_post_limits(_request(None), limiter, None)

    assert limiter.last_key == "rate:anon:draft:unknown"


async def test_logged_in_gets_full_limits_without_rate_limit() -> None:
    limiter = FakeRateLimiter(
        RateLimitResult(allowed=True, remaining=9, retry_after_seconds=3600)
    )

    limits = await get_post_limits(_request("1.2.3.4"), limiter, _user())

    assert limits == PostLimits.for_logged_in()
    assert limiter.last_key is None
