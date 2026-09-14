"""Business logic for posts."""

from __future__ import annotations

import base64
import math
import uuid
from collections.abc import Sequence
from datetime import UTC, datetime, timedelta
from typing import Annotated

from botocore.exceptions import ClientError
from fastapi import Depends

from app.core.config import settings
from app.core.exceptions import (
    ForbiddenError,
    NotFoundError,
    StorageError,
    ValidationError,
)
from app.features.posts import schemas
from app.features.posts.limits import PostLimits
from app.features.posts.models import ItemKind, Post, PostItem
from app.features.posts.repository import (
    DraftStore,
    DraftStoreDep,
    PostCache,
    PostCacheDep,
    PostItemRepository,
    PostItemRepositoryDep,
    PostRepository,
    PostRepositoryDep,
)
from app.features.storage.keys import s3_object_key
from app.features.storage.repository import FileRepository, FileRepositoryDep
from app.infrastructure.storage.multipart import (
    CompletedPart,
    MultipartManagerDep,
    MultipartUploadManager,
)
from app.infrastructure.storage.s3_client import S3Client, S3PublicDep


def _sha256_hex_to_base64(hex_digest: str) -> str:
    """Convert a hex SHA256 digest to the base64 form S3 expects."""
    return base64.b64encode(bytes.fromhex(hex_digest)).decode()


