"""Named httpx.AsyncClient sessions, their lifespan and dependency.

A single default client is created at startup. Features needing a dedicated
client (different base URL/timeouts) register another one in the lifespan
below; ``get_http_client(name)`` resolves it by name.
"""

from collections.abc import AsyncIterator, Awaitable, Callable
from contextlib import asynccontextmanager
from typing import Annotated

import httpx
from fastapi import Depends

from app.core.config import settings
from app.infrastructure.container import Container, ContainerDep


@asynccontextmanager
async def http_lifespan(container: Container) -> AsyncIterator[None]:
    clients: dict[str, httpx.AsyncClient] = {
        settings.http_default_client_name: httpx.AsyncClient(
            timeout=httpx.Timeout(settings.http_timeout_seconds)
        ),
    }
    container.http_clients = clients
    try:
        yield
    finally:
        for client in clients.values():
            await client.aclose()


def get_http_client(
    name: str = settings.http_default_client_name,
) -> Callable[[Container], Awaitable[httpx.AsyncClient]]:
    async def dependency(container: ContainerDep) -> httpx.AsyncClient:
        return container.http_clients[name]

    return dependency


HttpClientDep = Annotated[httpx.AsyncClient, Depends(get_http_client())]
