"""Simple put/get/delete/exists/presign S3 store (no chunking)."""

from typing import cast

from botocore.exceptions import ClientError

from app.core.config import settings
from app.infrastructure.storage.s3_client import S3Client


class BaseS3Store:
    def __init__(
        self,
        s3_client: S3Client,
        bucket: str,
        *,
        presign_client: S3Client | None = None,
    ) -> None:
        self._s3_client = s3_client
        self._bucket = bucket
        self._presign_client = presign_client or s3_client

    async def put(
        self, key: str, data: bytes, *, content_type: str | None = None
    ) -> None:
        kwargs: dict[str, str] = {}
        if content_type is not None:
            kwargs["ContentType"] = content_type
        await self._s3_client.put_object(
            Bucket=self._bucket, Key=key, Body=data, **kwargs
        )

    async def get(self, key: str) -> bytes | None:
        try:
            response = await self._s3_client.get_object(Bucket=self._bucket, Key=key)
        except ClientError as exc:
            if _is_no_such_key(exc):
                return None
            raise
        body = response["Body"]
        return cast(bytes, await body.read())

    async def delete(self, key: str) -> None:
        await self._s3_client.delete_object(Bucket=self._bucket, Key=key)

    async def exists(self, key: str) -> bool:
        try:
            await self._s3_client.head_object(Bucket=self._bucket, Key=key)
        except ClientError as exc:
            if _is_no_such_key(exc):
                return False
            raise
        return True

    async def presigned_get_url(
        self, key: str, *, expires_in: int = settings.s3_presign_ttl_seconds
    ) -> str:
        return await self._presign_client.generate_presigned_url(
            "get_object",
            Params={"Bucket": self._bucket, "Key": key},
            ExpiresIn=expires_in,
        )


def _is_no_such_key(exc: ClientError) -> bool:
    return exc.response["Error"]["Code"] in {"NoSuchKey", "404", "NotFound"}
