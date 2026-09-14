"""create posts, files and post_items tables

Revision ID: c1d2e3f4a5b6
Revises:
Create Date: 2026-09-13 00:00:00.000000

"""

from collections.abc import Sequence

import sqlalchemy as sa

from alembic import op

revision: str = "c1d2e3f4a5b6"
down_revision: str | None = None
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.create_table(
        "files",
        sa.Column(
            "id",
            sa.Uuid(),
            server_default=sa.text("uuidv7()"),
            nullable=False,
        ),
        sa.Column("hash", sa.String(length=64), nullable=False),
        sa.Column("size", sa.BigInteger(), nullable=False),
        sa.Column("content_type", sa.String(length=255), nullable=True),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            server_default=sa.func.now(),
            nullable=False,
        ),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            server_default=sa.func.now(),
            nullable=False,
        ),
        sa.PrimaryKeyConstraint("id"),
        sa.UniqueConstraint("hash", name="uq_files_hash"),
    )

    op.create_table(
        "posts",
        sa.Column(
            "id",
            sa.Uuid(),
            server_default=sa.text("uuidv7()"),
            nullable=False,
        ),
        sa.Column("title", sa.String(length=255), nullable=True),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            server_default=sa.func.now(),
            nullable=False,
        ),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            server_default=sa.func.now(),
            nullable=False,
        ),
        sa.PrimaryKeyConstraint("id"),
    )

    op.create_table(
        "post_items",
        sa.Column(
            "id",
            sa.Uuid(),
            server_default=sa.text("uuidv7()"),
            nullable=False,
        ),
        sa.Column("post_id", sa.Uuid(), nullable=False),
        sa.Column("position", sa.Integer(), nullable=False),
        sa.Column("kind", sa.String(length=16), nullable=False),
        sa.Column("text", sa.Text(), nullable=True),
        sa.Column("file_id", sa.Uuid(), nullable=True),
        sa.ForeignKeyConstraint(["post_id"], ["posts.id"], ondelete="CASCADE"),
        sa.ForeignKeyConstraint(["file_id"], ["files.id"], ondelete="RESTRICT"),
        sa.PrimaryKeyConstraint("id"),
        sa.UniqueConstraint("post_id", "position", name="uq_post_items_post_position"),
        sa.CheckConstraint(
            "kind IN ('paragraph', 'file')",
            name="ck_post_items_kind",
        ),
        sa.CheckConstraint(
            "(kind = 'paragraph' AND text IS NOT NULL AND file_id IS NULL) OR "
            "(kind = 'file' AND file_id IS NOT NULL AND text IS NULL)",
            name="ck_post_items_kind_payload",
        ),
    )


def downgrade() -> None:
    op.drop_table("post_items")
    op.drop_table("posts")
    op.drop_table("files")
