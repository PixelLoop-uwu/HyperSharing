package user

import (
  "bytes"
  "context"
  "crypto/sha256"
  "encoding/hex"
  "errors"
  "fmt"
  "image"
  "image/color"
  _ "image/png" // регистрирует декодер png
  "image/jpeg"
  _ "golang.org/x/image/webp" // регистрирует декодер webp (в т.ч. lossless)
  "golang.org/x/image/draw"
  "io"
  "net/http"
  "strings"
  "time"

  "github.com/go-chi/chi/v5"
  "github.com/google/uuid"

  "hypersharing/internal/httpx"
  "hypersharing/internal/platform/ratelimit"
  "hypersharing/internal/platform/s3x"
)

const avatarCacheControl = "public, max-age=31536000, immutable"

var (
  ErrUnsupportedImage = errors.New("user: unsupported image type")
  ErrInvalidImage     = errors.New("user: invalid image")
)

// AvatarConfig — параметры обработки аватара, собираются из env в app.go.
type AvatarConfig struct {
  MaxBytes int64
  MaxDim   int
  Size     int
  Quality  int
  Rate     int
  Window   time.Duration
  Prefix   string
}

// AvatarService — загрузка/удаление/раздача аватаров.
type AvatarService struct {
  repo    *Repository
  bucket  *s3x.Bucket
  limiter *ratelimit.Limiter
  cfg     AvatarConfig
}

func NewAvatarService(repo *Repository, bucket *s3x.Bucket, limiter *ratelimit.Limiter, cfg AvatarConfig) *AvatarService {
  return &AvatarService{repo: repo, bucket: bucket, limiter: limiter, cfg: cfg}
}

// AvatarKey — ключ объекта аватара в S3: prefix/<id>/<version>. Версия в ключе
// делает объект неизменяемым: параллельные загрузки пишут в разные ключи и не
// гоняются между собой, а старые версии подчищает воркер.
func AvatarKey(prefix string, id uuid.UUID, version string) string {
  p := strings.TrimRight(prefix, "/")
  if p == "" {
    return id.String() + "/" + version
  }
  return p + "/" + id.String() + "/" + version
}

// Upload перекодирует картинку в JPEG 256×256, кладёт в S3 и фиксирует версию
// в БД. Перезаписывает старый аватар (S3-ключ тот же, версия меняется → URL
// меняется).
func (a *AvatarService) Upload(ctx context.Context, userID uuid.UUID, data []byte) error {
  processed, err := processAvatar(data, a.cfg.Size, a.cfg.Quality, a.cfg.MaxDim)
  if err != nil {
    return err
  }
  version := versionHash(processed)
  key := AvatarKey(a.cfg.Prefix, userID, version)
  if _, err := a.bucket.Put(ctx, key, bytes.NewReader(processed), int64(len(processed)), "image/jpeg"); err != nil {
    return err
  }
  return a.repo.SetAvatarVersion(ctx, userID, version)
}

// Delete снимает аватар: версия → NULL. Объект в S3 подчистит воркер.
func (a *AvatarService) Delete(ctx context.Context, userID uuid.UUID) error {
  return a.repo.ClearAvatarVersion(ctx, userID)
}

func (a *AvatarService) UploadHandler(w http.ResponseWriter, r *http.Request) {
  u := UserFrom(r.Context())
  if u == nil { // RequireAuth гарантирует не-nil
    httpx.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
    return
  }

  ok, err := a.limiter.Allow(r.Context(), "avatar:"+u.ID.String(), a.cfg.Rate, a.cfg.Window)
  if err != nil {
    httpx.Internal(w, r, err)
    return
  }
  if !ok {
    httpx.Error(w, http.StatusTooManyRequests, "rate_limited", "too many avatar uploads")
    return
  }

  data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, a.cfg.MaxBytes))
  if err != nil {
    var tooBig *http.MaxBytesError
    if errors.As(err, &tooBig) {
      httpx.Error(w, http.StatusRequestEntityTooLarge, "too_large", "avatar too large")
    } else {
      httpx.Error(w, http.StatusBadRequest, "bad_request", "could not read body")
    }
    return
  }

  switch err := a.Upload(r.Context(), u.ID, data); {
  case err == nil:
    w.WriteHeader(http.StatusNoContent)
  case errors.Is(err, ErrUnsupportedImage):
    httpx.Error(w, http.StatusUnsupportedMediaType, "unsupported_image", "only png, jpeg or webp allowed")
  case errors.Is(err, ErrInvalidImage):
    httpx.Error(w, http.StatusBadRequest, "invalid_image", "invalid or corrupt image")
  default:
    httpx.Internal(w, r, err)
  }
}

