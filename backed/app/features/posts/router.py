"""Post endpoints."""

import uuid
from typing import Annotated

from fastapi import APIRouter, Query, status

from app.features.posts import schemas
from app.features.posts.limits import PostLimitsDep
from app.features.posts.service import PostServiceDep
from app.features.users.deps import CurrentUserDep

router = APIRouter(prefix="/posts", tags=["posts"])


@router.post(
    "/draft",
    response_model=schemas.PostDraftRead,
    status_code=status.HTTP_201_CREATED,
)
async def create_draft(
    data: schemas.PostDraftCreate,
    service: PostServiceDep,
    limits: PostLimitsDep,
    current_user: CurrentUserDep,
) -> schemas.PostDraftRead:
    return await service.create_draft(
        data, limits=limits, user_id=current_user.id if current_user else None
    )


@router.post(
    "/publish",
    response_model=schemas.PostFullRead,
    status_code=status.HTTP_201_CREATED,
)
async def publish_post(
    data: schemas.PostPublishIn,
    service: PostServiceDep,
    current_user: CurrentUserDep,
) -> schemas.PostFullRead:
    return await service.publish(
        data, user_id=current_user.id if current_user else None
    )


@router.get("/{post_id}", response_model=schemas.PostFullRead)
async def get_post(
    post_id: uuid.UUID,
    service: PostServiceDep,
) -> schemas.PostFullRead:
    return await service.get_full(post_id)


@router.get("", response_model=list[schemas.PostRead])
async def list_posts(
    service: PostServiceDep,
    offset: Annotated[int, Query(ge=0)] = 0,
    limit: Annotated[int, Query(ge=1, le=100)] = 100,
) -> list[schemas.PostRead]:
    posts = await service.list_posts(offset=offset, limit=limit)
    return [schemas.PostRead.model_validate(p) for p in posts]
