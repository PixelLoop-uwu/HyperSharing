package post

import (
  "context"
  "testing"
  "time"

  "github.com/alicebob/miniredis/v2"
  "github.com/google/uuid"
  "github.com/redis/go-redis/v9"

  "hypersharing/internal/draft"
  "hypersharing/internal/storage"
)

// newTestService собирает Service на miniredis без БД: здесь тестируем только
// валидацию (она отрабатывает до обращения к БД). Жизненный цикл драфта/поста —
// в integration_test.go (реальный Postgres).
func newTestService(t *testing.T) (*Service, *redis.Client) {
  t.Helper()
  mr := miniredis.RunT(t)
  rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
  t.Cleanup(func() { _ = rdb.Close() })
  quota := storage.NewQuota(rdb, 10000)
  return New(nil, rdb, quota, draft.NewStore(rdb, time.Hour), testLimits(), "/api/v1"), rdb
}

func TestServiceRejectsMimeWithControlChars(t *testing.T) {
  svc, _ := newTestService(t)
  ctx := context.Background()
  _, err := svc.CreateDraft(ctx, uuid.New(), CreateDraftInput{
    Sections: []SectionInput{{Files: []FileInput{{Name: "a", Mime: "text/plain\r\nX: 1", Size: 1}}}},
  })
  if !isValidationErr(err) {
    t.Fatalf("want ValidationError, got %v", err)
  }
}
