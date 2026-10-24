package httpx

import (
  "context"
  "crypto/rand"
  "log/slog"
  "net/http"
)

const RequestIDHeader = "X-Request-ID"

// Unexported type: no other package can produce an equal key,
// so values never collide with someone else's context keys.
type ctxKey struct{}

func RequestID(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    id := r.Header.Get(RequestIDHeader)
    if !validRequestID(id) {
      id = rand.Text()
    }
    w.Header().Set(RequestIDHeader, id)
    ctx := context.WithValue(r.Context(), ctxKey{}, id)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}

func RequestIDFrom(ctx context.Context) string {
  id, _ := ctx.Value(ctxKey{}).(string)
  return id
}

// Logger returns the default logger with the request id attached, if any.
// A nil ctx is allowed.
func Logger(ctx context.Context) *slog.Logger {
  if ctx == nil {
    return slog.Default()
  }
  if id := RequestIDFrom(ctx); id != "" {
    return slog.Default().With("request_id", id)
  }
  return slog.Default()
}

// validRequestID accepts only short, plain ids. The value ends up in logs,
// so arbitrary client input would allow log injection.
func validRequestID(id string) bool {
  if id == "" || len(id) > 64 {
    return false
  }
  for _, c := range id {
    switch {
    case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '_':
    default:
      return false
    }
  }
  return true
}