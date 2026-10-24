package draft

import (
  "context"
  "encoding/json"
  "fmt"
  "strconv"
  "time"

  "github.com/google/uuid"
  "github.com/redis/go-redis/v9"
)

// Store — доступ к драфтам и их слотам. Переходы статусов атомарны через Lua,
// чтобы в слот нельзя было загрузить дважды и чтобы «draft жив + слот pending»
// проверялось одним вызовом.
type Store struct {
  rdb *redis.Client
  ttl time.Duration
}

func NewStore(rdb *redis.Client, ttl time.Duration) *Store {
  return &Store{rdb: rdb, ttl: ttl}
}

// Create атомарно проверяет лимит активных драфтов (ZSET), пишет hash драфта с
// TTL, затем слоты отдельным pipeline. При ошибке после ZADD драфт останется в
// ZSET, но его счистит ZREMRANGEBYSCORE на следующей проверке.
func (s *Store) Create(ctx context.Context, id, owner string, totalSize int64, structure json.RawMessage, slots map[string]Slot, maxDrafts int) error {
  now := time.Now().Unix()
  res, err := createScript.Run(ctx, s.rdb,
    []string{draftsKey(owner), key(id)},
    now, now+int64(s.ttl.Seconds()), maxDrafts, id, owner, totalSize, string(structure), int64(s.ttl.Seconds()), int64((2 * s.ttl).Seconds()),
  ).Slice()
  if err != nil {
    return fmt.Errorf("draft: create: %w", err)
  }
  if len(res) > 0 {
    if code, _ := res[0].(string); code == "err" {
      return ErrTooManyDrafts
    }
  }

  pipe := s.rdb.Pipeline()
  for fileID, slot := range slots {
    raw, err := json.Marshal(slot)
    if err != nil {
      return fmt.Errorf("draft: marshal slot: %w", err)
    }
    pipe.HSet(ctx, filesKey(id), fileID, raw)
  }
  pipe.Expire(ctx, filesKey(id), s.ttl)
  if _, err := pipe.Exec(ctx); err != nil {
    return fmt.Errorf("draft: write slots: %w", err)
  }
  return nil
}

func (s *Store) Get(ctx context.Context, id string) (Draft, error) {
  m, err := s.rdb.HGetAll(ctx, key(id)).Result()
  if err != nil {
    return Draft{}, fmt.Errorf("draft: get: %w", err)
  }
  if len(m) == 0 {
    return Draft{}, ErrDraftNotFound
  }
  totalSize, _ := strconv.ParseInt(m[totalSizeField], 10, 64)
  return Draft{
    Owner:     m[ownerField],
    TotalSize: totalSize,
    Structure: json.RawMessage(m[structureField]),
  }, nil
}

func (s *Store) Slots(ctx context.Context, id string) (map[string]Slot, error) {
  m, err := s.rdb.HGetAll(ctx, filesKey(id)).Result()
  if err != nil {
    return nil, fmt.Errorf("draft: slots: %w", err)
  }
  slots := make(map[string]Slot, len(m))
  for fileID, raw := range m {
    var slot Slot
    if err := json.Unmarshal([]byte(raw), &slot); err != nil {
      return nil, fmt.Errorf("draft: decode slot %s: %w", fileID, err)
    }
    slots[fileID] = slot
  }
  return slots, nil
}

// Delete удаляет оба ключа драфта и вынимает id из ZSET активных.
func (s *Store) Delete(ctx context.Context, owner, id string) error {
  pipe := s.rdb.Pipeline()
  pipe.Del(ctx, key(id), filesKey(id))
  pipe.ZRem(ctx, draftsKey(owner), id)
  if _, err := pipe.Exec(ctx); err != nil {
    return fmt.Errorf("draft: delete: %w", err)
  }
  return nil
}

// OwnerOf возвращает владельца драфта. ErrDraftNotFound, если драфта нет.
func (s *Store) OwnerOf(ctx context.Context, id string) (string, error) {
  n, err := s.rdb.Exists(ctx, key(id)).Result()
  if err != nil {
    return "", fmt.Errorf("draft: exists: %w", err)
  }
  if n == 0 {
    return "", ErrDraftNotFound
  }
  owner, err := s.rdb.HGet(ctx, key(id), ownerField).Result()
  if err != nil && err != redis.Nil {
    return "", fmt.Errorf("draft: owner: %w", err)
  }
  return owner, nil
}

// AcquireSlot атомарно переводит слот pending → uploading, сверив заявленный
// размер и что драфт жив.
func (s *Store) AcquireSlot(ctx context.Context, id string, fileID uuid.UUID, size int64) (Slot, error) {
  res, err := acquireScript.Run(ctx, s.rdb,
    []string{key(id), filesKey(id)},
    fileID.String(), size,
  ).Slice()
  if err != nil {
    return Slot{}, fmt.Errorf("draft: acquire slot: %w", err)
  }
  return decodeSlotResult(res, "acquire")
}

