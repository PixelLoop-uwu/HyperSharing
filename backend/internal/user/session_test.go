package user

import (
  "context"
  "errors"
  "testing"
  "time"

  "github.com/alicebob/miniredis/v2"
  "github.com/google/uuid"
  "github.com/redis/go-redis/v9"
)

func newTestStore(t *testing.T) *SessionStore {
  t.Helper()
  mr := miniredis.RunT(t)
  rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
  t.Cleanup(func() { _ = rdb.Close() })
  return NewSessionStore(rdb, time.Hour, 24*time.Hour)
}

func TestSessionCreateResolve(t *testing.T) {
  s := newTestStore(t)
  ctx := context.Background()
  id := uuid.New()

  token, err := s.Create(ctx, id)
  if err != nil {
    t.Fatalf("create: %v", err)
  }
  got, err := s.Resolve(ctx, token)
  if err != nil {
    t.Fatalf("resolve: %v", err)
  }
  if got != id {
    t.Fatalf("resolve: got %v want %v", got, id)
  }
}

func TestSessionResolveUnknownToken(t *testing.T) {
  s := newTestStore(t)
  if _, err := s.Resolve(context.Background(), "no-such-token"); !errors.Is(err, ErrSessionNotFound) {
    t.Fatalf("want ErrSessionNotFound, got %v", err)
  }
}

func TestSessionRevoke(t *testing.T) {
  s := newTestStore(t)
  ctx := context.Background()
  id := uuid.New()

  token, _ := s.Create(ctx, id)
  if err := s.Revoke(ctx, token); err != nil {
    t.Fatalf("revoke: %v", err)
  }
  if _, err := s.Resolve(ctx, token); !errors.Is(err, ErrSessionNotFound) {
    t.Fatalf("after revoke want ErrSessionNotFound, got %v", err)
  }
}

func TestSessionRevokeAll(t *testing.T) {
  s := newTestStore(t)
  ctx := context.Background()
  id := uuid.New()

  t1, _ := s.Create(ctx, id)
  t2, _ := s.Create(ctx, id)
  if err := s.RevokeAll(ctx, id); err != nil {
    t.Fatalf("revoke all: %v", err)
  }
  for _, token := range []string{t1, t2} {
    if _, err := s.Resolve(ctx, token); !errors.Is(err, ErrSessionNotFound) {
      t.Fatalf("token should be gone, got %v", err)
    }
  }
}

func TestSessionAbsoluteMaxAge(t *testing.T) {
  s := newTestStore(t)
  s.maxAge = time.Millisecond
  ctx := context.Background()

  token, _ := s.Create(ctx, uuid.New())
  time.Sleep(2 * time.Millisecond)

  if _, err := s.Resolve(ctx, token); !errors.Is(err, ErrSessionNotFound) {
    t.Fatalf("want ErrSessionNotFound (max age), got %v", err)
  }
}
