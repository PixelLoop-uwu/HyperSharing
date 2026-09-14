"""Shared application exceptions.

Each exception carries an HTTP status code and a stable machine-readable
``code``. The global handler in ``app/main.py`` turns these into JSON errors.
"""


class AppError(Exception):
    """Base class for all application-level errors."""

    status_code: int = 500
    code: str = "internal_error"

    def __init__(
        self, detail: str = "", *, headers: dict[str, str] | None = None
    ) -> None:
        self.detail = detail or self.__class__.__name__
        self.headers = headers or {}
        super().__init__(detail)


class NotFoundError(AppError):
    status_code = 404
    code = "not_found"


class UnauthorizedError(AppError):
    status_code = 401
    code = "unauthorized"


class ForbiddenError(AppError):
    status_code = 403
    code = "forbidden"


class ConflictError(AppError):
    status_code = 409
    code = "conflict"


class ValidationError(AppError):
    status_code = 422
    code = "validation_error"


class RateLimitError(AppError):
    status_code = 429
    code = "rate_limited"


class StorageError(AppError):
    """Upstream storage (S3/MinIO) failure."""

    status_code = 502
    code = "storage_error"
