"""ASGI middleware: request-id binding for logs."""

import uuid
from collections.abc import Awaitable, Callable

from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response

from app.core.config import settings
from app.core.logging import request_id_ctx


class RequestIDMiddleware(BaseHTTPMiddleware):
    """Reads/generates a request id and binds it into the logging context."""

    async def dispatch(
        self, request: Request, call_next: Callable[[Request], Awaitable[Response]]
    ) -> Response:
        request_id = request.headers.get(settings.request_id_header) or uuid.uuid4().hex
        token = request_id_ctx.set(request_id)
        try:
            response = await call_next(request)
        finally:
            request_id_ctx.reset(token)
        response.headers[settings.request_id_header] = request_id
        return response
