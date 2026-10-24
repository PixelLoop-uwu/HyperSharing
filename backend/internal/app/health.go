package app

import (
  "context"
  "log/slog"
  "net/http"
  "time"
)

type check func(ctx context.Context) error

func health(checks map[string]check) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()

    for name, check := range checks {
      if err := check(ctx); err != nil {
        slog.Error("health check failed", "check", name, "err", err)
        http.Error(w, name+" unavailable", http.StatusServiceUnavailable)
        return
      }
    }
    w.WriteHeader(http.StatusOK)
  }
}