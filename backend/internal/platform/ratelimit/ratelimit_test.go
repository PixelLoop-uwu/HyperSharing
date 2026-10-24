package ratelimit

import (
  "context"
  "net/http"
  "net/http/httptest"
  "testing"
  "time"

  "github.com/alicebob/miniredis/v2"
  "github.com/redis/go-redis/v9"
)

func TestAllow(t *testing.T) {
  mr := miniredis.RunT(t)
  rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
  t.Cleanup(func() { _ = rdb.Close() })
  l := New(rdb)
  ctx := context.Background()

  for i := 0; i < 5; i++ {
    ok, err := l.Allow(ctx, "k", 5, time.Minute)
    if err != nil || !ok {
      t.Fatalf("call %d should be allowed: ok=%v err=%v", i, ok, err)
    }
  }
  if ok, _ := l.Allow(ctx, "k", 5, time.Minute); ok {
    t.Fatal("6th call should be rate limited")
  }
}

func TestAllowSeparateKeys(t *testing.T) {
  mr := miniredis.RunT(t)
  rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
  t.Cleanup(func() { _ = rdb.Close() })
  l := New(rdb)
  ctx := context.Background()

  // один ключ не влияет на другой
  for i := 0; i < 10; i++ {
    if ok, _ := l.Allow(ctx, "a", 1, time.Minute); ok {
      // первый вызов проходит, остальные — нет; неважно, проверяем лишь ключ b
    }
  }
  if ok, _ := l.Allow(ctx, "b", 1, time.Minute); !ok {
    t.Fatal("separate key should not be affected")
  }
}

func TestMiddleware(t *testing.T) {
  mr := miniredis.RunT(t)
  rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
  t.Cleanup(func() { _ = rdb.Close() })

  var calls int
  h := New(rdb).Middleware(func(*http.Request) string { return "k" }, 2, time.Minute)(
    http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { calls++; w.WriteHeader(http.StatusNoContent) }),
  )

  for i := 0; i < 2; i++ {
    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
    if rec.Code != http.StatusNoContent {
      t.Fatalf("call %d: status %d, want 204", i, rec.Code)
    }
  }
  rec := httptest.NewRecorder()
  h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
  if rec.Code != http.StatusTooManyRequests {
    t.Fatalf("want 429, got %d", rec.Code)
  }
  if calls != 2 {
    t.Fatalf("handler called %d times, want 2", calls)
  }
}