class PostService:
    def __init__(
        self,
        post_repo: PostRepository,
        post_item_repo: PostItemRepository,
        file_repo: FileRepository,
        draft_store: DraftStore,
        post_cache: PostCache,
        multipart_manager: MultipartUploadManager,
        s3_public_client: S3Client,
    ) -> None:
        self._post_repo = post_repo
        self._post_item_repo = post_item_repo
        self._file_repo = file_repo
        self._draft_store = draft_store
        self._post_cache = post_cache
        self._multipart_manager = multipart_manager
        self._s3_public_client = s3_public_client

    async def get_full(self, post_id: uuid.UUID) -> schemas.PostFullRead:
        """Return a post with its ordered content blocks and file URLs.

        Served from Redis when cached; posts are immutable, so a cached
        snapshot is always valid (the TTL accounts for the presigned URLs).
        """
        cached = await self._post_cache.get(post_id)
        if cached is not None:
            return cached

        post = await self._post_repo.get_by_id(post_id)
        if post is None:
            raise NotFoundError(f"post {post_id} not found")

        items: list[schemas.PostItemOut] = []
        for item, file in await self._post_item_repo.list_for_post(post_id):
            if item.kind == ItemKind.PARAGRAPH.value:
                items.append(
                    schemas.ParagraphItemOut(kind="paragraph", text=item.text or "")
                )
                continue

            if file is None:
                # Unreachable under the DB constraints (kind='file' requires a
                # file_id, FK is RESTRICT), kept defensive against bad data.
                raise StorageError(f"file reference missing for item {item.id}")
            url = await self._presign_get_url(file.hash)
            items.append(
                schemas.FileItemOut(
                    kind="file",
                    hash=file.hash,
                    size=file.size,
                    content_type=file.content_type,
                    url=url,
                )
            )

        full = schemas.PostFullRead(
            id=post.id,
            title=post.title,
            user_id=post.user_id,
            items=items,
            created_at=post.created_at,
        )
        await self._post_cache.set(full)
        return full

    async def list_posts(self, *, offset: int = 0, limit: int = 100) -> Sequence[Post]:
        return await self._post_repo.list(offset=offset, limit=limit)

    # --- Draft / publish ---

    def _validate_draft_limits(
        self,
        data: schemas.PostDraftCreate,
        file_meta: dict[str, tuple[int, str | None, list[str]]],
        limits: PostLimits,
    ) -> None:
        """Enforce post-size limits before any multipart upload is opened.

        Files are counted by unique hash (``file_meta``), so the same file
        embedded multiple times counts once toward the file/total limits.
        """
        file_count = len(file_meta)
        if file_count > limits.max_files:
            raise ValidationError(
                f"too many files: {file_count} (max {limits.max_files})"
            )

        total_size = sum(size for size, _, _ in file_meta.values())
        if total_size > limits.max_total_size:
            raise ValidationError(
                f"total file size {total_size} exceeds {limits.max_total_size} bytes"
            )

        for file_hash, (size, _, _) in file_meta.items():
            if size > limits.max_file_size:
                raise ValidationError(
                    f"file {file_hash} is {size} bytes, max {limits.max_file_size}"
                )

        paragraph_count = sum(
            1 for item in data.items if isinstance(item, schemas.ParagraphItemIn)
        )
        if paragraph_count > limits.max_paragraphs:
            raise ValidationError(
                f"too many text blocks: {paragraph_count} "
                f"(max {limits.max_paragraphs})"
            )

    async def create_draft(
        self,
        data: schemas.PostDraftCreate,
        *,
        limits: PostLimits,
        user_id: uuid.UUID | None,
    ) -> schemas.PostDraftRead:
        draft_id = uuid.uuid4()
        part_size = settings.s3_multipart_part_size

        # Deduplicate files by hash, preserving first-occurrence order.
        file_meta: dict[str, tuple[int, str | None, list[str]]] = {}
        for item in data.items:
            if isinstance(item, schemas.FileItemIn) and item.hash not in file_meta:
                file_meta[item.hash] = (
                    item.size,
                    item.content_type,
                    item.part_checksums,
                )

        self._validate_draft_limits(data, file_meta, limits)

        draft_files: dict[str, schemas.DraftFile] = {}
        uploads: list[schemas.FileUploadOut] = []
        for file_hash, (size, content_type, part_checksums) in file_meta.items():
            key = s3_object_key(file_hash)
            part_count = math.ceil(size / part_size)
            if len(part_checksums) != part_count:
                raise ValidationError(
                    f"expected {part_count} part checksums for {file_hash}, "
                    f"got {len(part_checksums)}"
                )
            upload_id = await self._multipart_manager.create(
                key, content_type=content_type
            )

            parts: list[schemas.PartPresignOut] = []
            for part_number in range(1, part_count + 1):
                checksum_b64 = _sha256_hex_to_base64(part_checksums[part_number - 1])
                url = await self._multipart_manager.presign_part(
                    key, upload_id, part_number, checksum_sha256=checksum_b64
                )
                parts.append(
                    schemas.PartPresignOut(
                        part_number=part_number,
                        url=url,
                        checksum_sha256=checksum_b64,
                    )
                )

            draft_files[file_hash] = schemas.DraftFile(
                hash=file_hash,
                size=size,
                content_type=content_type,
                upload_id=upload_id,
                part_size=part_size,
                part_count=part_count,
            )
            uploads.append(
                schemas.FileUploadOut(hash=file_hash, part_size=part_size, parts=parts)
            )

        items: list[schemas.DraftItem] = []
        for item in data.items:
            if isinstance(item, schemas.ParagraphItemIn):
                items.append(schemas.DraftParagraph(kind="paragraph", text=item.text))
            else:
                items.append(schemas.DraftFileRef(kind="file", hash=item.hash))

        draft = schemas.Draft(
            draft_id=draft_id,
            title=data.title,
            items=items,
            files=draft_files,
            user_id=user_id,
        )
        await self._draft_store.set(draft)

        expires_at = datetime.now(UTC) + timedelta(seconds=settings.draft_ttl_seconds)
        return schemas.PostDraftRead(
            draft_id=draft_id, expires_at=expires_at, files=uploads
        )

    async def publish(
        self, data: schemas.PostPublishIn, *, user_id: uuid.UUID | None
    ) -> schemas.PostFullRead:
        # Idempotency: ``posts.draft_id`` maps a draft 1:1 to its post. If a
        # previous (or concurrent) publish already created the post, replay it
        # instead of duplicating it or failing on the now-missing draft.
        existing = await self._post_repo.get_by_draft_id(data.draft_id)
        if existing is not None:
            return await self.get_full(existing.id)

        draft = await self._draft_store.get(data.draft_id)
        if draft is None:
            raise NotFoundError(f"draft {data.draft_id} not found")

        # A draft is publishable only by its creator: the session's user must
        # match the draft owner (None for anonymous drafts).
        if draft.user_id != user_id:
            raise ForbiddenError("draft belongs to another user")

        request_hashes = {f.hash for f in data.files}
        draft_hashes = set(draft.files)
        if request_hashes != draft_hashes:
            raise ValidationError("publish files do not match the draft")

        parts_by_hash: dict[str, list[CompletedPart]] = {}
        for file_complete in data.files:
            draft_file = draft.files[file_complete.hash]
            parts = sorted(file_complete.parts, key=lambda p: p.part_number)
            numbers = [p.part_number for p in parts]
            if len(parts) != draft_file.part_count or numbers != list(
                range(1, draft_file.part_count + 1)
            ):
                raise ValidationError(
                    f"expected parts 1..{draft_file.part_count} "
                    f"for {file_complete.hash}"
                )
            parts_by_hash[file_complete.hash] = [
                CompletedPart(part_number=p.part_number, etag=p.etag) for p in parts
            ]

        try:
            for file_hash, draft_file in draft.files.items():
                await self._multipart_manager.complete(
                    s3_object_key(file_hash),
                    draft_file.upload_id,
                    parts_by_hash[file_hash],
                    expected_size=draft_file.size,
                )
        except ClientError as exc:
            raise StorageError(f"multipart completion failed: {exc}") from exc

        candidate = Post(
            title=draft.title, user_id=draft.user_id, draft_id=draft.draft_id
        )
        post, created = await self._post_repo.create_or_get_by_draft_id(candidate)
        if not created:
            return await self.get_full(post.id)

        for position, item in enumerate(draft.items):
            if isinstance(item, schemas.DraftParagraph):
                await self._post_item_repo.create(
                    PostItem(
                        post_id=post.id,
                        position=position,
                        kind=ItemKind.PARAGRAPH.value,
                        text=item.text,
                    )
                )
            else:
                draft_file = draft.files[item.hash]
                stored_file = await self._file_repo.get_or_create(
                    hash=item.hash,
                    size=draft_file.size,
                    content_type=draft_file.content_type,
                )
                await self._post_item_repo.create(
                    PostItem(
                        post_id=post.id,
                        position=position,
                        kind=ItemKind.FILE.value,
                        file_id=stored_file.id,
                    )
                )

        await self._draft_store.delete(draft.draft_id)

        full = await self._build_full_post(post, draft)
        await self._post_cache.set(full)
        return full

    async def _build_full_post(
        self, post: Post, draft: schemas.Draft
    ) -> schemas.PostFullRead:
        items: list[schemas.PostItemOut] = []
        for item in draft.items:
            if isinstance(item, schemas.DraftParagraph):
                items.append(schemas.ParagraphItemOut(kind="paragraph", text=item.text))
            else:
                draft_file = draft.files[item.hash]
                url = await self._presign_get_url(item.hash)
                items.append(
                    schemas.FileItemOut(
                        kind="file",
                        hash=draft_file.hash,
                        size=draft_file.size,
                        content_type=draft_file.content_type,
                        url=url,
                    )
                )
        return schemas.PostFullRead(
            id=post.id,
            title=post.title,
            user_id=post.user_id,
            items=items,
            created_at=post.created_at,
        )

    async def _presign_get_url(self, file_hash: str) -> str:
        return await self._s3_public_client.generate_presigned_url(
            "get_object",
            Params={"Bucket": settings.s3_bucket, "Key": s3_object_key(file_hash)},
            ExpiresIn=settings.s3_presign_ttl_seconds,
        )


async def get_post_service(
    post_repo: PostRepositoryDep,
    post_item_repo: PostItemRepositoryDep,
    file_repo: FileRepositoryDep,
    draft_store: DraftStoreDep,
    post_cache: PostCacheDep,
    multipart_manager: MultipartManagerDep,
    s3_public_client: S3PublicDep,
) -> PostService:
    return PostService(
        post_repo=post_repo,
        post_item_repo=post_item_repo,
        file_repo=file_repo,
        draft_store=draft_store,
        post_cache=post_cache,
        multipart_manager=multipart_manager,
        s3_public_client=s3_public_client,
    )


PostServiceDep = Annotated[PostService, Depends(get_post_service)]
