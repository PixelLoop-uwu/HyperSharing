package config

import (
  "fmt"
  "net/url"
  "strings"
  "time"

  "github.com/caarlos0/env/v11"
  "github.com/joho/godotenv"
)

type Config struct {
  Debug bool   `env:"DEBUG"`
  Addr  string `env:"HTTP_ADDR" envDefault:"127.0.0.1:8080"`
  // TusAddr — отдельный порт для tus-загрузок: долгие PATCH не должны делить
  // таймауты с JSON-эндпоинтами (ReadTimeout против slowloris).
  TusAddr      string        `env:"TUS_HTTP_ADDR" envDefault:"127.0.0.1:8081"`
  ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"30s"` // медленное чтение тела (slowloris)
  WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30m"` // медленное чтение ответа при раздаче файлов

  DatabaseURL string `env:"DATABASE_URL,required"`
  RedisURL    string `env:"REDIS_URL" envDefault:"redis://localhost:6379/0"`

  S3 S3 `envPrefix:"S3_"`

  Uploads   Uploads   `envPrefix:"UPLOADS_"`
  Quota     Quota     `envPrefix:"QUOTA_"`
  Posts     Posts     `envPrefix:"POSTS_"`
  Worker    Worker    `envPrefix:"WORKER_"`
  Avatar    Avatar    `envPrefix:"AVATAR_"`
  RateLimit RateLimit `envPrefix:"RATELIMIT_"`

  Session Session `envPrefix:"SESSION_"`
  Argon2  Argon2  `envPrefix:"ARGON2_"`
  // Разрешённые Origin для изменяющих запросов (CSRF-защита), через запятую.
  AllowedOrigins []string `env:"ALLOWED_ORIGINS" envSeparator:","`
}

type S3 struct {
  Endpoint        string `env:"ENDPOINT_URL" envDefault:"http://localhost:9000"`
  PublicEndpoint  string `env:"PUBLIC_ENDPOINT_URL"` // пусто = Endpoint
  AccessKeyID     string `env:"ACCESS_KEY_ID,required"`
  SecretAccessKey string `env:"SECRET_ACCESS_KEY,required"`
  Region          string `env:"REGION" envDefault:"us-east-1"`
  Bucket          string `env:"BUCKET" envDefault:"hypersharing"`
  PathStyle       bool   `env:"PATH_STYLE" envDefault:"true"` // MinIO
}

// Uploads — параметры tus-загрузок в S3 (s3store).
type Uploads struct {
  ObjectPrefix   string `env:"OBJECT_PREFIX" envDefault:"files/"`    // префикс финальных объектов
  MetadataPrefix string `env:"METADATA_PREFIX" envDefault:"tusmeta/"` // префикс служебных .info/.part
}

// Quota — суточные лимиты загрузки. Резерв делает фича post при создании
// драфта; здесь только защитный пере-чек на старте загрузки.
type Quota struct {
  AnonDailyBytes int64 `env:"ANON_DAILY_BYTES" envDefault:"524288000"`  // 500 MiB/сутки на IP
  UserDailyBytes int64 `env:"USER_DAILY_BYTES" envDefault:"5368709120"` // 5 GiB/сутки на пользователя
}

// Posts — лимиты постов и драфтов (аноним vs залогиненный). См. backend/Claude.md.
type Posts struct {
  AnonMaxBytes    int64         `env:"ANON_MAX_BYTES" envDefault:"104857600"`      // 100 MiB
  UserMaxBytes    int64         `env:"USER_MAX_BYTES" envDefault:"734003200"`      // 700 MiB
  AnonMaxFiles    int           `env:"ANON_MAX_FILES" envDefault:"8"`
  UserMaxFiles    int           `env:"USER_MAX_FILES" envDefault:"24"`
  AnonMaxSections int           `env:"ANON_MAX_SECTIONS" envDefault:"16"`
  UserMaxSections int           `env:"USER_MAX_SECTIONS" envDefault:"64"`
  AnonMaxDrafts   int           `env:"ANON_MAX_DRAFTS" envDefault:"2"`
  UserMaxDrafts   int           `env:"USER_MAX_DRAFTS" envDefault:"3"`
  TitleMaxLen     int           `env:"TITLE_MAX_LEN" envDefault:"120"`
  SectionMaxLen   int           `env:"SECTION_MAX_LEN" envDefault:"10000"`
  AnonTTL         time.Duration `env:"ANON_TTL" envDefault:"168h"`                // срок жизни анонимного поста
  DraftTTL        time.Duration `env:"DRAFT_TTL" envDefault:"24h"`                // TTL драфта
  UserStorageQuota int64        `env:"USER_STORAGE_QUOTA" envDefault:"10737418240"` // 10 GiB суммарно
}

// Worker — параметры фоновой очистки (cmd/worker).
type Worker struct {
  Interval           time.Duration `env:"INTERVAL" envDefault:"10m"`   // период запуска джоб
  AbandonedUploadAge time.Duration `env:"ABANDONED_UPLOAD_AGE" envDefault:"24h"` // abort брошенных multipart/мета
  OrphanObjectAge    time.Duration `env:"ORPHAN_OBJECT_AGE" envDefault:"48h"`    // GC сиротских объектов
}

