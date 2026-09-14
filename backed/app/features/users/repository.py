"""User repository and the Redis-backed session store."""

from typing import Annotated

import redis.asyncio as aioredis
from fastapi import Depends
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.config import settings
from app.features.users.models import User
from app.features.users.schemas import SessionData
from app.infrastructure.cache.base_store import BaseRedisStore
from app.infrastructure.cache.redis_client import RedisDep
from app.infrastructure.db.repository import BaseRepository
from app.infrastructure.db.session import DbSessionDep


class UserRepository(BaseRepository[User]):
    def __init__(self, session: AsyncSession) -> None:
        super().__init__(session, User)

    async def get_by_email(self, email: str) -> User | None:
        stmt = select(User).where(User.email == email)
        return (await self._session.execute(stmt)).scalar_one_or_none()

    async def get_by_username(self, username: str) -> User | None:
        stmt = select(User).where(User.username == username)
        return (await self._session.execute(stmt)).scalar_one_or_none()


class SessionStore:
    """Stores opaque session tokens in Redis with a sliding TTL.

    Single session per token, no refresh token: ``touch`` re-``set``s the same
    payload so the TTL resets on each authenticated request.
    """

    def __init__(self, redis: aioredis.Redis, *, ttl: int) -> None:
        self._store = BaseRedisStore(redis, SessionData, ttl=ttl)

    @staticmethod
    def _key(token: str) -> str:
        return f"session:{token}"

    async def get(self, token: str) -> SessionData | None:
        return await self._store.get(self._key(token))

    async def set(self, token: str, session: SessionData) -> None:
        await self._store.set(self._key(token), session)

    async def touch(self, token: str, session: SessionData) -> None:
        """Slide the TTL forward without changing the payload."""
        await self._store.set(self._key(token), session)

    async def delete(self, token: str) -> None:
        await self._store.delete(self._key(token))


async def get_user_repository(session: DbSessionDep) -> UserRepository:
    return UserRepository(session)


async def get_session_store(redis: RedisDep) -> SessionStore:
    return SessionStore(redis, ttl=settings.session_ttl_seconds)


UserRepositoryDep = Annotated[UserRepository, Depends(get_user_repository)]
SessionStoreDep = Annotated[SessionStore, Depends(get_session_store)]