// FinishSlot помечает слот done и записывает storage_key. Идемпотентно для done;
// pending — ошибка (нельзя завершить незагруженный слот).
func (s *Store) FinishSlot(ctx context.Context, id string, fileID uuid.UUID, storageKey string) (Slot, error) {
  res, err := finishScript.Run(ctx, s.rdb,
    []string{filesKey(id)},
    fileID.String(), storageKey,
  ).Slice()
  if err != nil {
    return Slot{}, fmt.Errorf("draft: finish slot: %w", err)
  }
  return decodeSlotResult(res, "finish")
}

// ReleaseSlot возвращает слот uploading → pending после отмены/прерывания
// tus-загрузки. Для done или отсутствующего слота — no-op.
func (s *Store) ReleaseSlot(ctx context.Context, id string, fileID uuid.UUID) error {
  _, err := releaseScript.Run(ctx, s.rdb,
    []string{filesKey(id)},
    fileID.String(),
  ).Result()
  if err != nil {
    return fmt.Errorf("draft: release slot: %w", err)
  }
  return nil
}

// slotResult — контракт Lua-скриптов: {"ok", <json>} или {"err", "<code>"}.
func decodeSlotResult(res []interface{}, op string) (Slot, error) {
  if len(res) == 0 {
    return Slot{}, fmt.Errorf("draft: %s: empty script result", op)
  }
  code, _ := res[0].(string)
  if code == "err" {
    return Slot{}, slotError(res[1])
  }
  raw, _ := res[1].(string)
  var slot Slot
  if err := json.Unmarshal([]byte(raw), &slot); err != nil {
    return Slot{}, fmt.Errorf("draft: %s: decode slot: %w", op, err)
  }
  return slot, nil
}

func slotError(v interface{}) error {
  code, _ := v.(string)
  switch code {
  case "draft_not_found":
    return ErrDraftNotFound
  case "slot_not_found":
    return ErrSlotNotFound
  case "slot_not_pending":
    return ErrSlotNotPending
  case "size_mismatch":
    return ErrSizeMismatch
  default:
    return fmt.Errorf("draft: unknown slot error %q", code)
  }
}

// createScript: чистка просроченных + проверка лимита + ZADD + HSET hash.
// KEYS[1]=drafts:{owner}, KEYS[2]=draft:{id}
// ARGV: now, expiry, maxDrafts, draftID, owner, totalSize, structure, ttl, zsetTTL
var createScript = redis.NewScript(`
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
if redis.call('ZCARD', KEYS[1]) >= tonumber(ARGV[3]) then
  return {'err', 'too_many_drafts'}
end
redis.call('ZADD', KEYS[1], ARGV[2], ARGV[4])
redis.call('HSET', KEYS[2], 'owner', ARGV[5], 'total_size', ARGV[6], 'structure', ARGV[7])
redis.call('EXPIRE', KEYS[2], ARGV[8])
redis.call('EXPIRE', KEYS[1], ARGV[9])
return {'ok'}
`)

// acquireScript: драфт жив + слот pending + размер совпал → uploading.
var acquireScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then
  return {'err', 'draft_not_found'}
end
local raw = redis.call('HGET', KEYS[2], ARGV[1])
if not raw then
  return {'err', 'slot_not_found'}
end
local slot = cjson.decode(raw)
if slot.status ~= 'pending' then
  return {'err', 'slot_not_pending'}
end
if tonumber(slot.size) ~= tonumber(ARGV[2]) then
  return {'err', 'size_mismatch'}
end
slot.status = 'uploading'
slot.upload_id = ARGV[1]
redis.call('HSET', KEYS[2], ARGV[1], cjson.encode(slot))
return {'ok', cjson.encode(slot)}
`)

// finishScript: uploading → done (+storage_key). Идемпотентно для done.
var finishScript = redis.NewScript(`
local raw = redis.call('HGET', KEYS[1], ARGV[1])
if not raw then
  return {'err', 'slot_not_found'}
end
local slot = cjson.decode(raw)
if slot.status == 'pending' then
  return {'err', 'slot_not_pending'}
end
slot.status = 'done'
slot.storage_key = ARGV[2]
redis.call('HSET', KEYS[1], ARGV[1], cjson.encode(slot))
return {'ok', cjson.encode(slot)}
`)

// releaseScript: uploading → pending. done/pending/отсутствующий — no-op.
var releaseScript = redis.NewScript(`
local raw = redis.call('HGET', KEYS[1], ARGV[1])
if not raw then
  return {'ok', ''}
end
local slot = cjson.decode(raw)
if slot.status == 'uploading' then
  slot.status = 'pending'
  redis.call('HSET', KEYS[1], ARGV[1], cjson.encode(slot))
end
return {'ok', ''}
`)
