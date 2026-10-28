package draft

import (
  "context"
  "encoding/json"
  "errors"
  "testing"
  "time"

  "github.com/alicebob/miniredis/v2"
  "github.com/google/uuid"
  "github.com/redis/go-redis/v9"
)

func newTestStore(t *testing.T) (*Store, *redis.Client) {
  t.Helper()
  mr := miniredis.RunT(t)
  rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
  t.Cleanup(func() { _ = rdb.Close() })
  return NewStore(rdb, time.Hour), rdb
}

func seedSlot(t *testing.T, s *Store, rdb *redis.Client, draftID string, fileID uuid.UUID, status Status, size int64) {
  t.Helper()
  ctx := context.Background()
  structure, _ := json.Marshal(map[string]string{"title": "t"})
  slot := Slot{Name: "a.txt", Mime: "text/plain", Size: size, Section: uuid.NewString(), Position: 0, Status: status}
  if err := s.Create(ctx, draftID, OwnerUserPrefix+"abc", size, structure, map[string]Slot{fileID.String(): slot}, 2); err != nil {
    t.Fatalf("seed: %v", err)
  }
}

// --- CRUD ---

func TestCreateGetSlotsDelete(t *testing.T) {
  s, rdb := newTestStore(t)
  ctx := context.Background()
  id := uuid.NewString()
  owner := OwnerUserPrefix + "abc"

  structure := json.RawMessage(`{"title":"t","sections":[{"id":"` + uuid.NewString() + `","body":"b"}]}`)
  slots := map[string]Slot{
    uuid.NewString(): {Name: "a", Mime: "x", Size: 1, Section: "s", Status: Pending},
  }

  if err := s.Create(ctx, id, owner, 1, structure, slots, 2); err != nil {
    t.Fatalf("create: %v", err)
  }

  d, err := s.Get(ctx, id)
  if err != nil {
    t.Fatalf("get: %v", err)
  }
  if d.Owner != owner || d.TotalSize != 1 || string(d.Structure) != string(structure) {
    t.Fatalf("bad draft: %+v", d)
  }

  if got, err := s.Slots(ctx, id); err != nil || len(got) != 1 {
    t.Fatalf("slots: %v %v", got, err)
  }

  if err := s.Delete(ctx, owner, id); err != nil {
    t.Fatalf("delete: %v", err)
  }
  if n, _ := rdb.Exists(ctx, key(id)).Result(); n != 0 {
    t.Fatalf("draft key not deleted")
  }
  if n, _ := rdb.Exists(ctx, filesKey(id)).Result(); n != 0 {
    t.Fatalf("files key not deleted")
  }
}

func TestGetNotFound(t *testing.T) {
  s, _ := newTestStore(t)
  if _, err := s.Get(context.Background(), uuid.NewString()); !errors.Is(err, ErrDraftNotFound) {
    t.Fatalf("want ErrDraftNotFound, got %v", err)
  }
}

func TestActiveDraftLimit(t *testing.T) {
  s, _ := newTestStore(t)
  ctx := context.Background()
  owner := OwnerUserPrefix + "abc"

  for i := 0; i < 2; i++ {
    if err := s.Create(ctx, uuid.NewString(), owner, 0, json.RawMessage(`{}`), nil, 2); err != nil {
      t.Fatalf("create %d: %v", i, err)
    }
  }
  if err := s.Create(ctx, uuid.NewString(), owner, 0, json.RawMessage(`{}`), nil, 2); !errors.Is(err, ErrTooManyDrafts) {
    t.Fatalf("want ErrTooManyDrafts, got %v", err)
  }
}

// --- переходы статусов ---

func TestAcquireSlot(t *testing.T) {
  s, rdb := newTestStore(t)
  ctx := context.Background()
  draftID := uuid.NewString()
  fileID := uuid.New()
  seedSlot(t, s, rdb, draftID, fileID, Pending, 42)

  slot, err := s.AcquireSlot(ctx, draftID, fileID, 42)
  if err != nil {
    t.Fatalf("acquire: %v", err)
  }
  if slot.Status != Uploading {
    t.Fatalf("want uploading, got %q", slot.Status)
  }
}

