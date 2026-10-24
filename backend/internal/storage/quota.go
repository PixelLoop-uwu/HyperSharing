package storage

import (
  "context"
  "errors"
  "fmt"
  "strings"
  "time"

  "github.com/redis/go-redis/v9"
)

// Квота `quota:{owner}:{yyyymmdd}` живёт чуть больше суток — этого хватает,
// чтобы резерв не «пережил» свой день.
const quotaKeyTTL = 48 * time.Hour

// Quota — суточная квота загрузки. Резерв/возврат делает фича post
// (создание драфта / DELETE /drafts/{id}). Reserve — авторитетный и атомарный;
// повторной проверки на старте загрузки нет (она давала бы двойной учёт).
type Quota struct {
  rdb       *redis.Client
  anonDaily int64
  userDaily int64
}

func NewQuota(rdb *redis.Client, anonDaily, userDaily int64) *Quota {
  return &Quota{rdb: rdb, anonDaily: anonDaily, userDaily: userDaily}
}

// Reserve атомарно увеличивает сегодняшний счётчик владельца на bytes.
// ErrQuotaExceeded, если это превысит дневной лимит.
func (q *Quota) Reserve(ctx context.Context, owner string, bytes int64) error {
  res, err := quotaReserveScript.Run(ctx, q.rdb,
    []string{quotaKey(owner)},
    bytes, q.limitFor(owner), int64(quotaKeyTTL.Seconds()),
  ).Slice()
  if err != nil {
    return fmt.Errorf("quota: reserve: %w", err)
  }
  if len(res) > 0 {
    if code, _ := res[0].(string); code == "err" {
      return ErrQuotaExceeded
    }
  }
  return nil
}

// Release уменьшает счётчик владельца на bytes, не опуская ниже нуля.
// Используется только при явном DELETE /drafts/{id} (брошенный драфт резерв
// не возвращает).
func (q *Quota) Release(ctx context.Context, owner string, bytes int64) error {
  _, err := quotaReleaseScript.Run(ctx, q.rdb,
    []string{quotaKey(owner)},
    bytes,
  ).Result()
  if err != nil {
    return fmt.Errorf("quota: release: %w", err)
  }
  return nil
}

// Get возвращает зарезервированные байты владельца за сегодня (0, если ключа ещё
// нет). Для «условного» отображения квоты в профиле.
func (q *Quota) Get(ctx context.Context, owner string) (int64, error) {
  cur, err := q.rdb.Get(ctx, quotaKey(owner)).Int64()
  if err != nil && !errors.Is(err, redis.Nil) {
    return 0, fmt.Errorf("quota: get: %w", err)
  }
  return cur, nil
}

// Limit возвращает дневной лимит загрузки для владельца.
func (q *Quota) Limit(owner string) int64 {
  return q.limitFor(owner)
}

func (q *Quota) limitFor(owner string) int64 {
  if strings.HasPrefix(owner, "user:") {
    return q.userDaily
  }
  return q.anonDaily
}

func quotaKey(owner string) string {
  return fmt.Sprintf("quota:%s:%s", owner, time.Now().Format("20060102"))
}

// quotaReserveScript: INCRBY + проверка лимита; при превышении откат и ошибка.
var quotaReserveScript = redis.NewScript(`
local cur = redis.call('INCRBY', KEYS[1], ARGV[1])
if cur > tonumber(ARGV[2]) then
  redis.call('DECRBY', KEYS[1], ARGV[1])
  return {'err', 'quota_exceeded'}
end
redis.call('EXPIRE', KEYS[1], ARGV[3])
return {'ok', cur}
`)

// quotaReleaseScript: декремент с полом на нуле, TTL сохраняем.
var quotaReleaseScript = redis.NewScript(`
local cur = redis.call('GET', KEYS[1])
if not cur then
  return {'ok', 0}
end
local n = tonumber(cur) - tonumber(ARGV[1])
if n < 0 then n = 0 end
redis.call('SET', KEYS[1], n, 'KEEPTTL')
return {'ok', n}
`)
