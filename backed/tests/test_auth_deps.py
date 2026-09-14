"""Unit tests for the auth dependencies (get_current_user / get_active_user)."""

import uuid
from types import SimpleNamespace

import pytest

from app.core.exceptions import UnauthorizedError
from app.features.users import schemas
from app.features.users.deps import get_active_user, get_current_user
from app.features.users.models import User


def _user() -> User:
    user = User(email="a@b.com", username="alice", password_hash="x")
    user.id = uuid.uuid4()
    return user


class FakeSessionStore:
    def __init__(self, sessions: dict[str, schemas.SessionData] | None = None) -> None:
        self.sessions = dict(sessions or {})
        self.touched: list[tuple[str, schemas.SessionData]] = []

    async def get(self, token: str) -> schemas.SessionData | None:
        return self.sessions.get(token)

    async def touch(self, token: str, session: schemas.SessionData) -> None:
        self.touched.append((token, session))


class FakeUserRepository:
    def __init__(self, users: list[User] | None = None) -> None:
        self._users = {u.id: u for u in (users or [])}

    async def get_by_id(self, user_id: uuid.UUID) -> User | None:
        return self._users.get(user_id)


def _request(cookie: str | None) -> SimpleNamespace:
    cookies = {"sid": cookie} if cookie is not None else {}
    return SimpleNamespace(cookies=cookies)


async def test_no_cookie_is_anonymous() -> None:
    user = await get_current_user(
        _request(None), FakeSessionStore(), FakeUserRepository()
    )
    assert user is None


async def test_unknown_token_is_anonymous() -> None:
    user = await get_current_user(
        _request("ghost"), FakeSessionStore(), FakeUserRepository()
    )
    assert user is None


async def test_valid_session_returns_user_and_touches() -> None:
    user = _user()
    store = FakeSessionStore({"tok": schemas.SessionData(user_id=user.id)})
    repo = FakeUserRepository([user])

    result = await get_current_user(_request("tok"), store, repo)

    assert result is user
    assert store.touched == [("tok", schemas.SessionData(user_id=user.id))]


async def test_orphaned_session_is_anonymous() -> None:
    store = FakeSessionStore({"tok": schemas.SessionData(user_id=uuid.uuid4())})
    repo = FakeUserRepository([])

    result = await get_current_user(_request("tok"), store, repo)

    assert result is None


async def test_active_user_raises_when_anonymous() -> None:
    with pytest.raises(UnauthorizedError):
        await get_active_user(None)


async def test_active_user_returns_user() -> None:
    user = _user()
    assert await get_active_user(user) is user