func TestAcquireSlotSizeMismatch(t *testing.T) {
  s, rdb := newTestStore(t)
  draftID := uuid.NewString()
  fileID := uuid.New()
  seedSlot(t, s, rdb, draftID, fileID, Pending, 42)

  if _, err := s.AcquireSlot(context.Background(), draftID, fileID, 43); !errors.Is(err, ErrSizeMismatch) {
    t.Fatalf("want ErrSizeMismatch, got %v", err)
  }
}

func TestAcquireSlotDraftNotFound(t *testing.T) {
  s, _ := newTestStore(t)
  if _, err := s.AcquireSlot(context.Background(), uuid.NewString(), uuid.New(), 1); !errors.Is(err, ErrDraftNotFound) {
    t.Fatalf("want ErrDraftNotFound, got %v", err)
  }
}

func TestAcquireSlotTwice(t *testing.T) {
  s, rdb := newTestStore(t)
  ctx := context.Background()
  draftID := uuid.NewString()
  fileID := uuid.New()
  seedSlot(t, s, rdb, draftID, fileID, Pending, 42)

  if _, err := s.AcquireSlot(ctx, draftID, fileID, 42); err != nil {
    t.Fatalf("first acquire: %v", err)
  }
  if _, err := s.AcquireSlot(ctx, draftID, fileID, 42); !errors.Is(err, ErrSlotNotPending) {
    t.Fatalf("want ErrSlotNotPending, got %v", err)
  }
}

func TestFinishSlot(t *testing.T) {
  s, rdb := newTestStore(t)
  ctx := context.Background()
  draftID := uuid.NewString()
  fileID := uuid.New()
  seedSlot(t, s, rdb, draftID, fileID, Uploading, 42)

  slot, err := s.FinishSlot(ctx, draftID, fileID, "files/"+fileID.String())
  if err != nil {
    t.Fatalf("finish: %v", err)
  }
  if slot.Status != Done || slot.StorageKey != "files/"+fileID.String() {
    t.Fatalf("want done + storage_key, got %q %q", slot.Status, slot.StorageKey)
  }
}

func TestFinishSlotPendingFails(t *testing.T) {
  s, rdb := newTestStore(t)
  draftID := uuid.NewString()
  fileID := uuid.New()
  seedSlot(t, s, rdb, draftID, fileID, Pending, 42)

  if _, err := s.FinishSlot(context.Background(), draftID, fileID, "files/x"); !errors.Is(err, ErrSlotNotPending) {
    t.Fatalf("want ErrSlotNotPending, got %v", err)
  }
}

func TestReleaseSlotReturnsToPending(t *testing.T) {
  s, rdb := newTestStore(t)
  ctx := context.Background()
  draftID := uuid.NewString()
  fileID := uuid.New()
  seedSlot(t, s, rdb, draftID, fileID, Uploading, 42)

  if err := s.ReleaseSlot(ctx, draftID, fileID); err != nil {
    t.Fatalf("release: %v", err)
  }
  if _, err := s.AcquireSlot(ctx, draftID, fileID, 42); err != nil {
    t.Fatalf("reacquire after release: %v", err)
  }
}

func TestReleaseSlotDoneIsNoop(t *testing.T) {
  s, rdb := newTestStore(t)
  ctx := context.Background()
  draftID := uuid.NewString()
  fileID := uuid.New()
  seedSlot(t, s, rdb, draftID, fileID, Done, 42)

  if err := s.ReleaseSlot(ctx, draftID, fileID); err != nil {
    t.Fatalf("release: %v", err)
  }
  if _, err := s.AcquireSlot(ctx, draftID, fileID, 42); !errors.Is(err, ErrSlotNotPending) {
    t.Fatalf("done slot must stay done, got %v", err)
  }
}

func TestOwnerOf(t *testing.T) {
  s, rdb := newTestStore(t)
  ctx := context.Background()
  draftID := uuid.NewString()
  fileID := uuid.New()
  seedSlot(t, s, rdb, draftID, fileID, Pending, 42)

  owner, err := s.OwnerOf(ctx, draftID)
  if err != nil || owner != OwnerUserPrefix+"abc" {
    t.Fatalf("owner: %q %v", owner, err)
  }
  if _, err := s.OwnerOf(ctx, uuid.NewString()); !errors.Is(err, ErrDraftNotFound) {
    t.Fatalf("want ErrDraftNotFound, got %v", err)
  }
}
