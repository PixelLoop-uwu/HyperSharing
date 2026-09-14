"""SQLAlchemy models for posts."""

import uuid
from enum import StrEnum

from sqlalchemy import (
    CheckConstraint,
    ForeignKey,
    Integer,
    String,
    Text,
    UniqueConstraint,
)
from sqlalchemy.orm import Mapped, mapped_column

from app.infrastructure.db.base import Base, TimestampMixin


class ItemKind(StrEnum):
    """Kinds of content blocks within a post."""

    PARAGRAPH = "paragraph"
    FILE = "file"


class Post(Base, TimestampMixin):
    __tablename__ = "posts"

    title: Mapped[str | None] = mapped_column(String(255))
    # Owner of the post; NULL for anonymous posts. SET NULL keeps posts
    # readable if the owning account is later deleted.
    user_id: Mapped[uuid.UUID | None] = mapped_column(
        ForeignKey("users.id", ondelete="SET NULL")
    )
    # Idempotency key: the draft this post was published from. UNIQUE so a
    # retried /publish of the same draft can never create a duplicate post.
    # NULL only for rows predating this column; every new publish sets it.
    draft_id: Mapped[uuid.UUID | None] = mapped_column(unique=True, nullable=True)


class PostItem(Base):
    """An ordered content block: either a text paragraph or a file reference."""

    __tablename__ = "post_items"
    __table_args__ = (
        UniqueConstraint("post_id", "position", name="uq_post_items_post_position"),
        CheckConstraint("kind IN ('paragraph', 'file')", name="ck_post_items_kind"),
        CheckConstraint(
            "(kind = 'paragraph' AND text IS NOT NULL AND file_id IS NULL) OR "
            "(kind = 'file' AND file_id IS NOT NULL AND text IS NULL)",
            name="ck_post_items_kind_payload",
        ),
    )

    post_id: Mapped[uuid.UUID] = mapped_column(
        ForeignKey("posts.id", ondelete="CASCADE"), nullable=False
    )
    position: Mapped[int] = mapped_column(Integer, nullable=False)
    kind: Mapped[str] = mapped_column(String(16), nullable=False)
    text: Mapped[str | None] = mapped_column(Text)
    file_id: Mapped[uuid.UUID | None] = mapped_column(
        ForeignKey("files.id", ondelete="RESTRICT")
    )
