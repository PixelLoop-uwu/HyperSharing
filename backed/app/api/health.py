"""Health-check that actually exercises DB, Redis and S3."""

from fastapi import APIRouter
from fastapi.responses import JSONResponse
from sqlalchemy import text

from app.core.config import settings
from app.infrastructure.cache.redis_client import RedisDep
from app.infrastructure.db.session import DbSessionDep
from app.infrastructure.storage.s3_client import S3Dep

router = APIRouter()


@router.get("/healthz")
async def healthz(
    db: DbSessionDep,
    redis: RedisDep,
    s3: S3Dep,
) -> JSONResponse:
    checks: dict[str, str] = {}

    try:
        await db.execute(text("SELECT 1"))
        checks["database"] = "ok"
    except Exception as exc:  # noqa: BLE001 - report upstream as degraded
        checks["database"] = f"error: {exc}"

    try:
        await redis.ping()
        checks["redis"] = "ok"
    except Exception as exc:  # noqa: BLE001
        checks["redis"] = f"error: {exc}"

    try:
        await s3.head_bucket(Bucket=settings.s3_bucket)
        checks["s3"] = "ok"
    except Exception as exc:  # noqa: BLE001
        checks["s3"] = f"error: {exc}"

    healthy = all(value == "ok" for value in checks.values())
    return JSONResponse(
        status_code=200 if healthy else 503,
        content={"status": "ok" if healthy else "degraded", "checks": checks},
    )
