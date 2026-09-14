"""Generic CRUD repository.

Only the truly shared pattern is generalized (get/list/count/create/delete).
Complex filtering, joins and eager loading belong to feature repositories that
subclass this base.
"""

from collections.abc import Sequence
from typing import Any, Generic, TypeVar

from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.infrastructure.db.base import Base

ModelType = TypeVar("ModelType", bound=Base)


class BaseRepository(Generic[ModelType]):
    def __init__(self, session: AsyncSession, model: type[ModelType]) -> None:
        self._session = session
        self._model = model

    async def get_by_id(self, obj_id: Any) -> ModelType | None:
        return await self._session.get(self._model, obj_id)

    async def list(self, *, offset: int = 0, limit: int = 100) -> Sequence[ModelType]:
        """Ordered by primary key for deterministic pagination."""
        stmt = select(self._model).order_by(self._model.id).offset(offset).limit(limit)
        result = await self._session.execute(stmt)
        return result.scalars().all()

    async def count(self) -> int:
        stmt = select(func.count()).select_from(self._model)
        return int((await self._session.scalar(stmt)) or 0)

    async def create(self, obj: ModelType) -> ModelType:
        """Add and flush so server-generated ids are populated."""
        self._session.add(obj)
        await self._session.flush()
        return obj

    async def delete(self, obj: ModelType) -> None:
        await self._session.delete(obj)
