package storage

import (
  "net/http"
  "time"

  "github.com/jackc/pgx/v5/pgxpool"
  "github.com/redis/go-redis/v9"

  "hypersharing/internal/draft"
  "hypersharing/internal/platform/ratelimit"
  "hypersharing/internal/platform/s3x"
)

// Config — зависимости storage, собранные из env-конфига в app.go.
type Config struct {
  S3EndpointURL string
  S3AccessKeyID string
  S3SecretAccessKey string
  S3Region       string
  S3PathStyle    bool
  Bucket         string

  ObjectPrefix   string
  MetadataPrefix string

  QuotaAnonDailyBytes int64
  QuotaUserDailyBytes int64

  FileNotFoundLimit int
  DownloadLimit     int
  RateLimitWindow   time.Duration
  AnonTTL           time.Duration

  TusBasePath string // внешний путь tus, например "/api/v1/uploads/"
}

// Service собирает tus-загрузки, квоту и выдачу файлов. Слоты драфта берёт из
// общего draft.Store (владелец схемы — пакет draft).
type Service struct {
  drafts  *draft.Store
  quota   *Quota
  files   *FileRepository
  bucket  *s3x.Bucket
  limiter *ratelimit.Limiter
  rl      Config
  tus     http.Handler
}

func New(cfg Config, rdb *redis.Client, pool *pgxpool.Pool, bucket *s3x.Bucket, drafts *draft.Store, limiter *ratelimit.Limiter) (*Service, error) {
  quota := NewQuota(rdb, cfg.QuotaAnonDailyBytes, cfg.QuotaUserDailyBytes)
  files := NewFileRepository(pool)

  tus, err := newTusHandler(cfg, drafts)
  if err != nil {
    return nil, err
  }
  return &Service{drafts: drafts, quota: quota, files: files, bucket: bucket, limiter: limiter, rl: cfg, tus: tus}, nil
}

// Tus возвращает handler tus-загрузок (монтируется через http.StripPrefix).
func (s *Service) Tus() http.Handler {
  return s.tus
}

// Quota возвращает хранилище квоты — его резервирует фича post при создании
// драфта.
func (s *Service) Quota() *Quota {
  return s.quota
}

// ServeFile — обработчик GET /f/{shortKey}.
func (s *Service) ServeFile() http.HandlerFunc {
  return NewServeHandler(s.files, s.bucket, s.limiter, s.rl.FileNotFoundLimit, s.rl.DownloadLimit, s.rl.RateLimitWindow, s.rl.AnonTTL).ServeFile
}
