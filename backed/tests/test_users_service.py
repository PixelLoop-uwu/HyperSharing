"""Unit tests for UserService (register/login/logout)."""

import uuid

import pytest

from app.core.exceptions import ConflictError, RateLimitError, UnauthorizedError
from app.core.security import hash_password, verify_password
from app.features.users import schemas
from app.features.users.models import User
from app.features.users.service import UserService
from app.infrastructure.cache.rate_limiter import RateLimitResult


def _user(
    email: str = "a@b.com",
    username: str = "alice",
    password: str = "secret123",
) -> User:
    return User(email=email, username=username, password_hash=hash_password(password))


class FakeUserRepository:
    """In-memory stand-in for UserRepository."""

    def __init__(self, users: list[User] | None = None) -> None:
        self.users = list(users or [])

    async def get_by_email(self, email: str) -> User | None:
        return next((u for u in self.users if u.email == email), None)

    async def get_by_username(self, username: str) -> User | None:
        return next((u for u in self.users if u.username == username), None)

    async def create(self, user: User) -> User:
        user.id = uuid.uuid4()
        self.users.append(user)
        return user


class FakeSessionStore:
    def __init__(self) -> None:
        self.sessions: dict[str, schemas.SessionData] = {}
        self.deleted: list[str] = []

    async def set(self, token: str, session: schemas.SessionData) -> None:
        self.sessions[token] = session

    async def delete(self, token: str) -> None:
        self.deleted.append(token)
        self.sessions.pop(token, None)


class FakeRateLimiter:
    def __init__(self, result: RateLimitResult) -> None:
        self._result = result
        self.last_key: str | None = None

    async def hit(
        self, key: str, *, limit: int, window_seconds: int
    ) -> RateLimitResult:
        self.last_key = key
        return self._result


def _allowed() -> FakeRateLimiter:
    """A limiter that always allows the request."""
    return FakeRateLimiter(
        RateLimitResult(allowed=True, remaining=4, retry_after_seconds=0)
    )


def _service(
    repo: FakeUserRepository,
    store: FakeSessionStore,
    limiter: FakeRateLimiter,
) -> UserService:
    return UserService(repo, store, limiter)


async def test_register_creates_user_and_session() -> None:
    repo = FakeUserRepository()
    store = FakeSessionStore()
    svc = _service(repo, store, _allowed())

    token, user = await svc.register(
        schemas.RegisterIn(
            email="Alice@Example.COM", username="alice", password="secret123"
        )
    )

    assert user.email == "alice@example.com"
    assert user.username == "alice"
    assert verify_password("secret123", user.password_hash)
    assert token in store.sessions
    assert store.sessions[token].user_id == user.id


async def test_register_duplicate_email_conflicts() -> None:
    existing = _user()
    existing.id = uuid.uuid4()
    svc = _service(FakeUserRepository([existing]), FakeSessionStore(), _allowed())

    with pytest.raises(ConflictError):
        await svc.register(
            schemas.RegisterIn(email="a@b.com", username="bob", password="secret123")
        )


async def test_register_duplicate_username_conflicts() -> None:
    existing = _user()
    existing.id = uuid.uuid4()
    svc = _service(FakeUserRepository([existing]), FakeSessionStore(), _allowed())

    with pytest.raises(ConflictError):
        await svc.register(
            schemas.RegisterIn(
                email="other@b.com", username="alice", password="secret123"
            )
        )


async def test_login_returns_user_and_creates_session() -> None:
    existing = _user()
    existing.id = uuid.uuid4()
    store = FakeSessionStore()
    limiter = _allowed()
    svc = _service(FakeUserRepository([existing]), store, limiter)

    token, user = await svc.login(
        schemas.LoginIn(email="a@b.com", password="secret123")
    )

    assert user is existing
    assert token in store.sessions
    assert store.sessions[token].user_id == existing.id
    assert limiter.last_key == "rate:login:a@b.com"


async def test_login_wrong_password_unauthorized() -> None:
    existing = _user()
    existing.id = uuid.uuid4()
    svc = _service(FakeUserRepository([existing]), FakeSessionStore(), _allowed())

    with pytest.raises(UnauthorizedError):
        await svc.login(schemas.LoginIn(email="a@b.com", password="wrongpass"))


async def test_login_unknown_email_unauthorized() -> None:
    svc = _service(FakeUserRepository(), FakeSessionStore(), _allowed())

    with pytest.raises(UnauthorizedError):
        await svc.login(schemas.LoginIn(email="nobody@b.com", password="secret123"))


async def test_login_rate_limited() -> None:
    limiter = FakeRateLimiter(
        RateLimitResult(allowed=False, remaining=0, retry_after_seconds=42)
    )
    svc = _service(FakeUserRepository(), FakeSessionStore(), limiter)

    with pytest.raises(RateLimitError) as exc_info:
        await svc.login(schemas.LoginIn(email="a@b.com", password="secret123"))

    assert exc_info.value.headers == {"Retry-After": "42"}


async def test_logout_deletes_session() -> None:
    store = FakeSessionStore()
    store.sessions["tok"] = schemas.SessionData(user_id=uuid.uuid4())
    svc = _service(FakeUserRepository(), store, _allowed())

    await svc.logout("tok")

    assert "tok" not in store.sessions
    assert store.deleted == ["tok"]
