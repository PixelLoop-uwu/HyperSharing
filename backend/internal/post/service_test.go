package post

import (
  "context"
  "errors"
  "testing"
  "time"

  "github.com/alicebob/miniredis/v2"
  "github.com/redis/go-redis/v9"

  "hypersharing/internal/draft"
  "hypersharing/internal/storage"
)

// newTestService собирает Service на miniredis. pool = nil: анонимный flow БД не
// трогает (UserStorageUsed вызывается только для user-владельца), публикацию
// здесь не тестируем — она требует реального Postgres.
func newTestService(t *testing.T) (*Service, *redis.Client) {
  t.Helper()
  mr := miniredis.RunT(t)
  rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
  t.Cleanup(func() { _ = rdb.Close() })
  quota := storage.NewQuota(rdb, 1000, 10000)
  return New(nil, rdb, quota, draft.NewStore(rdb, time.Hour), testLimits(), "/api/v1"), rdb
}

func TestServiceCreateDraftAnon(t *testing.T) {
  svc, _ := newTestService(t)
  ctx := context.Background()
  owner := draft.OwnerIPPrefix + "deadbeef"

  d, err := svc.CreateDraft(ctx, owner, CreateDraftInput{
    Title:    "hello",
    Sections: []SectionInput{{Body: "text", Files: []FileInput{{Name: "a.txt", Mime: "text/plain", Size: 10}}}},
  })
  if err != nil {
    t.Fatalf("create: %v", err)
  }
  if d.ID == "" || len(d.Sections) != 1 || len(d.Sections[0].Files) != 1 {
    t.Fatalf("bad response: %+v", d)
  }

  // драфт и слот легли в Redis
  if _, err := svc.drafts.Get(ctx, d.ID); err != nil {
    t.Fatalf("draft not found after create: %v", err)
  }
}

func TestServiceCreateDraftTooManyActive(t *testing.T) {
  svc, _ := newTestService(t)
  ctx := context.Background()
  owner := draft.OwnerIPPrefix + "deadbeef"

  for i := 0; i < 2; i++ {
    if _, err := svc.CreateDraft(ctx, owner, CreateDraftInput{Sections: []SectionInput{{Body: "x"}}}); err != nil {
      t.Fatalf("create %d: %v", i, err)
    }
  }
  if _, err := svc.CreateDraft(ctx, owner, CreateDraftInput{Sections: []SectionInput{{Body: "x"}}}); !errors.Is(err, draft.ErrTooManyDrafts) {
    t.Fatalf("want ErrTooManyDrafts, got %v", err)
  }
}

func TestServiceCancelDraft(t *testing.T) {
  svc, _ := newTestService(t)
  ctx := context.Background()
  owner := draft.OwnerIPPrefix + "deadbeef"

  d, err := svc.CreateDraft(ctx, owner, CreateDraftInput{
    Sections: []SectionInput{{Files: []FileInput{{Name: "a", Mime: "x", Size: 10}}}},
  })
  if err != nil {
    t.Fatalf("create: %v", err)
  }
  if err := svc.CancelDraft(ctx, d.ID, owner); err != nil {
    t.Fatalf("cancel: %v", err)
  }
  if _, err := svc.drafts.Get(ctx, d.ID); !errors.Is(err, draft.ErrDraftNotFound) {
    t.Fatalf("draft must be gone, got %v", err)
  }
  if err := svc.CancelDraft(ctx, d.ID, owner); !errors.Is(err, draft.ErrDraftNotFound) {
    t.Fatalf("second cancel want ErrDraftNotFound, got %v", err)
  }
}

func TestServiceCancelDraftWrongOwner(t *testing.T) {
  svc, _ := newTestService(t)
  ctx := context.Background()

  d, err := svc.CreateDraft(ctx, draft.OwnerIPPrefix+"aaaa", CreateDraftInput{
    Sections: []SectionInput{{Files: []FileInput{{Name: "a", Mime: "x", Size: 10}}}},
  })
  if err != nil {
    t.Fatalf("create: %v", err)
  }
  // чужой владелец не может отменить драфт; для него это выглядит как отсутствующий (404)
  if err := svc.CancelDraft(ctx, d.ID, draft.OwnerIPPrefix+"bbbb"); !errors.Is(err, draft.ErrDraftNotFound) {
    t.Fatalf("want ErrDraftNotFound, got %v", err)
  }
  // драфт остался
  if _, err := svc.drafts.Get(ctx, d.ID); err != nil {
    t.Fatalf("draft must survive, got %v", err)
  }
}

func TestServiceRejectsMimeWithControlChars(t *testing.T) {
  svc, _ := newTestService(t)
  ctx := context.Background()
  _, err := svc.CreateDraft(ctx, draft.OwnerIPPrefix+"deadbeef", CreateDraftInput{
    Sections: []SectionInput{{Files: []FileInput{{Name: "a", Mime: "text/plain\r\nX: 1", Size: 1}}}},
  })
  if !isValidationErr(err) {
    t.Fatalf("want ValidationError, got %v", err)
  }
}
