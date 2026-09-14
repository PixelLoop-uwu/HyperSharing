"""Application entrypoint.

Run with: ``uvicorn app.main:app --reload``
"""

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from app.api.health import router as health_router
from app.api.router import api_router
from app.core.config import settings
from app.core.exceptions import AppError
from app.core.logging import configure_logging
from app.infrastructure.lifespan import lifespan
from app.middleware import RequestIDMiddleware

configure_logging(debug=settings.debug)

app = FastAPI(title=settings.app_name, lifespan=lifespan)

app.add_middleware(RequestIDMiddleware)


@app.exception_handler(AppError)
async def app_error_handler(request: Request, exc: AppError) -> JSONResponse:
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": {"code": exc.code, "detail": exc.detail}},
        headers=exc.headers,
    )


app.include_router(health_router)
app.include_router(api_router, prefix=settings.api_v1_prefix)
