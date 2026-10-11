# Development commands for Hypersharing.
# FastAPI backend (uv) + Vite frontend (npm). Shared infra runs in Docker.
# Run `just --list` to see all recipes.

# --- Infrastructure (Postgres + Redis + MinIO) ---

# Start the shared infrastructure only.
infra:
    docker compose up -d postgres redis minio minio-init

# Stop the shared infrastructure.
stop:
    docker compose down

# --- Backend ---

# Apply database migrations.
migrate:
    cd backend && uv run alembic upgrade head

# Run the backend API with hot reload (starts infra + migrates first).
backend: infra migrate
    cd backend && uv run uvicorn app.main:app --reload

# --- Frontend ---

# Run the frontend dev server (Vite HMR, proxies /api → :8000).
frontend:
    cd frontend && npm run dev

# --- Everything ---

# Run backend + frontend together (dev). Ctrl+C stops both.
dev:
    #!/usr/bin/env bash
    set -euo pipefail
    docker compose up -d postgres redis minio minio-init
    (cd backend && uv run alembic upgrade head)
    (cd backend && uv run uvicorn app.main:app --reload) &
    BACKEND_PID=$!
    (cd frontend && npm run dev) &
    FRONTEND_PID=$!
    trap 'kill "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true' EXIT INT TERM
    wait "$BACKEND_PID" "$FRONTEND_PID"

# Run the full containerized stack (backend + worker in Docker). Frontend stays dev-only.
up:
    docker compose up -d --build

# Follow logs of the containerized backend.
logs:
    docker compose logs -f app worker