// Avatar — параметры аватаров: перекодирование в JPEG 256×256, лимиты и
// rate limit на загрузку.
type Avatar struct {
  MaxBytes int64         `env:"MAX_BYTES" envDefault:"5242880"` // 5 MiB на вход
  MaxDim   int           `env:"MAX_DIM" envDefault:"2048"`      // защита от бомбы декомпрессии
  Size     int           `env:"SIZE" envDefault:"256"`          // сторона выходного квадрата
  Quality  int           `env:"QUALITY" envDefault:"82"`        // качество JPEG
  Rate     int           `env:"RATE" envDefault:"5"`            // загрузок за окно
  Window   time.Duration `env:"WINDOW" envDefault:"10m"`        // окно rate limit
  Prefix   string        `env:"PREFIX" envDefault:"avatars/"`   // префикс объектов в S3
}

// RateLimit — ограничения частоты на отдельных эндпоинтах (fixed window в Redis).
type RateLimit struct {
  Register     int           `env:"REGISTER" envDefault:"5"`
  Login        int           `env:"LOGIN" envDefault:"10"`
  LoginEmail   int           `env:"LOGIN_EMAIL" envDefault:"10"` // попыток логина на аккаунт (по хешу email)
  Draft        int           `env:"DRAFT" envDefault:"10"`
  FileNotFound int           `env:"FILE_NOT_FOUND" envDefault:"30"`
  Download     int           `env:"DOWNLOAD" envDefault:"100"` // успешных скачиваний /f/ на IP
  Window       time.Duration `env:"WINDOW" envDefault:"10m"`
}

type Session struct {
  TTL          time.Duration `env:"TTL" envDefault:"168h"`  // sliding idle timeout
  MaxAge       time.Duration `env:"MAX_AGE" envDefault:"720h"` // абсолютный максимум жизни
  CookieSecure bool          `env:"COOKIE_SECURE" envDefault:"true"` // true за HTTPS
}

type Argon2 struct {
  Memory      uint32 `env:"MEMORY" envDefault:"65536"` // KiB
  Iterations  uint32 `env:"ITERATIONS" envDefault:"3"`
  Parallelism uint8  `env:"PARALLELISM" envDefault:"4"`
  SaltLen     uint32 `env:"SALT_LEN" envDefault:"16"`
  KeyLen      uint32 `env:"KEY_LEN" envDefault:"32"`
}

func (s S3) PresignEndpoint() string {
  if s.PublicEndpoint != "" {
    return s.PublicEndpoint
  }
  return s.Endpoint
}

// HostPort разбирает Endpoint (полный URL, например "http://localhost:9000")
// в host:port без схемы и флаг TLS — то, что нужно minio-go.
func (s S3) HostPort() (host string, useSSL bool, err error) {
  u, err := url.Parse(s.Endpoint)
  if err != nil {
    return "", false, fmt.Errorf("parse S3_ENDPOINT_URL: %w", err)
  }
  if u.Host == "" {
    return "", false, fmt.Errorf("S3_ENDPOINT_URL has no host: %q", s.Endpoint)
  }
  return u.Host, u.Scheme == "https", nil
}

func Load() (Config, error) {
  _ = godotenv.Load()
  cfg, err := env.ParseAs[Config]()
  if err != nil {
    return cfg, err
  }
  return cfg, cfg.Validate()
}

// Validate проверяет инварианты, от которых зависит безопасность очистки S3.
// Пустой или вложенный префикс заставил бы воркер перечислять/удалять не те
// объекты (вплоть до чужих префиксов или всего бакета).
func (c Config) Validate() error {
  names := []string{"UPLOADS_OBJECT_PREFIX", "UPLOADS_METADATA_PREFIX", "AVATAR_PREFIX"}
  values := []string{c.Uploads.ObjectPrefix, c.Uploads.MetadataPrefix, c.Avatar.Prefix}
  trimmed := make([]string, len(names))
  for i, v := range values {
    trimmed[i] = strings.Trim(v, "/")
    if trimmed[i] == "" {
      return fmt.Errorf("%s must be a non-empty S3 key prefix", names[i])
    }
  }
  for i := 0; i < len(trimmed); i++ {
    for j := i + 1; j < len(trimmed); j++ {
      if prefixesOverlap(trimmed[i], trimmed[j]) {
        return fmt.Errorf("S3 prefixes %s=%q and %s=%q overlap (must not be equal or nested)", names[i], values[i], names[j], values[j])
      }
    }
  }
  if c.Worker.OrphanObjectAge <= c.Posts.DraftTTL {
    return fmt.Errorf("WORKER_ORPHAN_OBJECT_AGE must be greater than POSTS_DRAFT_TTL")
  }
  return nil
}

func prefixesOverlap(a, b string) bool {
  return a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}
