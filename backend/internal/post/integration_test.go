//go:build integration

package post

import (
  "context"
  "errors"
  "fmt"
  "testing"
  "time"

  "github.com/alicebob/miniredis/v2"
  "github.com/google/uuid"
  "github.com/jackc/pgx/v5/pgxpool"
  "github.com/redis/go-redis/v9"

  "hypersharing/internal/draft"
  "hypersharing/internal/storage"
  "hypersharing/internal/testutil"
  "hypersharing/internal/user"
)

// newSvc собирает post.Service на реальном Postgres + miniredis (Redis).
func newSvc(t *testing.T, limits Limits) (*Service, *pgxpool.Pool) {
  t.Helper()
  pool := testutil.Postgres(t)
  mr := miniredis.RunT(t)
  rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
  t.Cleanup(func() { _ = rdb.Close() })
  quota := storage.NewQuota(rdb, 10<<20)
  return New(pool, rdb, quota, draft.NewStore(rdb, time.Hour), limits, "/api/v1"), pool
}

func createUser(t *testing.T, pool *pgxpool.Pool, name string) uuid.UUID {
  t.Helper()
  u, err := user.NewRepository(pool).CreateUser(context.Background(), name, name+"@example.com", "x")
  if err != nil {
    t.Fatalf("create user: %v", err)
  }
  return u.ID
}

func TestPostLifecycle(t *testing.T) {
  svc, pool := newSvc(t, testLimits())
  ctx := context.Background()
  userID := createUser(t, pool, "alice")

  d, err := svc.CreateDraft(ctx, userID, CreateDraftInput{
    Title:    "hello",
    Sections: []SectionInput{{Body: "text", Files: []FileInput{{Name: "a.txt", Mime: "text/plain", Size: 10}}}},
  })
  if err != nil {
    t.Fatalf("create draft: %v", err)
  }

  fileID := uuid.MustParse(d.Sections[0].Files[0].ID)
  // имитируем завершённую tus-загрузку: pending → uploading → done
  if _, err := svc.drafts.AcquireSlot(ctx, d.ID, fileID, 10); err != nil {
    t.Fatalf("acquire slot: %v", err)
  }
  if _, err := svc.drafts.FinishSlot(ctx, d.ID, fileID, "files/"+fileID.String()); err != nil {
    t.Fatalf("finish slot: %v", err)
  }

  published, err := svc.Publish(ctx, d.ID, draft.Owner(userID))
  if err != nil {
    t.Fatalf("publish: %v", err)
  }

  got, err := svc.Get(ctx, published.ID)
  if err != nil {
    t.Fatalf("get: %v", err)
  }
  if got.Title != "hello" || len(got.Sections) != 1 || len(got.Sections[0].Files) != 1 {
    t.Fatalf("bad post: %+v", got)
  }
  if got.Sections[0].Files[0].URL == "" {
    t.Fatal("file url empty")
  }

  // публичное чтение по короткому ключу
  if published.ShortKey == "" {
    t.Fatal("post short_key empty")
  }
  byKey, err := svc.GetByShortKey(ctx, published.ShortKey)
  if err != nil {
    t.Fatalf("get by short key: %v", err)
  }
  if byKey.ID != published.ID {
    t.Fatalf("get by short key mismatch: %v != %v", byKey.ID, published.ID)
  }

  // идемпотентность публикации
  again, err := svc.Publish(ctx, d.ID, draft.Owner(userID))
  if err != nil {
    t.Fatalf("second publish: %v", err)
  }
  if again.ID != published.ID {
    t.Fatalf("idempotency: %v != %v", again.ID, published.ID)
  }
}

func TestCancelDraft(t *testing.T) {
  svc, pool := newSvc(t, testLimits())
  ctx := context.Background()
  userID := createUser(t, pool, "bob")

  d, err := svc.CreateDraft(ctx, userID, CreateDraftInput{
    Sections: []SectionInput{{Files: []FileInput{{Name: "a", Mime: "x", Size: 10}}}},
  })
  if err != nil {
    t.Fatalf("create draft: %v", err)
  }
  if err := svc.CancelDraft(ctx, d.ID, draft.Owner(userID)); err != nil {
    t.Fatalf("cancel: %v", err)
  }
  if _, err := svc.drafts.Get(ctx, d.ID); !errors.Is(err, draft.ErrDraftNotFound) {
    t.Fatalf("draft must be gone, got %v", err)
  }
}

func TestListMineAndDelete(t *testing.T) {
  svc, pool := newSvc(t, testLimits())
  ctx := context.Background()
  userID := createUser(t, pool, "alice")

  ids := make([]uuid.UUID, 2)
  for i := 0; i < 2; i++ {
    d, err := svc.CreateDraft(ctx, userID, CreateDraftInput{Title: fmt.Sprintf("post %d", i)})
    if err != nil {
      t.Fatalf("create draft %d: %v", i, err)
    }
    p, err := svc.Publish(ctx, d.ID, draft.Owner(userID))
    if err != nil {
      t.Fatalf("publish %d: %v", i, err)
    }
    ids[i] = p.ID
  }

  // страница 1 (limit 1) → есть курсор
  page1, err := svc.ListMine(ctx, userID, 1, "")
  if err != nil {
    t.Fatalf("list: %v", err)
  }
  if len(page1.Posts) != 1 || page1.NextCursor == nil {
    t.Fatalf("page1: %+v", page1)
  }
  // страница 2 — вторая запись, курсора больше нет
  page2, err := svc.ListMine(ctx, userID, 1, *page1.NextCursor)
  if err != nil {
    t.Fatalf("list page2: %v", err)
  }
  if len(page2.Posts) != 1 || page2.NextCursor != nil {
    t.Fatalf("page2: %+v", page2)
  }
  seen := map[uuid.UUID]bool{page1.Posts[0].ID: true, page2.Posts[0].ID: true}
  if !seen[ids[0]] || !seen[ids[1]] || page1.Posts[0].ID == page2.Posts[0].ID {
    t.Fatalf("pagination overlap/duplicate: %v %v", page1.Posts[0].ID, page2.Posts[0].ID)
  }

  if err := svc.Delete(ctx, ids[0], userID); err != nil {
    t.Fatalf("delete: %v", err)
  }
  if _, err := svc.Get(ctx, ids[0]); !errors.Is(err, ErrPostNotFound) {
    t.Fatalf("want ErrPostNotFound after delete, got %v", err)
  }
}
