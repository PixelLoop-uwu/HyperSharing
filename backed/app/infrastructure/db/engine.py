"""Async engine / sessionmaker creation and lifespan.

The engine and sessionmaker are created once at startup and stored on the
container. ``AsyncSession`` itself is *never* stored on the container — it is
created per-request by ``app.infrastructure.db.session``.
"""

from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from sqlalchemy.ext.asyncio import (
    AsyncEngine,
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)

from app.core.config import settings
from app.infrastructure.container import Container


def create_engine() -> AsyncEngine:
    return create_async_engine(
        settings.database_url,
        echo=settings.debug,
        pool_pre_ping=True,
    )


def create_sessionmaker(engine: AsyncEngine) -> async_sessionmaker[AsyncSession]:
    return async_sessionmaker(
        engine,
        expire_on_commit=False,
        autoflush=False,
    )


@asynccontextmanager
async def db_lifespan(container: Container) -> AsyncIterator[None]:
    engine = create_engine()
    container.db_engine = engine
    container.db_sessionmaker = create_sessionmaker(engine)
    try:
        yield
    finally:
        await engine.dispose()
