"""Typed container for application-scoped infrastructure clients.

Created once at startup and stored on ``app.state.container``. Each
infrastructure lifespan fills its own field; dependencies read typed fields
from it. This replaces loosely-typed ``app.state.<attr>`` access with a single
typed entry point, so client access needs no per-call ``cast``.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import TYPE_CHECKING, Annotated, TypeVar, cast

import httpx
import redis.asyncio as aioredis
from fastapi import Depends, Request
from sqlalchemy.ext.asyncio import AsyncEngine, AsyncSession, async_sessionmaker

if TYPE_CHECKING:
    from app.infrastructure.storage.s3_client import S3Client

T = TypeVar("T")


def require_not_none(value: T | None, name: str) -> T:
    """Narrow an optional container field, raising if a lifespan didn't set it."""
    if value is None:
        raise RuntimeError(f"Container field {name!r} is not initialized")
    return value


@dataclass
class Container:
    db_engine: AsyncEngine | None = None
    db_sessionmaker: async_sessionmaker[AsyncSession] | None = None
    redis: aioredis.Redis | None = None
    http_clients: dict[str, httpx.AsyncClient] = field(default_factory=dict)
    s3: S3Client | None = None
    s3_public: S3Client | None = None


def get_container(request: Request) -> Container:
    return cast(Container, request.app.state.container)


ContainerDep = Annotated[Container, Depends(get_container)]
