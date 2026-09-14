"""Pydantic schemas for the post API and the internal draft state."""

import uuid
from datetime import datetime
from typing import Annotated, Literal

from pydantic import BaseModel, ConfigDict, Field

from app.core.config import settings


class PostRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    title: str | None
    user_id: uuid.UUID | None
    created_at: datetime
    updated_at: datetime


# --- Draft request ---


class ParagraphItemIn(BaseModel):
    kind: Literal["paragraph"]
    text: str = Field(min_length=1)


class FileItemIn(BaseModel):
    kind: Literal["file"]
    hash: str = Field(pattern=settings.sha256_hex_pattern)
    size: int = Field(gt=0)
    content_type: str | None = None
    # SHA256 (hex) of each part's bytes, in part order. The client splits the
    # file into ``part_size`` chunks and hashes each chunk; the number of
    # entries must equal the number of parts the server derives from ``size``.
    part_checksums: list[Annotated[str, Field(pattern=settings.sha256_hex_pattern)]] = (
        Field(min_length=1)
    )


ItemIn = Annotated[ParagraphItemIn | FileItemIn, Field(discriminator="kind")]


class PostDraftCreate(BaseModel):
    title: str | None = Field(default=None, max_length=255)
    items: list[ItemIn] = Field(min_length=1)


class PartPresignOut(BaseModel):
    part_number: int
    url: str
    # Base64 SHA256 of this part; the client sends it back as the
    # ``x-amz-checksum-sha256`` header (with ``x-amz-sdk-checksum-algorithm:
    # SHA256``) when uploading, matching the value baked into the presigned URL.
    checksum_sha256: str


class FileUploadOut(BaseModel):
    hash: str
    part_size: int
    parts: list[PartPresignOut]


class PostDraftRead(BaseModel):
    draft_id: uuid.UUID
    expires_at: datetime
    files: list[FileUploadOut]


# --- Publish ---


class PartEtagIn(BaseModel):
    part_number: int = Field(ge=1)
    etag: str


class FileCompleteIn(BaseModel):
    hash: str = Field(pattern=settings.sha256_hex_pattern)
    parts: list[PartEtagIn] = Field(min_length=1)


class PostPublishIn(BaseModel):
    draft_id: uuid.UUID
    files: list[FileCompleteIn]


# --- Full post with items ---


class ParagraphItemOut(BaseModel):
    kind: Literal["paragraph"]
    text: str


class FileItemOut(BaseModel):
    kind: Literal["file"]
    hash: str
    size: int
    content_type: str | None
    url: str


PostItemOut = Annotated[ParagraphItemOut | FileItemOut, Field(discriminator="kind")]


class PostFullRead(BaseModel):
    id: uuid.UUID
    title: str | None
    user_id: uuid.UUID | None
    items: list[PostItemOut]
    created_at: datetime


# --- Internal draft state (serialized to Redis) ---


class DraftParagraph(BaseModel):
    kind: Literal["paragraph"]
    text: str


class DraftFileRef(BaseModel):
    kind: Literal["file"]
    hash: str


DraftItem = Annotated[DraftParagraph | DraftFileRef, Field(discriminator="kind")]


class DraftFile(BaseModel):
    """Multipart-upload state for one unique file in the draft."""

    hash: str
    size: int
    content_type: str | None
    upload_id: str
    part_size: int
    part_count: int


class Draft(BaseModel):
    draft_id: uuid.UUID
    title: str | None
    user_id: uuid.UUID | None = None
    items: list[DraftItem]
    files: dict[str, DraftFile]
