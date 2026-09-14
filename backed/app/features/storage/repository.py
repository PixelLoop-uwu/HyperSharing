"""File repository — deduplication by content hash."""

from typing import Annotated

from fastapi import Depends
from sqlalchemy import select
from sqlalchemy.dialects.postgresql import insert as pg_insert
from sqlalchemy.ext.asyncio import AsyncSession

from app.features.storage.models import File
from app.infrastructure.db.session import DbSessionDep


class FileRepository:
    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    async def get_or_create(
        self, *, hash: str, size: int, content_type: str | None
    ) -> File:
        """Return the row for ``hash``, inserting it if absent.

        The insert uses ``ON CONFLICT DO NOTHING`` so concurrent publishes of the
        same content race safely and still share a single ``files`` row.
        """
        insert_stmt = (
            pg_insert(File)
            .values(hash=hash, size=size, content_type=content_type)
            .on_conflict_do_nothing(index_elements=["hash"])
        )
        await self._session.execute(insert_stmt)
        select_stmt = select(File).where(File.hash == hash)
        return (await self._session.execute(select_stmt)).scalar_one()


async def get_file_repository(session: DbSessionDep) -> FileRepository:
    return FileRepository(session)


FileRepositoryDep = Annotated[FileRepository, Depends(get_file_repository)]
