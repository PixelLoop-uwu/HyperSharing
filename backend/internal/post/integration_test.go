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
  quota := storage.NewQuota(rdb, 1<<20, 10<<20)
  return New(pool, rdb, quota, draft.NewStore(rdb, time.Hour), limits), pool
}

func TestPostLifecycleAnon(t *testing.T) {
  svc, _ := newSvc(t, testLimits())
  ctx := context.Background()
  owner := draft.OwnerIPPrefix + "deadbeef"

  d, err := svc.CreateDraft(ctx, owner, CreateDraftInput{
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

  published, err := svc.Publish(ctx, d.ID, owner)
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

  // идемпотентность публикации
  again, err := svc.Publish(ctx, d.ID, owner)
  if err != nil {
    t.Fatalf("second publish: %v", err)
  }
  if again.ID != published.ID {
    t.Fatalf("idempotency: %v != %v", again.ID, published.ID)
  }
}

func TestPostExpiredAnon(t *testing.T) {
  limits := testLimits()
  limits.AnonTTL = 500 * time.Millisecond // достаточно для публикации, но быстро истекает
  svc, _ := newSvc(t, limits)
  ctx := context.Background()

  d, err := svc.CreateDraft(ctx, draft.OwnerIPPrefix+"deadbeef", CreateDraftInput{Sections: []SectionInput{{Body: "x"}}})
  if err != nil {
    t.Fatalf("create draft: %v", err)
  }
  p, err := svc.Publish(ctx, d.ID, draft.OwnerIPPrefix+"deadbeef")
  if err != nil {
    t.Fatalf("publish: %v", err)
  }

  time.Sleep(time.Second)
  if _, err := svc.Get(ctx, p.ID); !errors.Is(err, ErrPostNotFound) {
    t.Fatalf("want ErrPostNotFound (expired), got %v", err)
  }
}

func TestListMineAndDelete(t *testing.T) {
  svc, pool := newSvc(t, testLimits())
  ctx := context.Background()

  u, err := user.NewRepository(pool).CreateUser(ctx, "alice", "alice@example.com", "x")
  if err != nil {
    t.Fatalf("create user: %v", err)
  }
  owner := draft.OwnerUserPrefix + u.ID.String()

  ids := make([]uuid.UUID, 2)
  for i := 0; i < 2; i++ {
    d, err := svc.CreateDraft(ctx, owner, CreateDraftInput{Title: fmt.Sprintf("post %d", i)})
    if err != nil {
      t.Fatalf("create draft %d: %v", i, err)
    }
    p, err := svc.Publish(ctx, d.ID, owner)
    if err != nil {
      t.Fatalf("publish %d: %v", i, err)
    }
    ids[i] = p.ID
  }

  // страница 1 (limit 1) → есть курсор
  page1, err := svc.ListMine(ctx, u.ID, 1, "")
  if err != nil {
    t.Fatalf("list: %v", err)
  }
  if len(page1.Posts) != 1 || page1.NextCursor == nil {
    t.Fatalf("page1: %+v", page1)
  }
  // страница 2 — вторая запись, курсора больше нет
  page2, err := svc.ListMine(ctx, u.ID, 1, *page1.NextCursor)
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

  if err := svc.Delete(ctx, ids[0], u.ID); err != nil {
    t.Fatalf("delete: %v", err)
  }
  if _, err := svc.Get(ctx, ids[0]); !errors.Is(err, ErrPostNotFound) {
    t.Fatalf("want ErrPostNotFound after delete, got %v", err)
  }
}
