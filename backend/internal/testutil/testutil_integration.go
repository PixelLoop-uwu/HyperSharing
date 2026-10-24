//go:build integration

package testutil

import (
  "context"
  "database/sql"
  "fmt"
  "path/filepath"
  "runtime"
  "testing"
  "time"

  "github.com/jackc/pgx/v5/pgxpool"
  _ "github.com/jackc/pgx/v5/stdlib" // драйвер "pgx" для database/sql (goose)
  "github.com/pressly/goose/v3"
  "github.com/testcontainers/testcontainers-go"
  "github.com/testcontainers/testcontainers-go/modules/postgres"
  "github.com/testcontainers/testcontainers-go/wait"
)

// S3 — данные подключения к MinIO-контейнеру (клиента собирает сам тест, чтобы
// не тащить сюда импорт s3x и не создавать цикл импортов).
type S3 struct {
  Endpoint  string
  AccessKey string
  SecretKey string
}

// Postgres поднимает контейнер Postgres, прогоняет миграции goose и возвращает
// готовый пул.
func Postgres(t *testing.T) *pgxpool.Pool {
  t.Helper()
  ctx := context.Background()

  c, err := postgres.Run(ctx, "postgres:18-alpine",
    postgres.WithUsername("postgres"),
    postgres.WithPassword("postgres"),
    testcontainers.WithWaitStrategy(
      wait.ForLog("database system is ready to accept connections").WithStartupTimeout(2*time.Minute),
    ),
  )
  if err != nil {
    t.Fatalf("start postgres: %v", err)
  }
  t.Cleanup(func() { _ = c.Terminate(ctx) })

  dsn, err := c.ConnectionString(ctx, "sslmode=disable")
  if err != nil {
    t.Fatalf("connection string: %v", err)
  }

  // Постгрес логирует "ready" при временном старте (init), затем перезапускается —
  // подождём, пока он реально начнёт принимать соединения.
  pool, err := connectPostgres(ctx, dsn)
  if err != nil {
    t.Fatalf("connect postgres: %v", err)
  }
  t.Cleanup(pool.Close)

  db, err := sql.Open("pgx", dsn)
  if err != nil {
    t.Fatalf("sql open: %v", err)
  }
  defer db.Close()
  if err := goose.SetDialect("postgres"); err != nil {
    t.Fatal(err)
  }
  if err := goose.Up(db, migrationsDir()); err != nil {
    t.Fatalf("migrate: %v", err)
  }
  return pool
}

// connectPostgres — ретраи соединение, пока постгрес не примет его (см. выше).
func connectPostgres(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
  var lastErr error
  for i := 0; i < 30; i++ {
    pool, err := pgxpool.New(ctx, dsn)
    if err == nil {
      pctx, cancel := context.WithTimeout(ctx, 2*time.Second)
      err = pool.Ping(pctx)
      cancel()
      if err == nil {
        return pool, nil
      }
      pool.Close()
    }
    lastErr = err
    time.Sleep(500 * time.Millisecond)
  }
  return nil, lastErr
}

// MinIO поднимает контейнер MinIO и возвращает данные подключения к S3-API.
func MinIO(t *testing.T) S3 {
  t.Helper()
  ctx := context.Background()

  req := testcontainers.ContainerRequest{
    Image:        "minio/minio:latest",
    ExposedPorts: []string{"9000/tcp"},
    Env: map[string]string{
      "MINIO_ROOT_USER":     "minioadmin",
      "MINIO_ROOT_PASSWORD": "minioadmin",
    },
    Cmd:        []string{"server", "/data"},
    WaitingFor: wait.ForListeningPort("9000/tcp"),
  }
  c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
    ContainerRequest: req,
    Started:          true,
  })
  if err != nil {
    t.Fatalf("start minio: %v", err)
  }
  t.Cleanup(func() { _ = c.Terminate(ctx) })

  host, err := c.Host(ctx)
  if err != nil {
    t.Fatal(err)
  }
  port, err := c.MappedPort(ctx, "9000/tcp")
  if err != nil {
    t.Fatal(err)
  }
  return S3{
    Endpoint:  fmt.Sprintf("%s:%s", host, port.Port()),
    AccessKey: "minioadmin",
    SecretKey: "minioadmin",
  }
}

func migrationsDir() string {
  _, file, _, ok := runtime.Caller(0)
  if !ok {
    return "migrations"
  }
  return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}
