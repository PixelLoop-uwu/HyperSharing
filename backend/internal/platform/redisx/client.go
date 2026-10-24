package redisx

import (
  "context"
  "fmt"
  "time"

  "github.com/redis/go-redis/v9"
)

func New(ctx context.Context, url string) (*redis.Client, error) {
  opts, err := redis.ParseURL(url)
  if err != nil {
    return nil, fmt.Errorf("parse redis url: %w", err)
  }
  client := redis.NewClient(opts)

  pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  if err := client.Ping(pctx).Err(); err != nil {
    client.Close()
    return nil, fmt.Errorf("ping redis: %w", err)
  }
  return client, nil
}