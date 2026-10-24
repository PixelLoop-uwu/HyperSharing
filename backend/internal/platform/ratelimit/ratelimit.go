// Package ratelimit — простое fixed-window ограничение частоты в Redis.
package ratelimit

import (
  "context"
  "fmt"
  "net/http"
  "time"

  "github.com/redis/go-redis/v9"

  "hypersharing/internal/httpx"
)

type Limiter struct {
  rdb *redis.Client
}

func New(rdb *redis.Client) *Limiter {
  return &Limiter{rdb: rdb}
}

// Allow возвращает true, если ключу разрешён ещё один запрос в текущем окне.
// Окно фиксированное: счётчик INCR, TTL ставится только на первом запросе окна
// (иначе каждый запрос сдвигал бы окно).
func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
  n, err := allowScript.Run(ctx, l.rdb, []string{key}, int64(window.Seconds())).Int64()
  if err != nil {
    return false, fmt.Errorf("ratelimit: %w", err)
  }
  return n <= int64(limit), nil
}

// Middleware оборачивает handler проверкой частоты. Ключ берётся из запроса
// через key (например, IP). При превышении отвечает 429.
func (l *Limiter) Middleware(key func(*http.Request) string, limit int, window time.Duration) func(http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      ok, err := l.Allow(r.Context(), key(r), limit, window)
      if err != nil {
        httpx.Internal(w, r, err)
        return
      }
      if !ok {
        httpx.Error(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
        return
      }
      next.ServeHTTP(w, r)
    })
  }
}

var allowScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return n
`)
