"""Logging configuration (loguru).

Request id travels in a context variable: the middleware in
``app/middleware.py`` sets it per-request, and the ``patcher`` below injects it
into every record's ``extra`` so each log line is correlated to its request.
"""

import contextvars
import sys
from typing import Any

from loguru import logger

request_id_ctx: contextvars.ContextVar[str | None] = contextvars.ContextVar(
    "request_id", default=None
)


def _bind_request_id(record: Any) -> None:
    record["extra"]["request_id"] = request_id_ctx.get() or "-"


def configure_logging(*, debug: bool = False) -> None:
    logger.remove()
    logger.configure(patcher=_bind_request_id)
    if debug:
        logger.add(
            sys.stderr,
            level="DEBUG",
            format=(
                "<green>{time:HH:mm:ss}</green> | "
                "<level>{level: <8}</level> | "
                "<cyan>{extra[request_id]}</cyan> | "
                "<blue>{name}:{function}:{line}</blue> | "
                "<level>{message}</level>"
            ),
        )
    else:
        logger.add(sys.stderr, level="INFO", serialize=True)
