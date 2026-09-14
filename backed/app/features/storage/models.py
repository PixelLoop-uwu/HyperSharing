"""Content-addressed file registry.

``files`` is a deduplication table: one row per unique content hash. The actual
bytes live in S3 at a key derived from the hash, so identical files are stored
once and shared across posts.
"""

from sqlalchemy import BigInteger, String
from sqlalchemy.orm import Mapped, mapped_column

from app.infrastructure.db.base import Base, TimestampMixin


class File(Base, TimestampMixin):
    __tablename__ = "files"

    hash: Mapped[str] = mapped_column(String(64), unique=True, nullable=False)
    size: Mapped[int] = mapped_column(BigInteger, nullable=False)
    content_type: Mapped[str | None] = mapped_column(String(255))
