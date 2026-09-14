"""Post repositories and the Redis-backed draft store."""

import uuid
from collections.abc import Sequence
from typing import Annotated

import redis.asyncio as aioredis
from fastapi import Depends
from sqlalchemy import select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.config import settings
from app.features.posts.models import Post, PostItem
from app.features.posts.schemas import Draft, PostFullRead
from app.features.storage.models import File
from app.infrastructure.cache.base_store import BaseRedisStore
from app.infrastructure.cache.redis_client import RedisDep
from app.infrastructure.db.repository import BaseRepository
from app.infrastructure.db.session import DbSessionDep


class PostRepository(BaseRepository[Post]):
    def __init__(self, session: AsyncSession) -> None:
        super().__init__(session, Post)

    async def get_by_draft_id(self, draft_id: uuid.UUID) -> Post | None:
        stmt = select(Post).where(Post.draft_id == draft_id)
        result = await self._session.execute(stmt)
        return result.scalar_one_or_none()

    async def create_or_get_by_draft_id(self, post: Post) -> tuple[Post, bool]:
        """Insert ``post``; if its ``draft_id`` is already published, return the
        existing post instead of raising. Returns ``(post, created)``.

        The insert runs in a SAVEPOINT so a duplicate-key error rolls back only
        this statement, leaving the request transaction usable for the replay
        read.
        """
        try:
            async with self._session.begin_nested():
                self._session.add(post)
                await self._session.flush()
            return post, True
        except IntegrityError:
            if post.draft_id is None:
                raise
            existing = await self.get_by_draft_id(post.draft_id)
            if existing is None:
                raise
            return existing, False


class PostItemRepository(BaseRepository[PostItem]):
    def __init__(self, session: AsyncSession) -> None:
        super().__init__(session, PostItem)

    async def list_for_post(
        self, post_id: uuid.UUID
    ) -> Sequence[tuple[PostItem, File | None]]:
        """Return a post's content blocks in order, joined with their file.

        A ``LEFT OUTER JOIN`` is used because paragraphs have no file; the DB
        check constraint and the ``file_id`` FK guarantee a file item always
        resolves to a row in ``files``.
        """
        stmt = (
            select(PostItem, File)
            .outerjoin(File, PostItem.file_id == File.id)
            .where(PostItem.post_id == post_id)
            .order_by(PostItem.position)
        )
        result = await self._session.execute(stmt)
        return [(row[0], row[1]) for row in result.all()]


class DraftStore:
    """Persists drafts in Redis with a TTL; expired drafts are dropped."""

    def __init__(self, redis: aioredis.Redis, *, ttl: int) -> None:
        self._redis_store = BaseRedisStore(redis, Draft, ttl=ttl)

    @staticmethod
    def _key(draft_id: uuid.UUID) -> str:
        return f"post:draft:{draft_id}"

    async def get(self, draft_id: uuid.UUID) -> Draft | None:
        return await self._redis_store.get(self._key(draft_id))

    async def set(self, draft: Draft) -> None:
        await self._redis_store.set(self._key(draft.draft_id), draft)

    async def delete(self, draft_id: uuid.UUID) -> None:
        await self._redis_store.delete(self._key(draft_id))


class PostCache:
    """Caches published posts so GET doesn't re-read the DB on every request.

    Posts are immutable after publish, so invalidation is just TTL expiry. The
    payload includes presigned URLs, so ``ttl`` must stay below the presign TTL
    (``settings.post_cache_ttl_seconds``).
    """

    def __init__(self, redis: aioredis.Redis, *, ttl: int) -> None:
        self._redis_store = BaseRedisStore(redis, PostFullRead, ttl=ttl)

    @staticmethod
    def _key(post_id: uuid.UUID) -> str:
        return f"post:full:{post_id}"

    async def get(self, post_id: uuid.UUID) -> PostFullRead | None:
        return await self._redis_store.get(self._key(post_id))

    async def set(self, post: PostFullRead) -> None:
        await self._redis_store.set(self._key(post.id), post)


async def get_post_cache(redis: RedisDep) -> PostCache:
    return PostCache(redis, ttl=settings.post_cache_ttl_seconds)


async def get_post_repository(session: DbSessionDep) -> PostRepository:
    return PostRepository(session)


async def get_post_item_repository(session: DbSessionDep) -> PostItemRepository:
    return PostItemRepository(session)


async def get_draft_store(redis: RedisDep) -> DraftStore:
    return DraftStore(redis, ttl=settings.draft_ttl_seconds)


PostRepositoryDep = Annotated[PostRepository, Depends(get_post_repository)]
PostItemRepositoryDep = Annotated[PostItemRepository, Depends(get_post_item_repository)]
DraftStoreDep = Annotated[DraftStore, Depends(get_draft_store)]
PostCacheDep = Annotated[PostCache, Depends(get_post_cache)]
