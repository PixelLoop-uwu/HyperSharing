"""Unit tests for cleanup_abandoned_uploads (the arq cleanup job)."""

from datetime import UTC, datetime, timedelta
from types import SimpleNamespace

from app.features.posts.jobs import cleanup_abandoned_uploads


def _upload(key: str, upload_id: str, initiated: datetime) -> dict:
    return {"Key": key, "UploadId": upload_id, "Initiated": initiated}


class FakeS3:
    """Stands in for the S3 client: serves scripted pages and records aborts."""

    def __init__(self, pages: list[list[dict]]) -> None:
        self._pages = pages
        self._calls = 0
        self.aborted: list[tuple[str, str]] = []

    async def list_multipart_uploads(self, Bucket, **kwargs) -> dict:
        idx = self._calls
        self._calls += 1
        is_last = idx == len(self._pages) - 1
        return {"Uploads": self._pages[idx], "IsTruncated": not is_last}

    async def abort_multipart_upload(self, Bucket, Key, UploadId) -> dict:
        self.aborted.append((Key, UploadId))
        return {}


def _ctx(s3: FakeS3) -> dict:
    return {"container": SimpleNamespace(s3=s3)}


async def test_aborts_only_stale_uploads() -> None:
    stale = datetime.now(UTC) - timedelta(hours=2)
    fresh = datetime.now(UTC) - timedelta(minutes=5)
    s3 = FakeS3([[_upload("a", "u1", stale), _upload("b", "u2", fresh)]])

    aborted = await cleanup_abandoned_uploads(_ctx(s3))

    assert aborted == 1
    assert s3.aborted == [("a", "u1")]


async def test_paginates_listing() -> None:
    stale = datetime.now(UTC) - timedelta(hours=2)
    s3 = FakeS3([[_upload("a", "u1", stale)], [_upload("b", "u2", stale)]])

    aborted = await cleanup_abandoned_uploads(_ctx(s3))

    assert aborted == 2
    assert s3._calls == 2
    assert s3.aborted == [("a", "u1"), ("b", "u2")]


async def test_treats_naive_datetime_as_utc() -> None:
    naive_stale = (datetime.now(UTC) - timedelta(hours=2)).replace(tzinfo=None)
    s3 = FakeS3([[_upload("a", "u1", naive_stale)]])

    aborted = await cleanup_abandoned_uploads(_ctx(s3))

    assert aborted == 1
    assert s3.aborted == [("a", "u1")]
