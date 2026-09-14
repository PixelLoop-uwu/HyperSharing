"""Aggregate router for the versioned API.

Feature routers are included here as they are added.
"""

from fastapi import APIRouter

from app.features.posts.router import router as posts_router
from app.features.users.router import router as users_router

api_router = APIRouter()
api_router.include_router(posts_router)
api_router.include_router(users_router)
