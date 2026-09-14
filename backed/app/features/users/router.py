"""User endpoints (register/login/logout/me)."""

from fastapi import APIRouter, Request, Response, status

from app.core.config import settings
from app.features.users import schemas
from app.features.users.deps import ActiveSessionDep
from app.features.users.service import UserServiceDep

router = APIRouter(prefix="/auth", tags=["auth"])


def _set_session_cookie(response: Response, token: str) -> None:
    response.set_cookie(
        key=settings.session_cookie_name,
        value=token,
        max_age=settings.session_ttl_seconds,
        httponly=True,
        samesite="lax",
        secure=settings.session_cookie_secure,
        path="/",
    )


@router.post(
    "/register",
    response_model=schemas.UserRead,
    status_code=status.HTTP_201_CREATED,
)
async def register(
    data: schemas.RegisterIn,
    service: UserServiceDep,
    response: Response,
) -> schemas.UserRead:
    token, user = await service.register(data)
    _set_session_cookie(response, token)
    return schemas.UserRead.model_validate(user)


@router.post("/login", response_model=schemas.UserRead)
async def login(
    data: schemas.LoginIn,
    service: UserServiceDep,
    response: Response,
) -> schemas.UserRead:
    token, user = await service.login(data)
    _set_session_cookie(response, token)
    return schemas.UserRead.model_validate(user)


@router.post("/logout", status_code=status.HTTP_204_NO_CONTENT)
async def logout(
    request: Request,
    service: UserServiceDep,
    response: Response,
) -> None:
    token = request.cookies.get(settings.session_cookie_name)
    if token is not None:
        await service.logout(token)
    response.delete_cookie(settings.session_cookie_name, path="/")


@router.get("/me", response_model=schemas.UserRead)
async def me(user: ActiveSessionDep) -> schemas.UserRead:
    return schemas.UserRead.model_validate(user)
