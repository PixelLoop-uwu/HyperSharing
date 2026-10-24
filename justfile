set dotenv-load

# Start the shared infrastructure only.
infra:
  docker compose up -d --wait postgres redis minio
  docker compose up -d minio-init

# Apply database migrations.
migrate:
  cd backend && goose -dir migrations postgres "$DATABASE_URL" up

# Create the next migration: just migrate-new create_posts
migrate-new name:
  cd backend && goose -dir migrations -s create {{name}} sql

# Run the backend API with hot reload (starts infra + migrates first).
backend: infra migrate
  cd backend && air --build.cmd "go build -o ./tmp/app ./cmd/app" --build.bin "./tmp/app"

# Run the cleanup worker (infra + migrations first).
worker: infra migrate
  cd backend && go run ./cmd/worker

# Unit tests (fast, no Docker).
test:
  cd backend && go test ./...

# Integration tests (real Postgres/MinIO via testcontainers — needs Docker).
test-integration:
  cd backend && go test -tags=integration ./...

# Run backend + frontend together (dev). Ctrl+C stops both.
dev:
  #!/usr/bin/env bash
  set -euo pipefail
  docker compose up -d --wait postgres redis minio
  docker compose up -d minio-init
  (cd backend && goose -dir migrations postgres "$DATABASE_URL" up)
  (cd backend && air --build.cmd "go build -o ./tmp/app ./cmd/app" --build.bin "./tmp/app") &
  BACKEND_PID=$!
  (cd frontend && npm run dev) &
  FRONTEND_PID=$!
  trap 'kill "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true' EXIT INT TERM
  wait "$BACKEND_PID" "$FRONTEND_PID"