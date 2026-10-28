package storage

import (
  "context"
  "errors"
  "testing"

  "github.com/alicebob/miniredis/v2"
  "github.com/redis/go-redis/v9"
)

func newTestQuota(t *testing.T) *Quota {
  t.Helper()
  mr := miniredis.RunT(t)
  rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
  t.Cleanup(func() { _ = rdb.Close() })
  return NewQuota(rdb, 100)
}

func TestQuotaReserveUnderLimit(t *testing.T) {
  q := newTestQuota(t)
  ctx := context.Background()
  if err := q.Reserve(ctx, "ip:abc", 100); err != nil {
    t.Fatalf("under limit: %v", err)
  }
}

func TestQuotaReserveOverLimit(t *testing.T) {
  q := newTestQuota(t)
  ctx := context.Background()
  if err := q.Reserve(ctx, "ip:abc", 101); !errors.Is(err, ErrQuotaExceeded) {
    t.Fatalf("want ErrQuotaExceeded, got %v", err)
  }
}

func TestQuotaReserveAccumulates(t *testing.T) {
  q := newTestQuota(t)
  ctx := context.Background()
  if err := q.Reserve(ctx, "ip:abc", 60); err != nil {
    t.Fatalf("first reserve: %v", err)
  }
  if err := q.Reserve(ctx, "ip:abc", 41); !errors.Is(err, ErrQuotaExceeded) {
    t.Fatalf("want ErrQuotaExceeded on 60+41, got %v", err)
  }
}

func TestQuotaReleaseAllowsReReserve(t *testing.T) {
  q := newTestQuota(t)
  ctx := context.Background()
  if err := q.Reserve(ctx, "ip:abc", 100); err != nil {
    t.Fatalf("reserve: %v", err)
  }
  if err := q.Release(ctx, "ip:abc", 60); err != nil {
    t.Fatalf("release: %v", err)
  }
  // после возврата 60 можно снова занять 60 (40+60=100)
  if err := q.Reserve(ctx, "ip:abc", 60); err != nil {
    t.Fatalf("re-reserve after release: %v", err)
  }
}
