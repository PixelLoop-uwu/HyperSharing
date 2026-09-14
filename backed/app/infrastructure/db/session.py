"""Per-request ``AsyncSession`` dependency."""

from collections.abc import AsyncIterator
from typing import Annotated

from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from app.infrastructure.container import ContainerDep, require_not_none


async def get_db_session(container: ContainerDep) -> AsyncIterator[AsyncSession]:
    """Yield a request-scoped session, committing or rolling back as needed."""
    sessionmaker = require_not_none(container.db_sessionmaker, "db_sessionmaker")
    async with sessionmaker() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise


DbSessionDep = Annotated[AsyncSession, Depends(get_db_session)]
