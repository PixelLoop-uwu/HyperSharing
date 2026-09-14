"""Composition of every infrastructure lifespan.

Each module owns its own lifespan (created next to its code). Here they are
entered through a single ``AsyncExitStack`` so clients are closed in reverse
order even if startup fails partway. A single typed ``Container`` is populated
by each lifespan and exposed on ``app.state.container``.
"""

from collections.abc import AsyncIterator
from contextlib import AsyncExitStack, asynccontextmanager

from fastapi import FastAPI

from app.infrastructure.cache.redis_client import redis_lifespan
from app.infrastructure.container import Container
from app.infrastructure.db.engine import db_lifespan
from app.infrastructure.http.clients import http_lifespan
from app.infrastructure.storage.s3_client import s3_lifespan


@asynccontextmanager
async def populate_container() -> AsyncIterator[Container]:
    """Enter every infrastructure lifespan into a single typed ``Container``.

    Reused by both the API (``lifespan`` below) and the arq worker, so the two
    processes build their clients the same way and tear them down in reverse
    order even if startup fails partway.
    """
    container = Container()
    async with AsyncExitStack() as stack:
        await stack.enter_async_context(redis_lifespan(container))
        await stack.enter_async_context(db_lifespan(container))
        await stack.enter_async_context(http_lifespan(container))
        await stack.enter_async_context(s3_lifespan(container))
        yield container


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    async with populate_container() as container:
        app.state.container = container
        yield
