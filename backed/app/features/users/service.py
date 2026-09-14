"""Business logic for users and sessions."""

from typing import Annotated

from fastapi import Depends
from sqlalchemy.exc import IntegrityError

from app.core.config import settings
from app.core.exceptions import ConflictError, RateLimitError, UnauthorizedError
from app.core.security import generate_token, hash_password, verify_password
from app.features.users import schemas
from app.features.users.models import User
from app.features.users.repository import (
    SessionStore,
    SessionStoreDep,
    UserRepository,
    UserRepositoryDep,
)
from app.infrastructure.cache.rate_limiter import RateLimiter, RateLimiterDep


class UserService:
    def __init__(
        self,
        user_repo: UserRepository,
        session_store: SessionStore,
        rate_limiter: RateLimiter,
    ) -> None:
        self._user_repo = user_repo
        self._session_store = session_store
        self._rate_limiter = rate_limiter

    async def register(self, data: schemas.RegisterIn) -> tuple[str, User]:
        """Create a user and open a session for them; returns (token, user)."""
        email = data.email.strip().lower()
        username = data.username.strip()
        if await self._user_repo.get_by_email(email) is not None:
            raise ConflictError("email already registered")
        if await self._user_repo.get_by_username(username) is not None:
            raise ConflictError("username already taken")

        user = User(
            email=email,
            username=username,
            password_hash=hash_password(data.password),
        )
        try:
            await self._user_repo.create(user)
        except IntegrityError:
            # Race between the two checks and the INSERT: the unique indexes are
            # the source of truth. The request's DB session rolls back on raise.
            raise ConflictError("email or username already registered") from None

        token = generate_token()
        await self._session_store.set(token, schemas.SessionData(user_id=user.id))
        return token, user

    async def login(self, data: schemas.LoginIn) -> tuple[str, User]:
        """Verify credentials and open a session; returns (token, user)."""
        email = data.email.strip().lower()

        result = await self._rate_limiter.hit(
            f"rate:login:{email}",
            limit=settings.login_rate_limit,
            window_seconds=settings.login_rate_window_seconds,
        )
        if not result.allowed:
            raise RateLimitError(
                f"too many login attempts ({settings.login_rate_limit} per "
                f"{settings.login_rate_window_seconds}s)",
                headers={"Retry-After": str(result.retry_after_seconds)},
            )

        user = await self._user_repo.get_by_email(email)
        # One generic message for "unknown email" and "wrong password" so the
        # endpoint does not reveal which emails are registered.
        if user is None or not verify_password(data.password, user.password_hash):
            raise UnauthorizedError("invalid email or password")

        token = generate_token()
        await self._session_store.set(token, schemas.SessionData(user_id=user.id))
        return token, user

    async def logout(self, token: str) -> None:
        await self._session_store.delete(token)


async def get_user_service(
    user_repo: UserRepositoryDep,
    session_store: SessionStoreDep,
    rate_limiter: RateLimiterDep,
) -> UserService:
    return UserService(user_repo, session_store, rate_limiter)


UserServiceDep = Annotated[UserService, Depends(get_user_service)]
