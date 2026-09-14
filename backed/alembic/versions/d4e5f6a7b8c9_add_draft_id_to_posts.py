"""add draft_id to posts (idempotency key for publish)

Revision ID: d4e5f6a7b8c9
Revises: b7c8d9e0f1a2
Create Date: 2026-09-15 00:00:00.000000

"""

from collections.abc import Sequence

import sqlalchemy as sa

from alembic import op

revision: str = "d4e5f6a7b8c9"
down_revision: str | None = "b7c8d9e0f1a2"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    # NULL (not NOT NULL) so rows predating this column migrate cleanly; every
    # new publish sets it, and UNIQUE enforces one post per draft.
    op.add_column("posts", sa.Column("draft_id", sa.Uuid(), nullable=True))
    op.create_unique_constraint("uq_posts_draft_id", "posts", ["draft_id"])


def downgrade() -> None:
    op.drop_constraint("uq_posts_draft_id", "posts", type_="unique")
    op.drop_column("posts", "draft_id")
