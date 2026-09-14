"""Pydantic schemas for users and sessions."""

import uuid
from datetime import datetime

from pydantic import BaseModel, ConfigDict, Field

from app.core.config import settings


class UserRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    email: str
    username: str
    created_at: datetime


class RegisterIn(BaseModel):
    email: str = Field(max_length=255, pattern=settings.email_pattern)
    username: str = Field(pattern=settings.username_pattern)
    password: str = Field(min_length=settings.password_min_length, max_length=128)


class LoginIn(BaseModel):
    email: str = Field(max_length=255)
    # No min_length here: a short password is simply a wrong password (401),
    # not a malformed request (422).
    password: str


class SessionData(BaseModel):
    """What gets serialized to Redis for an authenticated session."""

    user_id: uuid.UUID
