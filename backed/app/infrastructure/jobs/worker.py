"""arq worker settings and lifecycle.

Run from the backend root:

    uv run arq app.infrastructure.jobs.worker.WorkerSettings

This package owns only the arq integration (Redis settings, ``WorkerSettings``,
lifecycle). Business tasks live in their feature packages
(``features/*/jobs.py``) and are registered below.

Jobs run in a separate process from the API and need their own infrastructure
clients. ``on_startup`` reuses the same lifespan composition as the API
(``populate_container``) and stashes the resulting ``Container`` in
``ctx["container"]``; ``on_shutdown`` tears it down in reverse order. arq passes
the worker a plain dict as ``ctx`` (not the ``Worker`` instance).
"""

from __future__ import annotations

from typing import Any

from arq.cron import CronJob, cron

from app.core.config import settings
from app.features.posts.jobs import cleanup_abandoned_uploads
from app.infrastructure.jobs.settings import make_redis_settings
from app.infrastructure.lifespan import populate_container


async def on_startup(ctx: dict[str, Any]) -> None:
    container_cm = populate_container()
    ctx["container_cm"] = container_cm
    ctx["container"] = await container_cm.__aenter__()


async def on_shutdown(ctx: dict[str, Any]) -> None:
    await ctx["container_cm"].__aexit__(None, None, None)


class WorkerSettings:
    functions = [cleanup_abandoned_uploads]
    cron_jobs: list[CronJob] = [
        cron(cleanup_abandoned_uploads, minute={0, 10, 20, 30, 40, 50}),
    ]
    redis_settings = make_redis_settings()
    max_jobs = settings.jobs_max_concurrent
    job_timeout = settings.job_timeout_seconds
    on_startup = on_startup
    on_shutdown = on_shutdown
