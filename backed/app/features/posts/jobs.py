"""Background jobs for the post feature (draft/publish upload flow).

Jobs run in the arq worker; they receive the worker ``ctx`` (a dict) and read
the shared infrastructure clients from ``ctx["container"]`` (a ``Container``,
populated by ``on_startup`` in ``app.infrastructure.jobs.worker``).
"""

from datetime import UTC, datetime, timedelta
from typing import Any

from app.core.config import settings
from app.infrastructure.container import require_not_none
from app.infrastructure.storage.multipart import MultipartUploadManager


async def cleanup_abandoned_uploads(ctx: dict[str, Any]) -> int:
    """Abort multipart uploads left incomplete after their draft expired.

    S3 is the source of truth here, not Redis: a draft's TTL deletes its
    ``upload_id`` from Redis, so the job cannot rely on the draft to know which
    uploads to abort. Instead:

    1. ``list_multipart_uploads`` on the bucket to find in-progress uploads;
    2. abort those whose ``Initiated`` is older than
       ``abandoned_upload_max_age_seconds``.

    Returns the number of aborted uploads.
    """
    container = ctx["container"]
    manager = MultipartUploadManager(
        require_not_none(container.s3, "s3"), settings.s3_bucket
    )

    cutoff = datetime.now(UTC) - timedelta(
        seconds=settings.abandoned_upload_max_age_seconds
    )

    aborted = 0
    for upload in await manager.list_uploads():
        initiated = upload["Initiated"]
        if initiated.tzinfo is None:
            # MinIO may return a naive datetime; treat it as UTC to compare.
            initiated = initiated.replace(tzinfo=UTC)
        if initiated > cutoff:
            continue
        await manager.abort(upload["Key"], upload["UploadId"])
        aborted += 1
    return aborted
