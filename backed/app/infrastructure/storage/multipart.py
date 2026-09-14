"""Multipart (chunked) upload manager.

Separate from ``BaseS3Store`` — this is the flow for the post-draft feature:
``create_multipart_upload`` → presigned URL per part → client uploads chunks
directly to S3 → ``complete_multipart_upload`` with ETag/size verification.
"""

from dataclasses import dataclass
from typing import Annotated, Any, cast

from botocore.exceptions import ClientError
from fastapi import Depends

from app.core.config import settings
from app.core.exceptions import StorageError
from app.infrastructure.storage.s3_client import S3Client, S3Dep, S3PublicDep


@dataclass(frozen=True)
class CompletedPart:
    part_number: int
    etag: str


class MultipartUploadManager:
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

    async def create(self, key: str, *, content_type: str | None = None) -> str:
        """Open a multipart upload and return its ``upload_id``.

        No ``ChecksumAlgorithm`` here: per-part integrity is enforced at
        ``upload_part`` time, where each part carries ``x-amz-checksum-sha256``
        (signed into its presigned URL). Setting the algorithm on the multipart
        upload itself breaks MinIO's ``CompleteMultipartUpload`` — it rejects
        the parts with ``InvalidPart`` — so it is intentionally left unset.
        """
        kwargs: dict[str, str] = {}
        if content_type is not None:
            kwargs["ContentType"] = content_type
        response = await self._s3_client.create_multipart_upload(
            Bucket=self._bucket,
            Key=key,
            **kwargs,
        )
        return cast(str, response["UploadId"])

    async def presign_part(
        self,
        key: str,
        upload_id: str,
        part_number: int,
        *,
        checksum_sha256: str,
        expires_in: int = settings.s3_presign_ttl_seconds,
    ) -> str:
        """Return a presigned ``upload_part`` URL for one part.

        The checksum headers are signed into the URL, so the client must send
        ``x-amz-checksum-sha256: <checksum_sha256>`` and
        ``x-amz-sdk-checksum-algorithm: SHA256`` verbatim when PUTting the part.
        """
        return await self._presign_client.generate_presigned_url(
            "upload_part",
            Params={
                "Bucket": self._bucket,
                "Key": key,
                "UploadId": upload_id,
                "PartNumber": part_number,
                "ChecksumAlgorithm": "SHA256",
                "ChecksumSHA256": checksum_sha256,
            },
            ExpiresIn=expires_in,
        )

    async def complete(
        self,
        key: str,
        upload_id: str,
        parts: list[CompletedPart],
        *,
        expected_size: int | None = None,
    ) -> dict[str, Any]:
        """Finish the upload, verifying the total size before finalizing.

        Idempotent: if the object already exists at ``key``, a prior attempt
        already assembled it (or the content was deduplicated from an earlier
        post). The key is content-addressed, so an existing object holds the
        same bytes; ``ContentLength`` still guards the expected size.
        """
        try:
            head = await self._s3_client.head_object(Bucket=self._bucket, Key=key)
        except ClientError as exc:
            if exc.response["Error"]["Code"] not in {"NoSuchKey", "404", "NotFound"}:
                raise
        else:
            if (
                expected_size is not None
                and int(head["ContentLength"]) != expected_size
            ):
                raise StorageError(
                    f"size mismatch for {key}: existing object has "
                    f"{head['ContentLength']} bytes, expected {expected_size}"
                )
            return {}

        if expected_size is not None:
            listed = await self._s3_client.list_parts(
                Bucket=self._bucket, Key=key, UploadId=upload_id
            )
            actual_size = sum(int(p["Size"]) for p in listed.get("Parts", []))
            if actual_size != expected_size:
                raise StorageError(
                    f"size mismatch for {key}: "
                    f"expected {expected_size}, got {actual_size}"
                )

        return await self._s3_client.complete_multipart_upload(
            Bucket=self._bucket,
            Key=key,
            UploadId=upload_id,
            MultipartUpload={
                "Parts": [{"PartNumber": p.part_number, "ETag": p.etag} for p in parts]
            },
        )

    async def list_uploads(self) -> list[dict[str, Any]]:
        """Return all in-progress multipart uploads (paginated).

        Each entry is the raw S3 ``Upload`` dict (``Key``, ``UploadId``,
        ``Initiated``, …). ``Initiated`` is a tz-aware ``datetime``.
        """
        uploads: list[dict[str, Any]] = []
        kwargs: dict[str, Any] = {}
        while True:
            response = await self._s3_client.list_multipart_uploads(
                Bucket=self._bucket, **kwargs
            )
            page = response.get("Uploads", [])
            uploads.extend(page)
            if not response.get("IsTruncated") or not page:
                return uploads
            last = page[-1]
            kwargs["KeyMarker"] = last["Key"]
            kwargs["UploadIdMarker"] = last["UploadId"]

    async def abort(self, key: str, upload_id: str) -> None:
        """Abandon an incomplete upload (cleanup job for expired drafts)."""
        await self._s3_client.abort_multipart_upload(
            Bucket=self._bucket, Key=key, UploadId=upload_id
        )


async def get_multipart_manager(
    s3: S3Dep, s3_public: S3PublicDep
) -> MultipartUploadManager:
    """Build the manager with the internal client plus the public presign client."""
    return MultipartUploadManager(s3, settings.s3_bucket, presign_client=s3_public)


MultipartManagerDep = Annotated[MultipartUploadManager, Depends(get_multipart_manager)]
