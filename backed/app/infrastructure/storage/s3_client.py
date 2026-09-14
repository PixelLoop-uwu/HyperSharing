"""S3 (MinIO) client, its lifespan and dependency.

The client is an ``aiobotocore`` async S3 client (``aiobotocore.client.S3``).
We type it with a minimal structural ``Protocol`` so the code stays mypy-clean
without forcing the ``types-aiobotocore-s3`` stub dependency — swap that in if
precise per-method signatures are wanted.

Two clients may exist: an internal one (``s3_endpoint_url``) for all backend
operations, and an optional public one (``s3_public_endpoint_url``) used only
to presign URLs for end users. The public client exists only when a distinct
endpoint is configured, because presigned URLs are signed over the host and
cannot be rewritten afterwards.
"""

from collections.abc import AsyncIterator
from contextlib import AsyncExitStack, asynccontextmanager
from typing import Annotated, Any, Protocol

import aioboto3
from fastapi import Depends

from app.core.config import settings
from app.infrastructure.container import Container, ContainerDep, require_not_none


class S3Client(Protocol):
    """The subset of the S3 client API this project relies on."""

    async def put_object(self, *args: Any, **kwargs: Any) -> dict[str, Any]: ...
    async def get_object(self, *args: Any, **kwargs: Any) -> dict[str, Any]: ...
    async def head_object(self, *args: Any, **kwargs: Any) -> dict[str, Any]: ...
    async def delete_object(self, *args: Any, **kwargs: Any) -> dict[str, Any]: ...
    async def head_bucket(self, *args: Any, **kwargs: Any) -> dict[str, Any]: ...
    async def create_multipart_upload(
        self, *args: Any, **kwargs: Any
    ) -> dict[str, Any]: ...
    async def upload_part(self, *args: Any, **kwargs: Any) -> dict[str, Any]: ...
    async def complete_multipart_upload(
        self, *args: Any, **kwargs: Any
    ) -> dict[str, Any]: ...
    async def abort_multipart_upload(
        self, *args: Any, **kwargs: Any
    ) -> dict[str, Any]: ...
    async def list_parts(self, *args: Any, **kwargs: Any) -> dict[str, Any]: ...
    async def list_multipart_uploads(
        self, *args: Any, **kwargs: Any
    ) -> dict[str, Any]: ...
    async def generate_presigned_url(self, *args: Any, **kwargs: Any) -> str: ...


async def get_s3(container: ContainerDep) -> S3Client:
    return require_not_none(container.s3, "s3")


S3Dep = Annotated[S3Client, Depends(get_s3)]


async def get_s3_public(container: ContainerDep) -> S3Client:
    """Client for presigning URLs; falls back to the internal one if unset."""
    return require_not_none(container.s3_public or container.s3, "s3")


S3PublicDep = Annotated[S3Client, Depends(get_s3_public)]


def _make_client(session: Any, endpoint_url: str) -> Any:
    return session.client(
        "s3",
        endpoint_url=endpoint_url,
        aws_access_key_id=settings.s3_access_key_id,
        aws_secret_access_key=settings.s3_secret_access_key,
        region_name=settings.s3_region,
    )


@asynccontextmanager
async def s3_lifespan(container: Container) -> AsyncIterator[None]:
    session = aioboto3.Session()
    async with AsyncExitStack() as stack:
        container.s3 = await stack.enter_async_context(
            _make_client(session, settings.s3_endpoint_url)
        )
        if (
            settings.s3_public_endpoint_url
            and settings.s3_public_endpoint_url != settings.s3_endpoint_url
        ):
            container.s3_public = await stack.enter_async_context(
                _make_client(session, settings.s3_public_endpoint_url)
            )
        yield
