"""Unit tests for MultipartUploadManager.complete idempotency."""

import pytest
from botocore.exceptions import ClientError

from app.core.exceptions import StorageError
from app.infrastructure.storage.multipart import CompletedPart, MultipartUploadManager


class FakeS3Client:
    """Stands in for the S3 client's multipart-related methods."""

    def __init__(self, *, object_exists: bool, content_length: int) -> None:
        self._object_exists = object_exists
        self._content_length = content_length
        self.completed: list[str] = []

    async def head_object(self, Bucket: str, Key: str) -> dict:
        if not self._object_exists:
            raise ClientError(
                {"Error": {"Code": "NoSuchKey", "Message": "not found"}},
                "HeadObject",
            )
        return {"ContentLength": self._content_length}

    async def list_parts(self, Bucket: str, Key: str, UploadId: str) -> dict:
        return {"Parts": [{"Size": self._content_length}]}

    async def complete_multipart_upload(
        self, Bucket: str, Key: str, UploadId: str, MultipartUpload: dict
    ) -> dict:
        self.completed.append(Key)
        return {"Key": Key}


def _manager(client: FakeS3Client) -> MultipartUploadManager:
    return MultipartUploadManager(client, "bucket")  # type: ignore[arg-type]


_PART = CompletedPart(part_number=1, etag='"etag"')


async def test_complete_skips_when_object_already_exists() -> None:
    client = FakeS3Client(object_exists=True, content_length=10)
    manager = _manager(client)

    result = await manager.complete(
        "objects/hash", "upload-id", [_PART], expected_size=10
    )

    assert result == {}
    assert client.completed == []


async def test_complete_assembles_when_object_missing() -> None:
    client = FakeS3Client(object_exists=False, content_length=10)
    manager = _manager(client)

    result = await manager.complete(
        "objects/hash", "upload-id", [_PART], expected_size=10
    )

    assert result == {"Key": "objects/hash"}
    assert client.completed == ["objects/hash"]


async def test_complete_rejects_existing_object_size_mismatch() -> None:
    client = FakeS3Client(object_exists=True, content_length=9)
    manager = _manager(client)

    with pytest.raises(StorageError):
        await manager.complete("objects/hash", "upload-id", [_PART], expected_size=10)
