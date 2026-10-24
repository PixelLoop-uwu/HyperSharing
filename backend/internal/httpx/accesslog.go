package httpx

import (
  "log/slog"
  "net/http"
  "time"
)

// statusWriter remembers what the handler wrote. http.ResponseWriter does not
// expose the status code, so we wrap it.
type statusWriter struct {
  http.ResponseWriter
  status int
  bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
  if w.status == 0 {
    w.status = code
  }
  w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
  if w.status == 0 {
    w.status = http.StatusOK
  }
  n, err := w.ResponseWriter.Write(b)
  w.bytes += n
  return n, err
}

// Unwrap lets http.ResponseController reach the original writer
// (Flush, SetWriteDeadline and so on).
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func AccessLog(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    sw := &statusWriter{ResponseWriter: w}

    next.ServeHTTP(sw, r)

    status := sw.status
    if status == 0 {
      status = http.StatusOK // handler wrote nothing: net/http sends 200
    }
    level := slog.LevelInfo
    if status >= 500 {
      level = slog.LevelError
    }
    Logger(r.Context()).Log(r.Context(), level, "request",
      "method", r.Method,
      "path", r.URL.Path,
      "status", status,
      "bytes", sw.bytes,
      "duration", time.Since(start),
    )
  })
}