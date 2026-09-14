"""Post-creation limits and the dependency that resolves them per request.

The limits themselves are a plain value object; ``get_post_limits`` is the
policy that turns the current request into the limits that apply to it
(rate-limiting anonymous callers and giving them the stricter set).
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Annotated

from fastapi import Depends, Request

from app.core.config import settings
from app.core.exceptions import RateLimitError
from app.features.users.deps import CurrentUserDep
from app.infrastructure.cache.rate_limiter import RateLimiterDep


@dataclass(frozen=True)
class PostLimits:
    max_files: int
    max_total_size: int
    max_file_size: int
    max_paragraphs: int

    @classmethod
    def for_logged_in(cls) -> PostLimits:
        return cls(
            settings.post_max_files,
            settings.post_max_total_size,
            settings.post_max_file_size,
            settings.post_max_paragraphs,
        )

    @classmethod
    def for_anonymous(cls) -> PostLimits:
        return cls(
            settings.anon_post_max_files,
            settings.anon_post_max_total_size,
            settings.anon_post_max_file_size,
            settings.anon_post_max_paragraphs,
        )


async def get_post_limits(
    request: Request,
    rate_limiter: RateLimiterDep,
    current_user: CurrentUserDep,
) -> PostLimits:
    """Resolve the limits for creating a post in this request.

    Authenticated sessions get the full limits and skip the anonymous creation
    rate limit. Anonymous callers are rate-limited and get the stricter set.
    """
    if current_user is not None:
        return PostLimits.for_logged_in()

    client_ip = request.client.host if request.client else "unknown"
    result = await rate_limiter.hit(
        f"rate:anon:draft:{client_ip}",
        limit=settings.anon_draft_rate_limit,
        window_seconds=settings.anon_draft_rate_window_seconds,
    )
    if not result.allowed:
        raise RateLimitError(
            f"anonymous draft rate limit exceeded "
            f"({settings.anon_draft_rate_limit} per "
            f"{settings.anon_draft_rate_window_seconds}s)",
            headers={"Retry-After": str(result.retry_after_seconds)},
        )
    return PostLimits.for_anonymous()


PostLimitsDep = Annotated[PostLimits, Depends(get_post_limits)]
