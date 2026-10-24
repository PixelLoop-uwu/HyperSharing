package storage

import (
  "errors"
  "mime"
  "net/http"
  "time"

  "github.com/go-chi/chi/v5"

  "hypersharing/internal/draft"
  "hypersharing/internal/httpx"
  "hypersharing/internal/platform/ratelimit"
  "hypersharing/internal/platform/s3x"
)

// ServeHandler выдаёт файлы по short_key: `GET /f/{shortKey}`. 404-ответы
// лимитируются по IP — против перебора коротких ключей.
type ServeHandler struct {
  files         *FileRepository
  bucket        *s3x.Bucket
  limiter       *ratelimit.Limiter
  notFoundLimit int
  downloadLimit int
  window        time.Duration
  anonTTL       time.Duration
}

func NewServeHandler(files *FileRepository, bucket *s3x.Bucket, limiter *ratelimit.Limiter, notFoundLimit, downloadLimit int, window, anonTTL time.Duration) *ServeHandler {
  return &ServeHandler{files: files, bucket: bucket, limiter: limiter, notFoundLimit: notFoundLimit, downloadLimit: downloadLimit, window: window, anonTTL: anonTTL}
}

func (h *ServeHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
  f, err := h.files.GetByShortKey(r.Context(), chi.URLParam(r, "shortKey"))
  if err != nil {
    if errors.Is(err, ErrFileNotFound) {
      h.notFound(w, r)
      return
    }
    httpx.Internal(w, r, err)
    return
  }

  // Просроченный анонимный пост: файл не отдаём, даже до прогона воркера.
  if f.AuthorID == nil && time.Since(f.PostCreatedAt) > h.anonTTL {
    h.notFound(w, r)
    return
  }

  // Успешные скачивания тоже лимитируются по IP — против расхода исходящего
  // трафика через свой же файл.
  ok, err := h.limiter.Allow(r.Context(), "rl:download:f:"+draft.IPFromAddr(r.RemoteAddr), h.downloadLimit, h.window)
  if err != nil {
    httpx.Internal(w, r, err)
    return
  }
  if !ok {
    httpx.Error(w, http.StatusTooManyRequests, "rate_limited", "too many downloads")
    return
  }

  obj, info, err := h.bucket.Get(r.Context(), f.StorageKey)
  if err != nil {
    if errors.Is(err, s3x.ErrNotFound) {
      h.notFound(w, r)
      return
    }
    httpx.Internal(w, r, err)
    return
  }
  defer obj.Close()

  contentType := f.MimeType
  if contentType == "" {
    contentType = "application/octet-stream"
  }
  // attachment + nosniff: файл никогда не исполняется в контексте сайта.
  w.Header().Set("Content-Type", contentType)
  w.Header().Set("X-Content-Type-Options", "nosniff")
  w.Header().Set("Content-Security-Policy", "sandbox")
  w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": f.Name}))

  // obj умеет Seek → Range, If-Modified-Since и докачка работают из коробки.
  http.ServeContent(w, r, f.Name, info.LastModified, obj)
}

// notFound отвечает 404, но при частых промахах одного IP — 429.
func (h *ServeHandler) notFound(w http.ResponseWriter, r *http.Request) {
  ok, err := h.limiter.Allow(r.Context(), "rl:404:f:"+draft.IPFromAddr(r.RemoteAddr), h.notFoundLimit, h.window)
  if err != nil {
    httpx.Internal(w, r, err)
    return
  }
  if !ok {
    httpx.Error(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
    return
  }
  httpx.Error(w, http.StatusNotFound, "not_found", "file not found")
}
