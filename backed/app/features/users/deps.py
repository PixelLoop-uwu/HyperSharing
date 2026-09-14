"""Authentication dependencies (cross-cutting for other features)."""

from typing import Annotated

from fastapi import Depends, Request

from app.core.config import settings
from app.core.exceptions import UnauthorizedError
from app.features.users.models import User
from app.features.users.repository import SessionStoreDep, UserRepositoryDep


async def get_current_user(
    request: Request,
    session_store: SessionStoreDep,
    user_repo: UserRepositoryDep,
) -> User | None:
    """Resolve the logged-in user from the session cookie, or ``None``.

    Anonymous requests and stale sessions (token unknown in Redis, or user
    deleted) both yield ``None`` so callers can treat them as "not logged in".
    On success the session TTL is slid forward.
    """
    token = request.cookies.get(settings.session_cookie_name)
    if not token:
        return None
    session = await session_store.get(token)
    if session is None:
        return None
    user = await user_repo.get_by_id(session.user_id)
    if user is None:
        return None
    await session_store.touch(token, session)
    return user


CurrentUserDep = Annotated[User | None, Depends(get_current_user)]


async def get_active_user(current_user: CurrentUserDep) -> User:
    """Like ``CurrentUserDep`` but raises 401 when there is no session."""
    if current_user is None:
        raise UnauthorizedError("authentication required")
    return current_user


ActiveSessionDep = Annotated[User, Depends(get_active_user)]
