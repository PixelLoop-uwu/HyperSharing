package postgres

import (
  "context"
  "fmt"
  "time"

  "github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, url string) (*pgxpool.Pool, error) {
  pool, err := pgxpool.New(ctx, url)
  if err != nil {
    return nil, fmt.Errorf("create pool: %w", err)
  }

  pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  if err := pool.Ping(pctx); err != nil {
    pool.Close()
    return nil, fmt.Errorf("ping postgres: %w", err)
  }
  return pool, nil
}