func (a *AvatarService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
  u := UserFrom(r.Context())
  if u == nil {
    httpx.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
    return
  }
  if err := a.Delete(r.Context(), u.ID); err != nil {
    httpx.Internal(w, r, err)
    return
  }
  w.WriteHeader(http.StatusNoContent)
}

// Serve — публичный GET /avatars/{userID}. Отдаёт аватар с долгим immutable
// кешем; смена аватара меняет URL (version в query), поэтому кеш безопасен.
func (a *AvatarService) Serve(w http.ResponseWriter, r *http.Request) {
  userID, err := uuid.Parse(chi.URLParam(r, "userID"))
  if err != nil {
    httpx.Error(w, http.StatusNotFound, "not_found", "not found")
    return
  }
  version, err := a.repo.GetAvatarVersion(r.Context(), userID)
  if err != nil {
    if errors.Is(err, ErrNotFound) {
      httpx.Error(w, http.StatusNotFound, "not_found", "not found")
      return
    }
    httpx.Internal(w, r, err)
    return
  }

  obj, info, err := a.bucket.Get(r.Context(), AvatarKey(a.cfg.Prefix, userID, version))
  if err != nil {
    if errors.Is(err, s3x.ErrNotFound) {
      httpx.Error(w, http.StatusNotFound, "not_found", "not found")
      return
    }
    httpx.Internal(w, r, err)
    return
  }
  defer obj.Close()

  w.Header().Set("Content-Type", "image/jpeg")
  w.Header().Set("Cache-Control", avatarCacheControl)
  http.ServeContent(w, r, "", info.LastModified, obj)
}

// processAvatar проверяет тип через DetectContentType (заголовку клиента не
// верим), декодирует, режет по центру до квадрата и ресайзит в size×size,
// кодирует в JPEG. Возвращает ErrUnsupportedImage / ErrInvalidImage.
func processAvatar(data []byte, size, quality, maxDim int) ([]byte, error) {
  switch http.DetectContentType(data) {
  case "image/png", "image/jpeg", "image/webp":
  default:
    return nil, fmt.Errorf("%w", ErrUnsupportedImage)
  }

  // Защита от бомбы декомпрессии: размеры читаем из заголовка, не декодируя.
  cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
  if err != nil {
    return nil, fmt.Errorf("%w: decode config: %v", ErrInvalidImage, err)
  }
  if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > maxDim || cfg.Height > maxDim {
    return nil, fmt.Errorf("%w: dimensions %dx%d", ErrInvalidImage, cfg.Width, cfg.Height)
  }

  src, _, err := image.Decode(bytes.NewReader(data))
  if err != nil {
    return nil, fmt.Errorf("%w: decode: %v", ErrInvalidImage, err)
  }

  srcRect := centeredSquare(src.Bounds())
  dst := image.NewRGBA(image.Rect(0, 0, size, size))
  // Заливаем белым: иначе прозрачные пиксели PNG/WebP дадут чёрный фон в JPEG.
  draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
  draw.CatmullRom.Scale(dst, dst.Bounds(), src, srcRect, draw.Over, nil)

  var buf bytes.Buffer
  if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: quality}); err != nil {
    return nil, fmt.Errorf("%w: encode: %v", ErrInvalidImage, err)
  }
  return buf.Bytes(), nil
}

// centeredSquare — вписанный по центру квадрат (кавер-кроп без искажений).
func centeredSquare(b image.Rectangle) image.Rectangle {
  side := min(b.Dx(), b.Dy())
  x0 := b.Min.X + (b.Dx()-side)/2
  y0 := b.Min.Y + (b.Dy()-side)/2
  return image.Rect(x0, y0, x0+side, y0+side)
}

func versionHash(jpeg []byte) string {
  sum := sha256.Sum256(jpeg)
  return hex.EncodeToString(sum[:8])
}
