"""S3 object key scheme for content-addressed storage."""


def s3_object_key(file_hash: str) -> str:
    """Return the object key for a content hash.

    Files are stored by hash so identical content is stored exactly once; the
    key is derived rather than stored in the database.
    """
    return f"objects/{file_hash}"
