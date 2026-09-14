"""add user_id to posts

Revision ID: b7c8d9e0f1a2
Revises: a1b2c3d4e5f6
Create Date: 2026-09-14 00:00:00.000000

"""

from collections.abc import Sequence

import sqlalchemy as sa

from alembic import op

revision: str = "b7c8d9e0f1a2"
down_revision: str | None = "a1b2c3d4e5f6"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.add_column("posts", sa.Column("user_id", sa.Uuid(), nullable=True))
    op.create_foreign_key(
        "fk_posts_user_id_users",
        "posts",
        "users",
        ["user_id"],
        ["id"],
        ondelete="SET NULL",
    )


def downgrade() -> None:
    op.drop_constraint("fk_posts_user_id_users", "posts", type_="foreignkey")
    op.drop_column("posts", "user_id")
